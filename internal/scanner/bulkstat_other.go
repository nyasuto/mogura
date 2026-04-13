//go:build !darwin

package scanner

import (
	"os"
	"path/filepath"
	"syscall"
	"time"
)

type bulkEntry struct {
	Name         string
	IsDir        bool
	IsSymlink    bool
	Size         int64
	PhysicalSize int64
	ModTime      time.Time
}

func readDirBulk(dirPath string) ([]bulkEntry, error) {
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

func readDirFallback(dirPath string) ([]bulkEntry, error) {
	return readDirBulk(dirPath)
}
