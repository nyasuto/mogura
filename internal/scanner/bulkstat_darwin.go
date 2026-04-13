//go:build darwin

package scanner

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

type bulkEntry struct {
	Name         string
	IsDir        bool
	IsSymlink    bool
	Size         int64
	PhysicalSize int64
	ModTime      time.Time
}

const (
	_ATTR_BIT_MAP_COUNT = 5

	_ATTR_CMN_RETURNED_ATTRS = 0x80000000
	_ATTR_CMN_NAME           = 0x00000001
	_ATTR_CMN_OBJTYPE        = 0x00000008
	_ATTR_CMN_MODTIME        = 0x00000400

	_ATTR_FILE_TOTALSIZE = 0x00000002
	_ATTR_FILE_ALLOCSIZE = 0x00000004

	_VREG = 1
	_VDIR = 2
	_VLNK = 5

	_bulkBufSize = 256 * 1024

	_commonAttrMask = _ATTR_CMN_RETURNED_ATTRS | _ATTR_CMN_NAME | _ATTR_CMN_OBJTYPE | _ATTR_CMN_MODTIME
	_fileAttrMask   = _ATTR_FILE_TOTALSIZE | _ATTR_FILE_ALLOCSIZE
)

type darwinAttrList struct {
	bitmapCount uint16
	reserved    uint16
	commonAttr  uint32
	volAttr     uint32
	dirAttr     uint32
	fileAttr    uint32
	forkAttr    uint32
}

func callGetattrlistbulk(fd int, list *darwinAttrList, buf []byte) (int, error) {
	r1, _, e1 := unix.Syscall6(
		unix.SYS_GETATTRLISTBULK,
		uintptr(fd),
		uintptr(unsafe.Pointer(list)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		0, 0,
	)
	if e1 != 0 {
		return 0, e1
	}
	return int(r1), nil
}

func newBulkAttrList() darwinAttrList {
	return darwinAttrList{
		bitmapCount: _ATTR_BIT_MAP_COUNT,
		commonAttr:  _commonAttrMask,
		fileAttr:    _fileAttrMask,
	}
}

func readDirBulk(path string) ([]bulkEntry, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_DIRECTORY, 0)
	if err != nil {
		return nil, err
	}
	defer unix.Close(fd) //nolint:errcheck

	attrList := newBulkAttrList()
	buf := make([]byte, _bulkBufSize)
	var entries []bulkEntry

	for {
		n, err := callGetattrlistbulk(fd, &attrList, buf)
		if err != nil {
			return entries, err
		}
		if n == 0 {
			break
		}

		offset := 0
		for i := 0; i < n; i++ {
			if offset+4 > len(buf) {
				break
			}
			recLen := int(binary.LittleEndian.Uint32(buf[offset : offset+4]))
			if recLen < 4 || offset+recLen > len(buf) {
				break
			}
			entry, ok := parseRecord(buf[offset : offset+recLen])
			if ok {
				entries = append(entries, entry)
			}
			offset += recLen
		}
	}

	return entries, nil
}

func parseRecord(rec []byte) (bulkEntry, bool) {
	// record layout (fixed part):
	//   [0:4]   uint32  record length
	//   [4:8]   uint32  returned commonattr
	//   [8:12]  uint32  returned volattr
	//   [12:16] uint32  returned dirattr
	//   [16:20] uint32  returned fileattr
	//   [20:24] uint32  returned forkattr
	//   [24:28] int32   name attrreference offset
	//   [28:32] uint32  name attrreference length
	//   [32:36] uint32  objtype (fsobj_type_t)
	//   [36:44] int64   modtime tv_sec
	//   [44:52] int64   modtime tv_nsec
	//   [52:60] int64   totalsize  (files only, if returned)
	//   [60:68] int64   allocsize  (files only, if returned)
	//   [variable] name string data
	const minRecordSize = 52
	if len(rec) < minRecordSize {
		return bulkEntry{}, false
	}

	retFile := binary.LittleEndian.Uint32(rec[16:20])

	nameRefOff := int(int32(binary.LittleEndian.Uint32(rec[24:28])))
	nameLen := int(binary.LittleEndian.Uint32(rec[28:32]))
	nameStart := 24 + nameRefOff
	if nameStart < 0 || nameLen == 0 || nameStart+nameLen > len(rec) {
		return bulkEntry{}, false
	}
	name := rec[nameStart : nameStart+nameLen]
	for len(name) > 0 && name[len(name)-1] == 0 {
		name = name[:len(name)-1]
	}

	objType := binary.LittleEndian.Uint32(rec[32:36])
	sec := int64(binary.LittleEndian.Uint64(rec[36:44]))
	nsec := int64(binary.LittleEndian.Uint64(rec[44:52]))

	entry := bulkEntry{
		Name:      string(name),
		IsDir:     objType == _VDIR,
		IsSymlink: objType == _VLNK,
		ModTime:   time.Unix(sec, nsec),
	}

	if objType == _VREG {
		pos := 52
		if retFile&_ATTR_FILE_TOTALSIZE != 0 && pos+8 <= len(rec) {
			entry.Size = int64(binary.LittleEndian.Uint64(rec[pos : pos+8]))
			pos += 8
		}
		if retFile&_ATTR_FILE_ALLOCSIZE != 0 && pos+8 <= len(rec) {
			entry.PhysicalSize = int64(binary.LittleEndian.Uint64(rec[pos : pos+8]))
		}
	}

	return entry, true
}

func readDirFallback(dirPath string) ([]bulkEntry, error) {
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	entries := make([]bulkEntry, 0, len(dirEntries))
	for _, de := range dirEntries {
		typ := de.Type()
		isSymlink := typ&os.ModeSymlink != 0

		entry := bulkEntry{
			Name:      de.Name(),
			IsDir:     de.IsDir(),
			IsSymlink: isSymlink,
		}

		if isSymlink {
			entries = append(entries, entry)
			continue
		}

		fullPath := filepath.Join(dirPath, de.Name())
		info, infoErr := os.Lstat(fullPath)
		if infoErr != nil {
			continue
		}
		entry.ModTime = info.ModTime()

		if !de.IsDir() {
			entry.Size = info.Size()
			entry.PhysicalSize = info.Size()
			if stat, ok := info.Sys().(*syscall.Stat_t); ok {
				entry.PhysicalSize = stat.Blocks * 512
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
