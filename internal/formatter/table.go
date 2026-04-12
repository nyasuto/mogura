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

	maxPathLen := 4 // "Path" header
	for _, e := range entries {
		if len(e.Path) > maxPathLen {
			maxPathLen = len(e.Path)
		}
	}

	fmt.Fprintf(w, "%-*s  %10s\n", maxPathLen, "Path", "Size")
	for i := 0; i < maxPathLen+12; i++ {
		fmt.Fprint(w, "-")
	}
	fmt.Fprintln(w)

	for _, e := range entries {
		fmt.Fprintf(w, "%-*s  %10s\n", maxPathLen, e.Path, internal.FormatSize(e.Size))
	}
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

	maxExtLen := 3
	for _, e := range entries {
		if len(e.Ext) > maxExtLen {
			maxExtLen = len(e.Ext)
		}
	}

	fmt.Fprintf(w, "%-*s  %10s  %6s\n", maxExtLen, "Ext", "Size", "Count")
	for i := 0; i < maxExtLen+20; i++ {
		fmt.Fprint(w, "-")
	}
	fmt.Fprintln(w)

	for _, e := range entries {
		fmt.Fprintf(w, "%-*s  %10s  %6d\n", maxExtLen, e.Ext, internal.FormatSize(e.Size), e.Count)
	}
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

	maxCatLen := 8
	for _, e := range entries {
		if len(string(e.Category)) > maxCatLen {
			maxCatLen = len(string(e.Category))
		}
	}

	fmt.Fprintf(w, "%-*s  %10s  %6s  %6s\n", maxCatLen, "Category", "Size", "Count", "%")
	for i := 0; i < maxCatLen+28; i++ {
		fmt.Fprint(w, "-")
	}
	fmt.Fprintln(w)

	for _, e := range entries {
		fmt.Fprintf(w, "%-*s  %10s  %6d  %5.1f%%\n", maxCatLen, string(e.Category), internal.FormatSize(e.Size), e.Count, e.Percent)
	}
}

func PrintTopFiles(w io.Writer, files []internal.FileInfo) {
	if len(files) == 0 {
		return
	}

	maxPathLen := 4
	for _, f := range files {
		if len(f.Path) > maxPathLen {
			maxPathLen = len(f.Path)
		}
	}

	fmt.Fprintf(w, "%-*s  %10s\n", maxPathLen, "File", "Size")
	for i := 0; i < maxPathLen+12; i++ {
		fmt.Fprint(w, "-")
	}
	fmt.Fprintln(w)

	for _, f := range files {
		fmt.Fprintf(w, "%-*s  %10s\n", maxPathLen, f.Path, internal.FormatSize(f.Size))
	}
}
