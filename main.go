package main

import (
	"flag"
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
	jsonFlag := flag.Bool("json", false, "JSON 形式で出力")
	treeFlag := flag.Bool("tree", false, "ツリー形式で出力")
	depth := flag.Int("depth", 3, "ツリー表示の深さ")
	top := flag.Int("top", 20, "巨大ファイル表示件数")
	olderThan := flag.Int("older-than", 365, "古いファイルの判定日数")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: mogura [flags] <path>")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	root := flag.Arg(0)

	_ = app.Run // 後続タスクで使用

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
	staleResult := analyzer.DetectStale(files, *olderThan, now)
	catStats := analyzer.AggregateByCategory(files)

	var wasteTotal int64
	for _, w := range wasteDirs {
		wasteTotal += w.Size
	}
	savingsEstimate := wasteTotal + staleResult.TotalSize

	switch {
	case *jsonFlag:
		tree := analyzer.BuildTree(files)
		tree = analyzer.Prune(tree, *depth)
		report := formatter.Report{
			TotalSize:       totalSize,
			ScannedAt:       now,
			DirTree:         tree,
			Extensions:      analyzer.AggregateByExt(files),
			Categories:      catStats,
			LargestFiles:    analyzer.TopNFiles(files, *top),
			WasteDirs:       wasteDirs,
			StaleSummary:    &formatter.StaleSummary{TotalSize: staleResult.TotalSize, TotalFiles: staleResult.TotalFiles, DaysThreshold: *olderThan},
			SavingsEstimate: savingsEstimate,
		}
		out, err := formatter.RenderJSON(report)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(out)

	case *treeFlag:
		tree := analyzer.BuildTree(files)
		tree = analyzer.Prune(tree, *depth)
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
		fmt.Printf("=== 巨大ファイル Top %d ===\n", *top)
		topFiles := analyzer.TopNFiles(files, *top)
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
