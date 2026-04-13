package analyzer

import (
	"mogura/internal"
	"time"
)

type AnalyzeOpts struct {
	TopN             int
	Depth            int
	OlderThanDays    int
	GitSizeThreshold int64
	Now              time.Time
}

func Analyze(files []internal.FileInfo, opts AnalyzeOpts) Result {
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}
	if opts.TopN <= 0 {
		opts.TopN = 20
	}
	if opts.GitSizeThreshold <= 0 {
		opts.GitSizeThreshold = DefaultGitSizeThreshold
	}
	if opts.OlderThanDays <= 0 {
		opts.OlderThanDays = 365
	}

	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}

	wasteDirs := DetectWaste(files)
	wasteDirs = append(wasteDirs, DetectLargeGitDirs(files, opts.GitSizeThreshold)...)

	tree := BuildTree(files)
	tree = Prune(tree, opts.Depth)

	return Result{
		TotalSize:     totalSize,
		DirSizes:      AggregateByDir(files),
		ExtStats:      AggregateByExt(files),
		CategoryStats: AggregateByCategory(files),
		TopFiles:      TopNFiles(files, opts.TopN),
		DirTree:       tree,
		WasteDirs:     wasteDirs,
		StaleSummary:  DetectStale(files, opts.OlderThanDays, opts.Now),
	}
}
