package main

import (
	"fmt"
	"os"

	"mogura/internal"
	"mogura/internal/analyzer"
	"mogura/internal/formatter"
	"mogura/internal/scanner"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: mogura <path>")
		os.Exit(1)
	}

	root := os.Args[1]

	files, err := scanner.Scan(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}

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
	catStats := analyzer.AggregateByCategory(files)
	formatter.PrintCategoryTable(os.Stdout, catStats)

	fmt.Println()
	fmt.Println("=== 巨大ファイル Top 20 ===")
	topFiles := analyzer.TopNFiles(files, 20)
	formatter.PrintTopFiles(os.Stdout, topFiles)
}
