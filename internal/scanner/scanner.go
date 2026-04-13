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
	Exclude    []string
	OnProgress func(scanned int, currentDir string)
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
