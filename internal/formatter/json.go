package formatter

import (
	"encoding/json"
	"fmt"
	"io"
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
	DiffSummary     []analyzer.DirDiff                           `json:"diff_summary,omitempty"`
}

func RenderJSON(r Report) (string, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func buildReport(result analyzer.Result) Report {
	var wasteTotal int64
	for _, wd := range result.WasteDirs {
		wasteTotal += wd.Size
	}

	return Report{
		TotalSize:    result.TotalSize,
		ScannedAt:    result.ScannedAt,
		DirTree:      result.DirTree,
		Extensions:   result.ExtStats,
		Categories:   result.CategoryStats,
		LargestFiles: result.TopFiles,
		WasteDirs:    result.WasteDirs,
		StaleSummary: &StaleSummary{
			TotalSize:     result.StaleSummary.TotalSize,
			TotalFiles:    result.StaleSummary.TotalFiles,
			DaysThreshold: result.OlderThanDays,
		},
		SavingsEstimate: wasteTotal + result.StaleSummary.TotalSize,
		DiffSummary:     result.DiffSummary,
	}
}

func FormatJSON(result analyzer.Result, w io.Writer) error {
	report := buildReport(result)
	out, err := RenderJSON(report)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, out)
	return err
}
