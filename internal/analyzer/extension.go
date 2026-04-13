package analyzer

import "mogura/internal"

type ExtStats struct {
	Size         int64 `json:"size"`
	PhysicalSize int64 `json:"physical_size"`
	Count        int   `json:"count"`
}

func AggregateByExt(files []internal.FileInfo) map[string]ExtStats {
	result := make(map[string]ExtStats)
	for _, f := range files {
		ext := f.Ext
		if ext == "" {
			ext = "(no ext)"
		}
		s := result[ext]
		s.Size += f.Size
		s.PhysicalSize += f.PhysicalSize
		s.Count++
		result[ext] = s
	}
	return result
}
