package formatter

import (
	"encoding/json"
	"time"

	"mogura/internal"
	"mogura/internal/analyzer"
)

type StaleSummary struct {
	TotalSize     int64 `json:"total_size"`
	TotalFiles    int   `json:"total_files"`
	DaysThreshold int   `json:"days_threshold"`
}

type Report struct {
	TotalSize       int64                                        `json:"total_size"`
	ScannedAt       time.Time                                    `json:"scanned_at"`
	DirTree         analyzer.DirNode                             `json:"dir_tree"`
	Extensions      map[string]analyzer.ExtStats                 `json:"extensions"`
	Categories      map[analyzer.Category]analyzer.CategoryStats `json:"categories"`
	LargestFiles    []internal.FileInfo                          `json:"largest_files"`
	WasteDirs       []analyzer.WasteDir                          `json:"waste_dirs,omitempty"`
	StaleSummary    *StaleSummary                                `json:"stale_summary,omitempty"`
	SavingsEstimate int64                                        `json:"savings_estimate,omitempty"`
}

func RenderJSON(r Report) (string, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
