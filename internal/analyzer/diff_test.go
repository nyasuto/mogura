package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPrevResult(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantTotal int64
		wantExts  int
		wantCats  int
		wantFiles int
		wantWaste int
		wantStale int64
		wantErr   bool
	}{
		{
			name: "valid report",
			json: `{
				"total_size": 1000000,
				"scanned_at": "2026-01-01T00:00:00Z",
				"dir_tree": {"name": "root", "size": 1000000, "file_count": 10},
				"extensions": {".go": {"size": 500, "count": 5}},
				"categories": {"コード": {"size": 500, "count": 5, "percent": 50}},
				"largest_files": [{"path": "/a.go", "size": 200, "dir": "/", "ext": ".go", "mod_time": "2026-01-01T00:00:00Z"}],
				"waste_dirs": [{"path": "/node_modules", "size": 300, "kind": "node_modules"}],
				"stale_summary": {"total_size": 100, "total_files": 2, "days_threshold": 365},
				"savings_estimate": 400
			}`,
			wantTotal: 1000000,
			wantExts:  1,
			wantCats:  1,
			wantFiles: 1,
			wantWaste: 1,
			wantStale: 100,
		},
		{
			name: "minimal report",
			json: `{
				"total_size": 500,
				"scanned_at": "2026-01-01T00:00:00Z",
				"dir_tree": {"name": "root", "size": 500, "file_count": 1}
			}`,
			wantTotal: 500,
		},
		{
			name:    "invalid json",
			json:    `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "prev.json")
			if err := os.WriteFile(path, []byte(tt.json), 0644); err != nil {
				t.Fatal(err)
			}

			result, err := LoadPrevResult(path)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.TotalSize != tt.wantTotal {
				t.Errorf("TotalSize = %d, want %d", result.TotalSize, tt.wantTotal)
			}
			if len(result.ExtStats) != tt.wantExts {
				t.Errorf("ExtStats count = %d, want %d", len(result.ExtStats), tt.wantExts)
			}
			if len(result.CategoryStats) != tt.wantCats {
				t.Errorf("CategoryStats count = %d, want %d", len(result.CategoryStats), tt.wantCats)
			}
			if len(result.TopFiles) != tt.wantFiles {
				t.Errorf("TopFiles count = %d, want %d", len(result.TopFiles), tt.wantFiles)
			}
			if len(result.WasteDirs) != tt.wantWaste {
				t.Errorf("WasteDirs count = %d, want %d", len(result.WasteDirs), tt.wantWaste)
			}
			if result.StaleSummary.TotalSize != tt.wantStale {
				t.Errorf("StaleSummary.TotalSize = %d, want %d", result.StaleSummary.TotalSize, tt.wantStale)
			}
		})
	}
}

func TestLoadPrevResult_FileNotFound(t *testing.T) {
	_, err := LoadPrevResult("/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}
