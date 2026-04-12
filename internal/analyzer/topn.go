package analyzer

import (
	"mogura/internal"
	"sort"
)

func TopNFiles(files []internal.FileInfo, n int) []internal.FileInfo {
	if n <= 0 {
		n = 20
	}

	sorted := make([]internal.FileInfo, len(files))
	copy(sorted, files)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Size > sorted[j].Size
	})

	if len(sorted) > n {
		sorted = sorted[:n]
	}
	return sorted
}
