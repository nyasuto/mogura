package analyzer

import "mogura/internal"

type DirSizeInfo struct {
	Size         int64 `json:"size"`
	PhysicalSize int64 `json:"physical_size"`
}

func AggregateByDir(files []internal.FileInfo) map[string]DirSizeInfo {
	result := make(map[string]DirSizeInfo)
	for _, f := range files {
		info := result[f.Dir]
		info.Size += f.Size
		info.PhysicalSize += f.PhysicalSize
		result[f.Dir] = info
	}
	return result
}
