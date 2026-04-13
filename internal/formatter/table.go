package formatter

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"mogura/internal"
	"mogura/internal/analyzer"
)

type Row []string

type Table struct {
	Header     Row
	Rows       []Row
	RightAlign []bool
}

func (t *Table) Render() string {
	if len(t.Header) == 0 {
		return ""
	}

	cols := len(t.Header)

	widths := make([]int, cols)
	for i, h := range t.Header {
		if len(h) > widths[i] {
			widths[i] = len(h)
		}
	}
	for _, row := range t.Rows {
		for i := 0; i < cols && i < len(row); i++ {
			if len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
	}

	var b strings.Builder

	formatRow := func(row Row) {
		for i := 0; i < cols; i++ {
			if i > 0 {
				b.WriteString("  ")
			}
			val := ""
			if i < len(row) {
				val = row[i]
			}
			if i < len(t.RightAlign) && t.RightAlign[i] {
				b.WriteString(fmt.Sprintf("%*s", widths[i], val))
			} else {
				b.WriteString(fmt.Sprintf("%-*s", widths[i], val))
			}
		}
		b.WriteByte('\n')
	}

	formatRow(t.Header)

	totalWidth := 0
	for i, w := range widths {
		totalWidth += w
		if i > 0 {
			totalWidth += 2
		}
	}
	b.WriteString(strings.Repeat("-", totalWidth))
	b.WriteByte('\n')

	for _, row := range t.Rows {
		formatRow(row)
	}

	return b.String()
}

type DirEntry struct {
	Path string
	Size int64
}

func PrintDirTable(w io.Writer, dirSizes map[string]int64, limit int) {
	entries := make([]DirEntry, 0, len(dirSizes))
	for path, size := range dirSizes {
		entries = append(entries, DirEntry{Path: path, Size: size})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Size > entries[j].Size
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	tbl := Table{
		Header:     Row{"Path", "Size"},
		RightAlign: []bool{false, true},
	}
	for _, e := range entries {
		tbl.Rows = append(tbl.Rows, Row{e.Path, internal.FormatSize(e.Size)})
	}
	fmt.Fprint(w, tbl.Render())
}

type ExtEntry struct {
	Ext   string
	Size  int64
	Count int
}

func PrintExtTable(w io.Writer, extStats map[string]analyzer.ExtStats, limit int) {
	entries := make([]ExtEntry, 0, len(extStats))
	for ext, s := range extStats {
		entries = append(entries, ExtEntry{Ext: ext, Size: s.Size, Count: s.Count})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Size > entries[j].Size
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	tbl := Table{
		Header:     Row{"Ext", "Size", "Count"},
		RightAlign: []bool{false, true, true},
	}
	for _, e := range entries {
		tbl.Rows = append(tbl.Rows, Row{e.Ext, internal.FormatSize(e.Size), fmt.Sprintf("%d", e.Count)})
	}
	fmt.Fprint(w, tbl.Render())
}

type CatEntry struct {
	Category analyzer.Category
	Size     int64
	Count    int
	Percent  float64
}

func PrintCategoryTable(w io.Writer, catStats map[analyzer.Category]analyzer.CategoryStats) {
	entries := make([]CatEntry, 0, len(catStats))
	for cat, s := range catStats {
		entries = append(entries, CatEntry{Category: cat, Size: s.Size, Count: s.Count, Percent: s.Percent})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Size > entries[j].Size
	})

	tbl := Table{
		Header:     Row{"Category", "Size", "Count", "%"},
		RightAlign: []bool{false, true, true, true},
	}
	for _, e := range entries {
		tbl.Rows = append(tbl.Rows, Row{
			string(e.Category),
			internal.FormatSize(e.Size),
			fmt.Sprintf("%d", e.Count),
			fmt.Sprintf("%.1f%%", e.Percent),
		})
	}
	fmt.Fprint(w, tbl.Render())
}

func FormatTable(result analyzer.Result, w io.Writer) {
	fmt.Fprintf(w, "Total: %s (%d files)\n\n", internal.FormatSize(result.TotalSize), result.FileCount)

	fmt.Fprintln(w, "=== ディレクトリ別 Top 10 ===")
	PrintDirTable(w, result.DirSizes, 10)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "=== 拡張子別 Top 10 ===")
	PrintExtTable(w, result.ExtStats, 10)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "=== カテゴリ別内訳 ===")
	PrintCategoryTable(w, result.CategoryStats)

	fmt.Fprintln(w)
	fmt.Fprintf(w, "=== 巨大ファイル Top %d ===\n", len(result.TopFiles))
	PrintTopFiles(w, result.TopFiles)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "=== サマリ ===")
	summary := RenderSummary(SummaryInput{
		TotalSize:  result.TotalSize,
		Categories: result.CategoryStats,
		WasteDirs:  result.WasteDirs,
		Stale:      result.StaleSummary,
	})
	fmt.Fprint(w, summary)
}

func PrintTopFiles(w io.Writer, files []internal.FileInfo) {
	if len(files) == 0 {
		return
	}

	tbl := Table{
		Header:     Row{"File", "Size"},
		RightAlign: []bool{false, true},
	}
	for _, f := range files {
		tbl.Rows = append(tbl.Rows, Row{f.Path, internal.FormatSize(f.Size)})
	}
	fmt.Fprint(w, tbl.Render())
}
