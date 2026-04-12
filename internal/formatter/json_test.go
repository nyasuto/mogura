package formatter

import (
	"encoding/json"
	"testing"
	"time"

	"mogura/internal"
	"mogura/internal/analyzer"
)

func TestRenderJSON(t *testing.T) {
	fixedTime := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name   string
		report Report
	}{
		{
			name: "empty report",
			report: Report{
				TotalSize:    0,
				ScannedAt:    fixedTime,
				DirTree:      analyzer.DirNode{},
				Extensions:   map[string]analyzer.ExtStats{},
				Categories:   map[analyzer.Category]analyzer.CategoryStats{},
				LargestFiles: []internal.FileInfo{},
			},
		},
		{
			name: "report with data",
			report: Report{
				TotalSize: 1024000,
				ScannedAt: fixedTime,
				DirTree: analyzer.DirNode{
					Name: "root",
					Size: 1024000,
					Children: []analyzer.DirNode{
						{Name: "src", Size: 512000, FileCount: 10},
						{Name: "docs", Size: 512000, FileCount: 5},
					},
					FileCount: 15,
				},
				Extensions: map[string]analyzer.ExtStats{
					".go":  {Size: 512000, Count: 10},
					".md":  {Size: 512000, Count: 5},
				},
				Categories: map[analyzer.Category]analyzer.CategoryStats{
					analyzer.CategoryCode:     {Size: 512000, Count: 10, Percent: 50.0},
					analyzer.CategoryDocument: {Size: 512000, Count: 5, Percent: 50.0},
				},
				LargestFiles: []internal.FileInfo{
					{Path: "/root/src/main.go", Size: 102400, Dir: "/root/src", Ext: ".go"},
				},
			},
		},
		{
			name: "report with waste and stale",
			report: Report{
				TotalSize: 2048000,
				ScannedAt: fixedTime,
				DirTree:   analyzer.DirNode{Name: "root", Size: 2048000},
				Extensions: map[string]analyzer.ExtStats{},
				Categories: map[analyzer.Category]analyzer.CategoryStats{},
				LargestFiles: []internal.FileInfo{},
				WasteDirs: []analyzer.WasteDir{
					{Path: "/root/node_modules", Size: 500000, Kind: "node_modules"},
					{Path: "/root/.cache", Size: 300000, Kind: "cache"},
				},
				StaleSummary: &StaleSummary{
					TotalSize:     200000,
					TotalFiles:    50,
					DaysThreshold: 365,
				},
				SavingsEstimate: 1000000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderJSON(tt.report)
			if err != nil {
				t.Fatalf("RenderJSON returned error: %v", err)
			}

			if got == "" {
				t.Fatal("RenderJSON returned empty string")
			}

			var parsed Report
			if err := json.Unmarshal([]byte(got), &parsed); err != nil {
				t.Fatalf("output is not valid JSON: %v", err)
			}

			if parsed.TotalSize != tt.report.TotalSize {
				t.Errorf("TotalSize = %d, want %d", parsed.TotalSize, tt.report.TotalSize)
			}

			if !parsed.ScannedAt.Equal(tt.report.ScannedAt) {
				t.Errorf("ScannedAt = %v, want %v", parsed.ScannedAt, tt.report.ScannedAt)
			}

			if parsed.DirTree.Name != tt.report.DirTree.Name {
				t.Errorf("DirTree.Name = %q, want %q", parsed.DirTree.Name, tt.report.DirTree.Name)
			}

			if len(parsed.DirTree.Children) != len(tt.report.DirTree.Children) {
				t.Errorf("DirTree.Children count = %d, want %d", len(parsed.DirTree.Children), len(tt.report.DirTree.Children))
			}

			if len(parsed.LargestFiles) != len(tt.report.LargestFiles) {
				t.Errorf("LargestFiles count = %d, want %d", len(parsed.LargestFiles), len(tt.report.LargestFiles))
			}

			if len(parsed.WasteDirs) != len(tt.report.WasteDirs) {
				t.Errorf("WasteDirs count = %d, want %d", len(parsed.WasteDirs), len(tt.report.WasteDirs))
			}

			if (parsed.StaleSummary == nil) != (tt.report.StaleSummary == nil) {
				t.Errorf("StaleSummary nil = %v, want %v", parsed.StaleSummary == nil, tt.report.StaleSummary == nil)
			} else if parsed.StaleSummary != nil {
				if parsed.StaleSummary.TotalSize != tt.report.StaleSummary.TotalSize {
					t.Errorf("StaleSummary.TotalSize = %d, want %d", parsed.StaleSummary.TotalSize, tt.report.StaleSummary.TotalSize)
				}
				if parsed.StaleSummary.TotalFiles != tt.report.StaleSummary.TotalFiles {
					t.Errorf("StaleSummary.TotalFiles = %d, want %d", parsed.StaleSummary.TotalFiles, tt.report.StaleSummary.TotalFiles)
				}
				if parsed.StaleSummary.DaysThreshold != tt.report.StaleSummary.DaysThreshold {
					t.Errorf("StaleSummary.DaysThreshold = %d, want %d", parsed.StaleSummary.DaysThreshold, tt.report.StaleSummary.DaysThreshold)
				}
			}

			if parsed.SavingsEstimate != tt.report.SavingsEstimate {
				t.Errorf("SavingsEstimate = %d, want %d", parsed.SavingsEstimate, tt.report.SavingsEstimate)
			}
		})
	}
}

func TestRenderJSON_indent(t *testing.T) {
	r := Report{
		TotalSize: 100,
		ScannedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		DirTree:   analyzer.DirNode{Name: "root", Size: 100},
	}

	got, err := RenderJSON(r)
	if err != nil {
		t.Fatalf("RenderJSON returned error: %v", err)
	}

	if got[0] != '{' {
		t.Error("output should start with '{'")
	}

	if !json.Valid([]byte(got)) {
		t.Error("output is not valid JSON")
	}

	if len(got) < 10 {
		t.Error("output seems too short for indented JSON")
	}
}
