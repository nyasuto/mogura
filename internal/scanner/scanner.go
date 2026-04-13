package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"mogura/internal"
)

type ScanOpts struct {
	Exclude       []string
	OnProgress    func(scanned int, currentDir string)
	OneFileSystem bool
}

func deviceID(path string) (uint64, error) {
	var stat syscall.Stat_t
	if err := syscall.Lstat(path, &stat); err != nil {
		return 0, err
	}
	return uint64(stat.Dev), nil
}

func isGlobPattern(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[")
}

func matchesExclude(name string, exactSet map[string]bool, globs []string) bool {
	if exactSet[name] {
		return true
	}
	for _, g := range globs {
		if matched, _ := filepath.Match(g, name); matched {
			return true
		}
	}
	return false
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

	exactSet := make(map[string]bool)
	var globs []string
	for _, e := range opt.Exclude {
		if isGlobPattern(e) {
			globs = append(globs, e)
		} else {
			exactSet[e] = true
		}
	}

	var rootDev uint64
	if opt.OneFileSystem {
		rootDev, err = deviceID(root)
		if err != nil {
			return nil, fmt.Errorf("cannot stat root: %w", err)
		}
	}

	var files []internal.FileInfo
	scanned := 0

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s: %v\n", path, err)
			return nil
		}

		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		if d.IsDir() {
			if path != root && matchesExclude(d.Name(), exactSet, globs) {
				return filepath.SkipDir
			}
			if opt.OneFileSystem && path != root {
				dev, devErr := deviceID(path)
				if devErr != nil {
					fmt.Fprintf(os.Stderr, "warning: %s: %v\n", path, devErr)
					return filepath.SkipDir
				}
				if dev != rootDev {
					return filepath.SkipDir
				}
			}
			return nil
		}

		if matchesExclude(d.Name(), exactSet, globs) {
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
		scanned++

		if opt.OnProgress != nil {
			opt.OnProgress(scanned, filepath.Dir(path))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
