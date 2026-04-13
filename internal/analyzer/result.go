package analyzer

import "mogura/internal"

type Result struct {
	TotalSize     int64                    `json:"total_size"`
	DirSizes      map[string]int64         `json:"dir_sizes"`
	ExtStats      map[string]ExtStats      `json:"ext_stats"`
	CategoryStats map[Category]CategoryStats `json:"category_stats"`
	TopFiles      []internal.FileInfo      `json:"top_files"`
	DirTree       DirNode                  `json:"dir_tree"`
	WasteDirs     []WasteDir               `json:"waste_dirs"`
	StaleSummary  StaleResult              `json:"stale_summary"`
}
