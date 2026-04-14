package scanner

import (
	"fmt"
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
	Workers       int
	UseBulkStat   *bool
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

	var rootDev uint64
	if opt.OneFileSystem {
		rootDev, err = deviceID(root)
		if err != nil {
			return nil, fmt.Errorf("cannot stat root: %w", err)
		}
	}

	ps := newParallelScanner(root, opt)
	ps.rootDev = rootDev
	ps.enqueue(root)
	ps.start()

	files := ps.collect()
	return files, nil
}
