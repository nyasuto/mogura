//go:build darwin

package scanner

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseRecord(t *testing.T) {
	tests := []struct {
		name    string
		rec     func() []byte
		wantOK  bool
		wantEnt bulkEntry
	}{
		{
			name: "regular file with size",
			rec: func() []byte {
				buf := make([]byte, 80)
				nameStr := "hello.txt"
				nameOffset := 52 + 16 - 24 // name data starts at 68, relative to offset 24
				binary.LittleEndian.PutUint32(buf[0:4], 80)
				binary.LittleEndian.PutUint32(buf[4:8], _commonAttrMask)
				binary.LittleEndian.PutUint32(buf[16:20], _fileAttrMask)
				binary.LittleEndian.PutUint32(buf[24:28], uint32(nameOffset))
				binary.LittleEndian.PutUint32(buf[28:32], uint32(len(nameStr)+1))
				binary.LittleEndian.PutUint32(buf[32:36], _VREG)
				binary.LittleEndian.PutUint64(buf[36:44], uint64(1700000000))
				binary.LittleEndian.PutUint64(buf[44:52], 0)
				binary.LittleEndian.PutUint64(buf[52:60], 4096)
				binary.LittleEndian.PutUint64(buf[60:68], 8192)
				copy(buf[68:], nameStr)
				return buf
			},
			wantOK: true,
			wantEnt: bulkEntry{
				Name:         "hello.txt",
				IsDir:        false,
				IsSymlink:    false,
				Size:         4096,
				PhysicalSize: 8192,
				ModTime:      time.Unix(1700000000, 0),
			},
		},
		{
			name: "directory entry",
			rec: func() []byte {
				buf := make([]byte, 72)
				nameStr := "subdir"
				nameOffset := 52 - 24
				binary.LittleEndian.PutUint32(buf[0:4], 72)
				binary.LittleEndian.PutUint32(buf[4:8], _commonAttrMask)
				binary.LittleEndian.PutUint32(buf[16:20], 0)
				binary.LittleEndian.PutUint32(buf[24:28], uint32(nameOffset))
				binary.LittleEndian.PutUint32(buf[28:32], uint32(len(nameStr)+1))
				binary.LittleEndian.PutUint32(buf[32:36], _VDIR)
				binary.LittleEndian.PutUint64(buf[36:44], uint64(1600000000))
				binary.LittleEndian.PutUint64(buf[44:52], 0)
				copy(buf[52:], nameStr)
				return buf
			},
			wantOK: true,
			wantEnt: bulkEntry{
				Name:    "subdir",
				IsDir:   true,
				ModTime: time.Unix(1600000000, 0),
			},
		},
		{
			name: "symlink entry",
			rec: func() []byte {
				buf := make([]byte, 72)
				nameStr := "link"
				nameOffset := 52 - 24
				binary.LittleEndian.PutUint32(buf[0:4], 72)
				binary.LittleEndian.PutUint32(buf[4:8], _commonAttrMask)
				binary.LittleEndian.PutUint32(buf[16:20], 0)
				binary.LittleEndian.PutUint32(buf[24:28], uint32(nameOffset))
				binary.LittleEndian.PutUint32(buf[28:32], uint32(len(nameStr)+1))
				binary.LittleEndian.PutUint32(buf[32:36], _VLNK)
				binary.LittleEndian.PutUint64(buf[36:44], uint64(1500000000))
				binary.LittleEndian.PutUint64(buf[44:52], 0)
				copy(buf[52:], nameStr)
				return buf
			},
			wantOK: true,
			wantEnt: bulkEntry{
				Name:      "link",
				IsSymlink: true,
				ModTime:   time.Unix(1500000000, 0),
			},
		},
		{
			name: "record too short",
			rec: func() []byte {
				return make([]byte, 10)
			},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, ok := parseRecord(tt.rec())
			if ok != tt.wantOK {
				t.Fatalf("parseRecord ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if entry.Name != tt.wantEnt.Name {
				t.Errorf("Name = %q, want %q", entry.Name, tt.wantEnt.Name)
			}
			if entry.IsDir != tt.wantEnt.IsDir {
				t.Errorf("IsDir = %v, want %v", entry.IsDir, tt.wantEnt.IsDir)
			}
			if entry.IsSymlink != tt.wantEnt.IsSymlink {
				t.Errorf("IsSymlink = %v, want %v", entry.IsSymlink, tt.wantEnt.IsSymlink)
			}
			if entry.Size != tt.wantEnt.Size {
				t.Errorf("Size = %d, want %d", entry.Size, tt.wantEnt.Size)
			}
			if entry.PhysicalSize != tt.wantEnt.PhysicalSize {
				t.Errorf("PhysicalSize = %d, want %d", entry.PhysicalSize, tt.wantEnt.PhysicalSize)
			}
			if !entry.ModTime.Equal(tt.wantEnt.ModTime) {
				t.Errorf("ModTime = %v, want %v", entry.ModTime, tt.wantEnt.ModTime)
			}
		})
	}
}

