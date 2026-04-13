package main

import (
	"fmt"
	"os"
	"time"

	"mogura/internal"
	"mogura/internal/analyzer"
	"mogura/internal/app"
	"mogura/internal/formatter"
	"mogura/internal/scanner"
)

func main() {
	cfg, err := app.ParseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	root := cfg.TargetPath

	files, err := scanner.Scan(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
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
	case app.FormatJSON:
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
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(out)

	case app.FormatTree:
		tree := analyzer.BuildTree(files)
		tree = analyzer.Prune(tree, cfg.Depth)
		fmt.Print(formatter.RenderTree(tree))

	default:
		fmt.Printf("Total: %s (%d files)\n\n", internal.FormatSize(totalSize), len(files))

		fmt.Println("=== ディレクトリ別 Top 10 ===")
		dirSizes := analyzer.AggregateByDir(files)
		formatter.PrintDirTable(os.Stdout, dirSizes, 10)

		fmt.Println()
		fmt.Println("=== 拡張子別 Top 10 ===")
		extStats := analyzer.AggregateByExt(files)
		formatter.PrintExtTable(os.Stdout, extStats, 10)

		fmt.Println()
		fmt.Println("=== カテゴリ別内訳 ===")
		formatter.PrintCategoryTable(os.Stdout, catStats)

		fmt.Println()
		fmt.Printf("=== 巨大ファイル Top %d ===\n", cfg.TopN)
		topFiles := analyzer.TopNFiles(files, cfg.TopN)
		formatter.PrintTopFiles(os.Stdout, topFiles)

		fmt.Println()
		fmt.Println("=== サマリ ===")
		summary := formatter.RenderSummary(formatter.SummaryInput{
			TotalSize:  totalSize,
			Categories: catStats,
			WasteDirs:  wasteDirs,
			Stale:      staleResult,
		})
		fmt.Print(summary)
	}
}
