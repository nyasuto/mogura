package formatter

import (
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
