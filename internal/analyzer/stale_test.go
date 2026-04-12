package analyzer

import (
	"testing"
	"time"

	"mogura/internal"
)

func TestDetectStale(t *testing.T) {
	now := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	old := now.AddDate(-2, 0, 0)
	recent := now.AddDate(0, -1, 0)

	tests := []struct {
		name      string
		files     []internal.FileInfo
		days      int
		wantTotal int64
		wantFiles int
		wantDirs  int
		wantFirst StaleDirSummary
	}{
		{
			name:      "empty input",
			files:     nil,
			days:      365,
			wantTotal: 0,
			wantFiles: 0,
			wantDirs:  0,
		},
		{
			name: "no stale files",
			files: []internal.FileInfo{
				{Path: "/a/b.txt", Size: 100, Dir: "/a", ModTime: recent},
			},
			days:      365,
			wantTotal: 0,
			wantFiles: 0,
			wantDirs:  0,
		},
		{
			name: "all stale",
			files: []internal.FileInfo{
				{Path: "/a/b.txt", Size: 100, Dir: "/a", ModTime: old},
				{Path: "/a/c.txt", Size: 200, Dir: "/a", ModTime: old},
			},
			days:      365,
			wantTotal: 300,
			wantFiles: 2,
			wantDirs:  1,
			wantFirst: StaleDirSummary{Dir: "/a", Size: 300, FileCount: 2},
		},
		{
			name: "mixed stale and recent",
			files: []internal.FileInfo{
				{Path: "/a/old.txt", Size: 500, Dir: "/a", ModTime: old},
				{Path: "/a/new.txt", Size: 100, Dir: "/a", ModTime: recent},
				{Path: "/b/old.txt", Size: 300, Dir: "/b", ModTime: old},
			},
			days:      365,
			wantTotal: 800,
			wantFiles: 2,
			wantDirs:  2,
			wantFirst: StaleDirSummary{Dir: "/a", Size: 500, FileCount: 1},
		},
		{
			name: "sorted by size descending",
			files: []internal.FileInfo{
				{Path: "/small/a.txt", Size: 100, Dir: "/small", ModTime: old},
				{Path: "/big/b.txt", Size: 1000, Dir: "/big", ModTime: old},
				{Path: "/mid/c.txt", Size: 500, Dir: "/mid", ModTime: old},
			},
			days:      365,
			wantTotal: 1600,
			wantFiles: 3,
			wantDirs:  3,
			wantFirst: StaleDirSummary{Dir: "/big", Size: 1000, FileCount: 1},
		},
		{
			name: "exact boundary not stale",
			files: []internal.FileInfo{
				{Path: "/a/exact.txt", Size: 100, Dir: "/a", ModTime: now.AddDate(0, 0, -365)},
			},
			days:      365,
			wantTotal: 0,
			wantFiles: 0,
			wantDirs:  0,
		},
		{
			name: "one day past boundary is stale",
			files: []internal.FileInfo{
				{Path: "/a/past.txt", Size: 100, Dir: "/a", ModTime: now.AddDate(0, 0, -366)},
			},
			days:      365,
			wantTotal: 100,
			wantFiles: 1,
			wantDirs:  1,
			wantFirst: StaleDirSummary{Dir: "/a", Size: 100, FileCount: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectStale(tt.files, tt.days, now)

			if got.TotalSize != tt.wantTotal {
				t.Errorf("TotalSize: got %d, want %d", got.TotalSize, tt.wantTotal)
			}
			if got.TotalFiles != tt.wantFiles {
				t.Errorf("TotalFiles: got %d, want %d", got.TotalFiles, tt.wantFiles)
			}
			if len(got.Dirs) != tt.wantDirs {
				t.Errorf("Dirs count: got %d, want %d", len(got.Dirs), tt.wantDirs)
			}
			if tt.wantDirs > 0 {
				first := got.Dirs[0]
				if first.Dir != tt.wantFirst.Dir {
					t.Errorf("first dir: got %q, want %q", first.Dir, tt.wantFirst.Dir)
				}
				if first.Size != tt.wantFirst.Size {
					t.Errorf("first size: got %d, want %d", first.Size, tt.wantFirst.Size)
				}
				if first.FileCount != tt.wantFirst.FileCount {
					t.Errorf("first file count: got %d, want %d", first.FileCount, tt.wantFirst.FileCount)
				}
			}
		})
	}
}
