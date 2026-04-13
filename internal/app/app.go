package app

import (
	"flag"
	"fmt"
	"io"
	"time"

	"mogura/internal"
	"mogura/internal/analyzer"
	"mogura/internal/formatter"
	"mogura/internal/scanner"
)

type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
	FormatTree OutputFormat = "tree"
	FormatHTML OutputFormat = "html"
)

type Config struct {
	TargetPath    string
	TopN          int
	Depth         int
	OutputFormat  OutputFormat
	Exclude       []string
	OlderThanDays int
}

func ParseFlags(args []string) (Config, error) {
	fs := flag.NewFlagSet("mogura", flag.ContinueOnError)

	jsonFlag := fs.Bool("json", false, "JSON 形式で出力")
	treeFlag := fs.Bool("tree", false, "ツリー形式で出力")
	depth := fs.Int("depth", 3, "ツリー表示の深さ")
	top := fs.Int("top", 20, "巨大ファイル表示件数")
	olderThan := fs.Int("older-than", 365, "古いファイルの判定日数")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "usage: mogura [flags] <path>\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	if fs.NArg() < 1 {
		fs.Usage()
		return Config{}, fmt.Errorf("path argument is required")
	}

	format := FormatText
	switch {
	case *jsonFlag:
		format = FormatJSON
	case *treeFlag:
		format = FormatTree
	}

	return Config{
		TargetPath:    fs.Arg(0),
		TopN:          *top,
		Depth:         *depth,
		OutputFormat:  format,
		OlderThanDays: *olderThan,
	}, nil
}

func Run(cfg Config, stdout io.Writer, stderr io.Writer) error {
	files, err := scanner.Scan(cfg.TargetPath)
	if err != nil {
		return err
	}

	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}

	now := time.Now()
	wasteDirs := analyzer.DetectWaste(files)
	wasteDirs = append(wasteDirs, analyzer.DetectLargeGitDirs(files, 100*1024*1024)...)
	staleResult := analyzer.DetectStale(files, cfg.OlderThanDays, now)
	catStats := analyzer.AggregateByCategory(files)

	var wasteTotal int64
	for _, w := range wasteDirs {
		wasteTotal += w.Size
	}
	savingsEstimate := wasteTotal + staleResult.TotalSize

	switch cfg.OutputFormat {
	case FormatJSON:
		tree := analyzer.BuildTree(files)
		tree = analyzer.Prune(tree, cfg.Depth)
		report := formatter.Report{
			TotalSize:       totalSize,
			ScannedAt:       now,
			DirTree:         tree,
			Extensions:      analyzer.AggregateByExt(files),
			Categories:      catStats,
			LargestFiles:    analyzer.TopNFiles(files, cfg.TopN),
			WasteDirs:       wasteDirs,
			StaleSummary:    &formatter.StaleSummary{TotalSize: staleResult.TotalSize, TotalFiles: staleResult.TotalFiles, DaysThreshold: cfg.OlderThanDays},
			SavingsEstimate: savingsEstimate,
		}
		out, err := formatter.RenderJSON(report)
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, out)

	case FormatTree:
		tree := analyzer.BuildTree(files)
		tree = analyzer.Prune(tree, cfg.Depth)
		fmt.Fprint(stdout, formatter.RenderTree(tree))

	default:
		fmt.Fprintf(stdout, "Total: %s (%d files)\n\n", internal.FormatSize(totalSize), len(files))

		fmt.Fprintln(stdout, "=== ディレクトリ別 Top 10 ===")
		dirSizes := analyzer.AggregateByDir(files)
		formatter.PrintDirTable(stdout, dirSizes, 10)

		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "=== 拡張子別 Top 10 ===")
		extStats := analyzer.AggregateByExt(files)
		formatter.PrintExtTable(stdout, extStats, 10)

		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "=== カテゴリ別内訳 ===")
		formatter.PrintCategoryTable(stdout, catStats)

		fmt.Fprintln(stdout)
		fmt.Fprintf(stdout, "=== 巨大ファイル Top %d ===\n", cfg.TopN)
		topFiles := analyzer.TopNFiles(files, cfg.TopN)
		formatter.PrintTopFiles(stdout, topFiles)

		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "=== サマリ ===")
		summary := formatter.RenderSummary(formatter.SummaryInput{
			TotalSize:  totalSize,
			Categories: catStats,
			WasteDirs:  wasteDirs,
			Stale:      staleResult,
		})
		fmt.Fprint(stdout, summary)
	}

	return nil
}
