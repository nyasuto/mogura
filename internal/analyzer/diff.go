package analyzer

import (
	"encoding/json"
	"fmt"
	"mogura/internal"
	"os"
	"sort"
	"time"
)

type DirDiff struct {
	Path     string `json:"path"`
	PrevSize int64  `json:"prev_size"`
	CurrSize int64  `json:"curr_size"`
	Delta    int64  `json:"delta"`
}

type jsonReport struct {
	TotalSize    int64                      `json:"total_size"`
	ScannedAt    time.Time                  `json:"scanned_at"`
	DirTree      DirNode                    `json:"dir_tree"`
	Extensions   map[string]ExtStats        `json:"extensions"`
	Categories   map[Category]CategoryStats `json:"categories"`
	LargestFiles []internal.FileInfo        `json:"largest_files"`
	WasteDirs    []WasteDir                 `json:"waste_dirs"`
	StaleSummary *struct {
		TotalSize     int64 `json:"total_size"`
		TotalFiles    int   `json:"total_files"`
		DaysThreshold int   `json:"days_threshold"`
	} `json:"stale_summary"`
	SavingsEstimate int64 `json:"savings_estimate"`
}

func ComputeDiff(prev, curr Result) []DirDiff {
	seen := make(map[string]bool)
	var diffs []DirDiff

	for path, currInfo := range curr.DirSizes {
		seen[path] = true
		prevInfo := prev.DirSizes[path]
		diffs = append(diffs, DirDiff{
			Path:     path,
			PrevSize: prevInfo.Size,
			CurrSize: currInfo.Size,
			Delta:    currInfo.Size - prevInfo.Size,
		})
	}

	for path, prevInfo := range prev.DirSizes {
		if seen[path] {
			continue
		}
		diffs = append(diffs, DirDiff{
			Path:     path,
			PrevSize: prevInfo.Size,
			CurrSize: 0,
			Delta:    -prevInfo.Size,
		})
	}

	sort.Slice(diffs, func(i, j int) bool {
		return diffs[i].Delta > diffs[j].Delta
	})

	return diffs
}

func LoadPrevResult(path string) (Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Result{}, fmt.Errorf("failed to read previous result: %w", err)
	}

	var report jsonReport
	if err := json.Unmarshal(data, &report); err != nil {
		return Result{}, fmt.Errorf("failed to parse previous result: %w", err)
	}

	result := Result{
		TotalSize:     report.TotalSize,
		ScannedAt:     report.ScannedAt,
		DirTree:       report.DirTree,
		ExtStats:      report.Extensions,
		CategoryStats: report.Categories,
		TopFiles:      report.LargestFiles,
		WasteDirs:     report.WasteDirs,
	}

	if report.StaleSummary != nil {
		result.OlderThanDays = report.StaleSummary.DaysThreshold
		result.StaleSummary = StaleResult{
			TotalSize:  report.StaleSummary.TotalSize,
			TotalFiles: report.StaleSummary.TotalFiles,
		}
	}

	return result, nil
}
