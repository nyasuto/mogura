package formatter

import (
	"fmt"
	"io"
	"sort"

	"mogura/internal"
)

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
