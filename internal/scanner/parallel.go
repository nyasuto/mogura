package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"mogura/internal"
)

type dirTask struct {
	path string
}

type parallelScanner struct {
	opts     ScanOpts
	workers  int
	rootPath string
	rootDev  uint64

	exactSet map[string]bool
	globs    []string

	pendingCh chan dirTask
	taskCh    chan dirTask
	resultCh  chan internal.FileInfo
	wg        sync.WaitGroup
}

func resolveWorkers(n int) int {
	if n > 0 {
		return n
	}
	return runtime.NumCPU()
}

func newParallelScanner(root string, opts ScanOpts) *parallelScanner {
	exactSet := make(map[string]bool)
	var globs []string
	for _, e := range opts.Exclude {
		if isGlobPattern(e) {
			globs = append(globs, e)
		} else {
			exactSet[e] = true
		}
	}

	workers := resolveWorkers(opts.Workers)

	return &parallelScanner{
		opts:      opts,
		workers:   workers,
		rootPath:  root,
		exactSet:  exactSet,
		globs:     globs,
		pendingCh: make(chan dirTask, workers*4),
		taskCh:    make(chan dirTask, workers),
		resultCh:  make(chan internal.FileInfo, workers*64),
	}
}

func (ps *parallelScanner) dispatcher() {
	var queue []dirTask
	pendingCh := ps.pendingCh

	for {
		if len(queue) == 0 {
			task, ok := <-pendingCh
			if !ok {
				break
			}
			queue = append(queue, task)
			continue
		}

		if pendingCh == nil {
			ps.taskCh <- queue[0]
			queue = queue[1:]
			continue
		}

		select {
		case task, ok := <-pendingCh:
			if !ok {
				pendingCh = nil
			} else {
				queue = append(queue, task)
			}
		case ps.taskCh <- queue[0]:
			queue = queue[1:]
		}
	}

	close(ps.taskCh)
}

func (ps *parallelScanner) start() {
	go ps.dispatcher()

	var workerWg sync.WaitGroup
	for i := 0; i < ps.workers; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for task := range ps.taskCh {
				ps.processDir(task)
				ps.wg.Done()
			}
		}()
	}

	go func() {
		ps.wg.Wait()
		close(ps.pendingCh)
		workerWg.Wait()
		close(ps.resultCh)
	}()
}

func (ps *parallelScanner) collect() []internal.FileInfo {
	var files []internal.FileInfo
	scanned := 0
	for fi := range ps.resultCh {
		files = append(files, fi)
		scanned++
		if ps.opts.OnProgress != nil {
			ps.opts.OnProgress(scanned, fi.Dir)
		}
	}
	return files
}

func (ps *parallelScanner) enqueue(path string) {
	ps.wg.Add(1)
	ps.pendingCh <- dirTask{path: path}
}

func (ps *parallelScanner) processDir(task dirTask) {
	entries, err := os.ReadDir(task.path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %s: %v\n", task.path, err)
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(task.path, name)

		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}

		if entry.IsDir() {
			if matchesExclude(name, ps.exactSet, ps.globs) {
				continue
			}
			if ps.opts.OneFileSystem {
				dev, devErr := deviceID(path)
				if devErr != nil {
					fmt.Fprintf(os.Stderr, "warning: %s: %v\n", path, devErr)
					continue
				}
				if dev != ps.rootDev {
					continue
				}
			}
			ps.wg.Add(1)
			ps.pendingCh <- dirTask{path: path}
			continue
		}

		if matchesExclude(name, ps.exactSet, ps.globs) {
			continue
		}

		info, infoErr := entry.Info()
		if infoErr != nil {
			fmt.Fprintf(os.Stderr, "warning: %s: %v\n", path, infoErr)
			continue
		}

		var physicalSize int64
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			physicalSize = stat.Blocks * 512
		} else {
			physicalSize = info.Size()
		}

		ps.resultCh <- internal.FileInfo{
			Path:         path,
			Size:         info.Size(),
			PhysicalSize: physicalSize,
			Dir:          task.path,
			Ext:          strings.ToLower(filepath.Ext(name)),
			ModTime:      info.ModTime(),
		}
	}
}
