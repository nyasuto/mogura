package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"mogura/internal"
)

type ScanOpts struct {
	Exclude []string
}

func Scan(root string, opts ...ScanOpts) ([]internal.FileInfo, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	if _, err := os.Lstat(root); err != nil {
		return nil, err
	}

	var opt ScanOpts
	if len(opts) > 0 {
		opt = opts[0]
	}

	excludeSet := make(map[string]bool, len(opt.Exclude))
	for _, e := range opt.Exclude {
		excludeSet[e] = true
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
			if path != root && excludeSet[d.Name()] {
				return filepath.SkipDir
			}
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
