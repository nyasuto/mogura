package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"mogura/internal"
)

func Scan(root string) ([]internal.FileInfo, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	if _, err := os.Lstat(root); err != nil {
		return nil, err
	}

	var files []internal.FileInfo

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s: %v\n", path, err)
			return nil
		}

		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s: %v\n", path, err)
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))

		files = append(files, internal.FileInfo{
			Path:    path,
			Size:    info.Size(),
			Dir:     filepath.Dir(path),
			Ext:     ext,
			ModTime: info.ModTime(),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
