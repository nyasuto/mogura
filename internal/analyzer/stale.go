package analyzer

import (
	"mogura/internal"
	"sort"
	"time"
)

type StaleDirSummary struct {
	Dir       string `json:"dir"`
	Size      int64  `json:"size"`
	FileCount int    `json:"file_count"`
}

type StaleResult struct {
	TotalSize  int64             `json:"total_size"`
	TotalFiles int               `json:"total_files"`
	Dirs       []StaleDirSummary `json:"dirs"`
}

func DetectStale(files []internal.FileInfo, days int, now time.Time) StaleResult {
	cutoff := now.AddDate(0, 0, -days)

	dirSizes := make(map[string]int64)
	dirCounts := make(map[string]int)

	for _, f := range files {
		if f.ModTime.Before(cutoff) {
			dirSizes[f.Dir] += f.Size
			dirCounts[f.Dir]++
		}
	}

	result := StaleResult{}
	for dir, size := range dirSizes {
		result.Dirs = append(result.Dirs, StaleDirSummary{
			Dir:       dir,
			Size:      size,
			FileCount: dirCounts[dir],
		})
		result.TotalSize += size
		result.TotalFiles += dirCounts[dir]
	}

	sort.Slice(result.Dirs, func(i, j int) bool {
		return result.Dirs[i].Size > result.Dirs[j].Size
	})

	return result
}
