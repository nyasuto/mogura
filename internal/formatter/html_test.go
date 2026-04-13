package formatter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"mogura/internal"
	"mogura/internal/analyzer"
)

func TestFormatHTML_containsData(t *testing.T) {
	result := analyzer.Result{
		TotalSize: 1024000,
		FileCount: 15,
		ScannedAt: time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
		DirSizes: map[string]int64{
			"/root/src":  512000,
			"/root/docs": 512000,
		},
		ExtStats: map[string]analyzer.ExtStats{
			".go": {Size: 512000, Count: 10},
		},
		CategoryStats: map[analyzer.Category]analyzer.CategoryStats{
			analyzer.CategoryCode: {Size: 512000, Count: 10, Percent: 50.0},
		},
		TopFiles: []internal.FileInfo{
			{Path: "/root/src/main.go", Size: 102400, Dir: "/root/src", Ext: ".go"},
		},
		DirTree: analyzer.DirNode{
			Name: "root",
			Size: 1024000,
			Children: []analyzer.DirNode{
				{Name: "src", Size: 512000, FileCount: 10},
				{Name: "docs", Size: 512000, FileCount: 5},
			},
			FileCount: 15,
		},
	}

	var buf bytes.Buffer
	err := FormatHTML(result, &buf)
	if err != nil {
		t.Fatalf("FormatHTML returned error: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "const data =") {
		t.Error("output does not contain 'const data ='")
	}

	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("output does not contain '<!DOCTYPE html>'")
	}

	if !strings.Contains(html, "Mogura") {
		t.Error("output does not contain 'Mogura'")
	}

	if !strings.Contains(html, "total_size") {
		t.Error("output does not contain 'total_size' in JSON data")
	}

	if !strings.Contains(html, "1024000") {
		t.Error("output does not contain expected total_size value")
	}
}

func TestFormatHTML_emptyResult(t *testing.T) {
	result := analyzer.Result{
		ScannedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		DirTree:   analyzer.DirNode{Name: "root"},
	}

	var buf bytes.Buffer
	err := FormatHTML(result, &buf)
	if err != nil {
		t.Fatalf("FormatHTML returned error: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "const data =") {
		t.Error("output does not contain 'const data ='")
	}
}
