package analyzer

import (
	"testing"

	"mogura/internal"
)

func TestAggregateByExt(t *testing.T) {
	tests := []struct {
		name  string
		files []internal.FileInfo
		want  map[string]ExtStats
	}{
		{
			name:  "empty input",
			files: nil,
			want:  map[string]ExtStats{},
		},
		{
			name: "single file",
			files: []internal.FileInfo{
				{Path: "/a/b.txt", Size: 100, Dir: "/a", Ext: ".txt"},
			},
			want: map[string]ExtStats{".txt": {Size: 100, Count: 1}},
		},
		{
			name: "multiple files same ext",
			files: []internal.FileInfo{
				{Path: "/a/b.txt", Size: 100, Dir: "/a", Ext: ".txt"},
				{Path: "/a/c.txt", Size: 200, Dir: "/a", Ext: ".txt"},
			},
			want: map[string]ExtStats{".txt": {Size: 300, Count: 2}},
		},
		{
			name: "multiple extensions",
			files: []internal.FileInfo{
				{Path: "/a/b.txt", Size: 100, Dir: "/a", Ext: ".txt"},
				{Path: "/a/c.go", Size: 50, Dir: "/a", Ext: ".go"},
				{Path: "/a/d.txt", Size: 200, Dir: "/a", Ext: ".txt"},
				{Path: "/a/e.go", Size: 75, Dir: "/a", Ext: ".go"},
			},
			want: map[string]ExtStats{
				".txt": {Size: 300, Count: 2},
				".go":  {Size: 125, Count: 2},
			},
		},
		{
			name: "no extension",
			files: []internal.FileInfo{
				{Path: "/a/Makefile", Size: 500, Dir: "/a", Ext: ""},
			},
			want: map[string]ExtStats{"(no ext)": {Size: 500, Count: 1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AggregateByExt(tt.files)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d exts, want %d", len(got), len(tt.want))
			}
			for ext, wantStats := range tt.want {
				gotStats := got[ext]
				if gotStats.Size != wantStats.Size {
					t.Errorf("ext %s: size got %d, want %d", ext, gotStats.Size, wantStats.Size)
				}
				if gotStats.Count != wantStats.Count {
					t.Errorf("ext %s: count got %d, want %d", ext, gotStats.Count, wantStats.Count)
				}
			}
		})
	}
}
