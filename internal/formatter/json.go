package formatter

import (
	"encoding/json"
	"time"

	"mogura/internal"
	"mogura/internal/analyzer"
)

type Report struct {
	TotalSize    int64                                  `json:"total_size"`
	ScannedAt    time.Time                              `json:"scanned_at"`
	DirTree      analyzer.DirNode                       `json:"dir_tree"`
	Extensions   map[string]analyzer.ExtStats           `json:"extensions"`
	Categories   map[analyzer.Category]analyzer.CategoryStats `json:"categories"`
	LargestFiles []internal.FileInfo                    `json:"largest_files"`
}

func RenderJSON(r Report) (string, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
