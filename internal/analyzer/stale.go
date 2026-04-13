package analyzer

import (
	"mogura/internal"
	"sort"
	"time"
)

type StaleDirSummary struct {
	Dir          string `json:"dir"`
	Size         int64  `json:"size"`
	PhysicalSize int64  `json:"physical_size"`
	FileCount    int    `json:"file_count"`
}

type StaleResult struct {
	TotalSize         int64             `json:"total_size"`
	TotalPhysicalSize int64             `json:"total_physical_size"`
	TotalFiles        int               `json:"total_files"`
	Dirs              []StaleDirSummary `json:"dirs"`
}

func DetectStale(files []internal.FileInfo, days int, now time.Time) StaleResult {
	cutoff := now.AddDate(0, 0, -days)

	dirSizes := make(map[string]int64)
	dirPhysical := make(map[string]int64)
	dirCounts := make(map[string]int)

	for _, f := range files {
		if f.ModTime.Before(cutoff) {
			dirSizes[f.Dir] += f.Size
			dirPhysical[f.Dir] += f.PhysicalSize
			dirCounts[f.Dir]++
		}
	}

	result := StaleResult{}
	for dir, size := range dirSizes {
		result.Dirs = append(result.Dirs, StaleDirSummary{
			Dir:          dir,
			Size:         size,
			PhysicalSize: dirPhysical[dir],
			FileCount:    dirCounts[dir],
		})
		result.TotalSize += size
		result.TotalPhysicalSize += dirPhysical[dir]
		result.TotalFiles += dirCounts[dir]
	}

	sort.Slice(result.Dirs, func(i, j int) bool {
		return result.Dirs[i].Size > result.Dirs[j].Size
	})

	return result
}
