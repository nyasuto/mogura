package analyzer

import "mogura/internal"

func AggregateByDir(files []internal.FileInfo) map[string]int64 {
	result := make(map[string]int64)
	for _, f := range files {
		result[f.Dir] += f.Size
	}
	return result
}
