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

func TestComputeDiff(t *testing.T) {
	tests := []struct {
		name      string
		prev      map[string]int64
		curr      map[string]int64
		wantCount int
		wantFirst DirDiff
		wantLast  DirDiff
	}{
		{
			name:      "directory grew",
			prev:      map[string]int64{"/home": 1000},
			curr:      map[string]int64{"/home": 3000},
			wantCount: 1,
			wantFirst: DirDiff{Path: "/home", PrevSize: 1000, CurrSize: 3000, Delta: 2000},
			wantLast:  DirDiff{Path: "/home", PrevSize: 1000, CurrSize: 3000, Delta: 2000},
		},
		{
			name:      "directory shrank",
			prev:      map[string]int64{"/tmp": 5000},
			curr:      map[string]int64{"/tmp": 2000},
			wantCount: 1,
			wantFirst: DirDiff{Path: "/tmp", PrevSize: 5000, CurrSize: 2000, Delta: -3000},
			wantLast:  DirDiff{Path: "/tmp", PrevSize: 5000, CurrSize: 2000, Delta: -3000},
		},
		{
			name:      "new directory",
			prev:      map[string]int64{},
			curr:      map[string]int64{"/new": 1000},
			wantCount: 1,
			wantFirst: DirDiff{Path: "/new", PrevSize: 0, CurrSize: 1000, Delta: 1000},
			wantLast:  DirDiff{Path: "/new", PrevSize: 0, CurrSize: 1000, Delta: 1000},
		},
		{
			name:      "deleted directory",
			prev:      map[string]int64{"/old": 2000},
			curr:      map[string]int64{},
			wantCount: 1,
			wantFirst: DirDiff{Path: "/old", PrevSize: 2000, CurrSize: 0, Delta: -2000},
			wantLast:  DirDiff{Path: "/old", PrevSize: 2000, CurrSize: 0, Delta: -2000},
		},
		{
			name:      "mixed changes sorted by delta desc",
			prev:      map[string]int64{"/a": 1000, "/b": 5000},
			curr:      map[string]int64{"/a": 4000, "/b": 2000, "/d": 500},
			wantCount: 3,
			wantFirst: DirDiff{Path: "/a", PrevSize: 1000, CurrSize: 4000, Delta: 3000},
			wantLast:  DirDiff{Path: "/b", PrevSize: 5000, CurrSize: 2000, Delta: -3000},
		},
		{
			name:      "both empty",
			prev:      map[string]int64{},
			curr:      map[string]int64{},
			wantCount: 0,
		},
		{
			name:      "no change",
			prev:      map[string]int64{"/same": 1000},
			curr:      map[string]int64{"/same": 1000},
			wantCount: 1,
			wantFirst: DirDiff{Path: "/same", PrevSize: 1000, CurrSize: 1000, Delta: 0},
			wantLast:  DirDiff{Path: "/same", PrevSize: 1000, CurrSize: 1000, Delta: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prev := Result{DirSizes: tt.prev}
			curr := Result{DirSizes: tt.curr}
			diffs := ComputeDiff(prev, curr)

			if len(diffs) != tt.wantCount {
				t.Fatalf("got %d diffs, want %d", len(diffs), tt.wantCount)
			}
			if tt.wantCount == 0 {
				return
			}

			first := diffs[0]
			if first != tt.wantFirst {
				t.Errorf("first diff = %+v, want %+v", first, tt.wantFirst)
			}
			last := diffs[len(diffs)-1]
			if last != tt.wantLast {
				t.Errorf("last diff = %+v, want %+v", last, tt.wantLast)
			}

			for i := 1; i < len(diffs); i++ {
				if diffs[i].Delta > diffs[i-1].Delta {
					t.Errorf("not sorted by delta desc: [%d].Delta=%d > [%d].Delta=%d",
						i, diffs[i].Delta, i-1, diffs[i-1].Delta)
				}
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