func TestReadDirBulk(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file2.dat"), make([]byte, 8192), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(tmpDir, "file1.txt"), filepath.Join(tmpDir, "link1")); err != nil {
		t.Fatal(err)
	}

	entries, err := readDirBulk(tmpDir)
	if err != nil {
		t.Fatalf("readDirBulk error: %v", err)
	}

	byName := make(map[string]bulkEntry)
	for _, e := range entries {
		byName[e.Name] = e
	}

	if len(byName) < 4 {
		t.Fatalf("expected at least 4 entries, got %d: %v", len(entries), entries)
	}

	f1, ok := byName["file1.txt"]
	if !ok {
		t.Fatal("missing file1.txt")
	}
	if f1.IsDir || f1.IsSymlink {
		t.Errorf("file1.txt: wrong type flags")
	}
	if f1.Size != 5 {
		t.Errorf("file1.txt: Size = %d, want 5", f1.Size)
	}

	f2, ok := byName["file2.dat"]
	if !ok {
		t.Fatal("missing file2.dat")
	}
	if f2.Size != 8192 {
		t.Errorf("file2.dat: Size = %d, want 8192", f2.Size)
	}
	if f2.PhysicalSize <= 0 {
		t.Errorf("file2.dat: PhysicalSize = %d, expected > 0", f2.PhysicalSize)
	}

	sub, ok := byName["subdir"]
	if !ok {
		t.Fatal("missing subdir")
	}
	if !sub.IsDir {
		t.Errorf("subdir: IsDir = false, want true")
	}

	lnk, ok := byName["link1"]
	if !ok {
		t.Fatal("missing link1")
	}
	if !lnk.IsSymlink {
		t.Errorf("link1: IsSymlink = false, want true")
	}
}

func TestReadDirBulkConsistentWithLstat(t *testing.T) {
	tmpDir := t.TempDir()

	for i := 0; i < 50; i++ {
		name := filepath.Join(tmpDir, "f"+string(rune('A'+i%26))+string(rune('0'+i/26)))
		if err := os.WriteFile(name, make([]byte, i*100), 0644); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := readDirBulk(tmpDir)
	if err != nil {
		t.Fatalf("readDirBulk error: %v", err)
	}

	bulkMap := make(map[string]bulkEntry)
	for _, e := range entries {
		bulkMap[e.Name] = e
	}

	osEntries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, de := range osEntries {
		info, err := de.Info()
		if err != nil {
			t.Fatal(err)
		}
		be, ok := bulkMap[de.Name()]
		if !ok {
			t.Errorf("readDirBulk missing entry: %s", de.Name())
			continue
		}
		if be.Size != info.Size() {
			t.Errorf("%s: bulk Size=%d, lstat Size=%d", de.Name(), be.Size, info.Size())
		}
		if be.IsDir != de.IsDir() {
			t.Errorf("%s: bulk IsDir=%v, lstat IsDir=%v", de.Name(), be.IsDir, de.IsDir())
		}
	}
}

func TestReadDirBulkEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	entries, err := readDirBulk(tmpDir)
	if err != nil {
		t.Fatalf("readDirBulk error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for empty dir, got %d", len(entries))
	}
}

func TestReadDirBulkNonExistent(t *testing.T) {
	_, err := readDirBulk("/nonexistent_path_for_testing_12345")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}
