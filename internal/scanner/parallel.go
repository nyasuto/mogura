package scanner

import (
	"runtime"
	"sync"

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

	taskCh   chan dirTask
	resultCh chan internal.FileInfo
	wg       sync.WaitGroup
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
		opts:     opts,
		workers:  workers,
		rootPath: root,
		exactSet: exactSet,
		globs:    globs,
		taskCh:   make(chan dirTask, workers*4),
		resultCh: make(chan internal.FileInfo, workers*64),
	}
}

func (ps *parallelScanner) start() {
	for i := 0; i < ps.workers; i++ {
		ps.wg.Add(1)
		go func() {
			defer ps.wg.Done()
			for task := range ps.taskCh {
				ps.processDir(task)
			}
		}()
	}
}

func (ps *parallelScanner) wait() {
	ps.wg.Wait()
	close(ps.resultCh)
}

func (ps *parallelScanner) processDir(_ dirTask) {
	// Will be implemented in workerFn task
}
