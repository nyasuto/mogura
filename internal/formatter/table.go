package formatter

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"mogura/internal"
	"mogura/internal/analyzer"
)

const barWidth = 20

var ansiRe = regexp.MustCompile(`\033\[[0-9;]*m`)

func displayWidth(s string) int {
	return len(ansiRe.ReplaceAllString(s, ""))
}

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
		if dw := displayWidth(h); dw > widths[i] {
			widths[i] = dw
		}
	}
	for _, row := range t.Rows {
		for i := 0; i < cols && i < len(row); i++ {
			if dw := displayWidth(row[i]); dw > widths[i] {
				widths[i] = dw
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
			dw := displayWidth(val)
			pad := widths[i] - dw
			if pad < 0 {
				pad = 0
			}
			if i < len(t.RightAlign) && t.RightAlign[i] {
				b.WriteString(strings.Repeat(" ", pad))
				b.WriteString(val)
			} else {
				b.WriteString(val)
				b.WriteString(strings.Repeat(" ", pad))
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

func FormatSizeWithPhysical(size, physicalSize int64) string {
	if physicalSize > 0 && size > 0 && physicalSize < size*9/10 {
		return fmt.Sprintf("%s (実 %s)", internal.FormatSize(size), internal.FormatSize(physicalSize))
	}
	return internal.FormatSize(size)
}

type DirEntry struct {
	Path         string
	Size         int64
	PhysicalSize int64
}

func effectiveSize(size, physical int64, usePhysical bool) int64 {
	if usePhysical {
		return physical
	}
	return size
}

func formatSizeForMode(size, physical int64, usePhysical bool) string {
	if usePhysical {
		return FormatSizeWithPhysical(physical, size)
	}
	return FormatSizeWithPhysical(size, physical)
}

func PrintDirTable(w io.Writer, dirSizes map[string]analyzer.DirSizeInfo, limit int, usePhysical bool) {
	entries := make([]DirEntry, 0, len(dirSizes))
	for path, info := range dirSizes {
		entries = append(entries, DirEntry{Path: path, Size: info.Size, PhysicalSize: info.PhysicalSize})
	}

	sort.Slice(entries, func(i, j int) bool {
		return effectiveSize(entries[i].Size, entries[i].PhysicalSize, usePhysical) >
			effectiveSize(entries[j].Size, entries[j].PhysicalSize, usePhysical)
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	var maxSize int64
	if len(entries) > 0 {
		maxSize = effectiveSize(entries[0].Size, entries[0].PhysicalSize, usePhysical)
	}

	tbl := Table{
		Header:     Row{"Path", "Size", ""},
		RightAlign: []bool{false, true, false},
	}
	for _, e := range entries {
		s := effectiveSize(e.Size, e.PhysicalSize, usePhysical)
		bar := RenderBar(int(s), int(maxSize), barWidth)
		tbl.Rows = append(tbl.Rows, Row{e.Path, formatSizeForMode(e.Size, e.PhysicalSize, usePhysical), bar})
	}
	fmt.Fprint(w, tbl.Render())
}

type ExtEntry struct {
	Ext          string
	Size         int64
	PhysicalSize int64
	Count        int
}

func PrintExtTable(w io.Writer, extStats map[string]analyzer.ExtStats, limit int, usePhysical bool) {
	entries := make([]ExtEntry, 0, len(extStats))
	for ext, s := range extStats {
		entries = append(entries, ExtEntry{Ext: ext, Size: s.Size, PhysicalSize: s.PhysicalSize, Count: s.Count})
	}

	sort.Slice(entries, func(i, j int) bool {
		return effectiveSize(entries[i].Size, entries[i].PhysicalSize, usePhysical) >
			effectiveSize(entries[j].Size, entries[j].PhysicalSize, usePhysical)
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	var maxSize int64
	if len(entries) > 0 {
		maxSize = effectiveSize(entries[0].Size, entries[0].PhysicalSize, usePhysical)
	}

	tbl := Table{
		Header:     Row{"Ext", "Size", "Count", ""},
		RightAlign: []bool{false, true, true, false},
	}
	for _, e := range entries {
		s := effectiveSize(e.Size, e.PhysicalSize, usePhysical)
		bar := RenderBar(int(s), int(maxSize), barWidth)
		tbl.Rows = append(tbl.Rows, Row{e.Ext, formatSizeForMode(e.Size, e.PhysicalSize, usePhysical), fmt.Sprintf("%d", e.Count), bar})
	}
	fmt.Fprint(w, tbl.Render())
}

type CatEntry struct {
	Category     analyzer.Category
	Size         int64
	PhysicalSize int64
	Count        int
	Percent      float64
}

func PrintCategoryTable(w io.Writer, catStats map[analyzer.Category]analyzer.CategoryStats, usePhysical bool) {
	entries := make([]CatEntry, 0, len(catStats))
	for cat, s := range catStats {
		entries = append(entries, CatEntry{Category: cat, Size: s.Size, PhysicalSize: s.PhysicalSize, Count: s.Count, Percent: s.Percent})
	}

	sort.Slice(entries, func(i, j int) bool {
		return effectiveSize(entries[i].Size, entries[i].PhysicalSize, usePhysical) >
			effectiveSize(entries[j].Size, entries[j].PhysicalSize, usePhysical)
	})

	var maxSize int64
	if len(entries) > 0 {
		maxSize = effectiveSize(entries[0].Size, entries[0].PhysicalSize, usePhysical)
	}

	tbl := Table{
		Header:     Row{"Category", "Size", "Count", "%", ""},
		RightAlign: []bool{false, true, true, true, false},
	}
	for _, e := range entries {
		s := effectiveSize(e.Size, e.PhysicalSize, usePhysical)
		bar := RenderBar(int(s), int(maxSize), barWidth)
		tbl.Rows = append(tbl.Rows, Row{
			string(e.Category),
			formatSizeForMode(e.Size, e.PhysicalSize, usePhysical),
			fmt.Sprintf("%d", e.Count),
			fmt.Sprintf("%.1f%%", e.Percent),
			bar,
		})
	}
	fmt.Fprint(w, tbl.Render())
}

func FormatTable(result analyzer.Result, w io.Writer) {
	usePhysical := result.SizeMode == "physical"
	totalDisplay := result.TotalSize
	if usePhysical {
		totalDisplay = result.TotalPhysicalSize
	}

	fmt.Fprintf(w, "Total: %s (%d files)\n\n", internal.FormatSize(totalDisplay), result.FileCount)

	fmt.Fprintln(w, "=== ディレクトリ別 Top 10 ===")
	PrintDirTable(w, result.DirSizes, 10, usePhysical)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "=== 拡張子別 Top 10 ===")
	PrintExtTable(w, result.ExtStats, 10, usePhysical)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "=== カテゴリ別内訳 ===")
	PrintCategoryTable(w, result.CategoryStats, usePhysical)

	fmt.Fprintln(w)
	fmt.Fprintf(w, "=== 巨大ファイル Top %d ===\n", len(result.TopFiles))
	PrintTopFiles(w, result.TopFiles, usePhysical)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "=== サマリ ===")
	summary := RenderSummary(SummaryInput{
		TotalSize:         result.TotalSize,
		TotalPhysicalSize: result.TotalPhysicalSize,
		Categories:        result.CategoryStats,
		WasteDirs:         result.WasteDirs,
		Stale:             result.StaleSummary,
		SavingsEstimate:   result.SavingsEstimate,
	})
	fmt.Fprint(w, summary)
}

const (
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorReset = "\033[0m"
)

func formatDelta(delta int64) string {
	if delta > 0 {
		return colorRed + "+" + internal.FormatSize(delta) + colorReset
	}
	if delta < 0 {
		return colorGreen + "-" + internal.FormatSize(-delta) + colorReset
	}
	return "0 B"
}

func FormatDiffTable(diffs []analyzer.DirDiff, w io.Writer, limit int) {
	if limit > 0 && len(diffs) > limit {
		diffs = diffs[:limit]
	}

	tbl := Table{
		Header:     Row{"Path", "Prev", "Curr", "Delta"},
		RightAlign: []bool{false, true, true, true},
	}
	for _, d := range diffs {
		tbl.Rows = append(tbl.Rows, Row{
			d.Path,
			internal.FormatSize(d.PrevSize),
			internal.FormatSize(d.CurrSize),
			formatDelta(d.Delta),
		})
	}
	fmt.Fprint(w, tbl.Render())
}

func PrintTopFiles(w io.Writer, files []internal.FileInfo, usePhysical bool) {
	if len(files) == 0 {
		return
	}

	tbl := Table{
		Header:     Row{"File", "Size"},
		RightAlign: []bool{false, true},
	}
	for _, f := range files {
		tbl.Rows = append(tbl.Rows, Row{f.Path, formatSizeForMode(f.Size, f.PhysicalSize, usePhysical)})
	}
	fmt.Fprint(w, tbl.Render())
}
