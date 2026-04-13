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
				{Path: "/a/b.txt", Size: 100, PhysicalSize: 80, Dir: "/a", Ext: ".txt"},
			},
			want: map[string]ExtStats{".txt": {Size: 100, PhysicalSize: 80, Count: 1}},
		},
		{
			name: "multiple files same ext",
			files: []internal.FileInfo{
				{Path: "/a/b.txt", Size: 100, PhysicalSize: 90, Dir: "/a", Ext: ".txt"},
				{Path: "/a/c.txt", Size: 200, PhysicalSize: 150, Dir: "/a", Ext: ".txt"},
			},
			want: map[string]ExtStats{".txt": {Size: 300, PhysicalSize: 240, Count: 2}},
		},
		{
			name: "multiple extensions",
			files: []internal.FileInfo{
				{Path: "/a/b.txt", Size: 100, PhysicalSize: 100, Dir: "/a", Ext: ".txt"},
				{Path: "/a/c.go", Size: 50, PhysicalSize: 50, Dir: "/a", Ext: ".go"},
				{Path: "/a/d.txt", Size: 200, PhysicalSize: 200, Dir: "/a", Ext: ".txt"},
				{Path: "/a/e.go", Size: 75, PhysicalSize: 75, Dir: "/a", Ext: ".go"},
			},
			want: map[string]ExtStats{
				".txt": {Size: 300, PhysicalSize: 300, Count: 2},
				".go":  {Size: 125, PhysicalSize: 125, Count: 2},
			},
		},
		{
			name: "no extension",
			files: []internal.FileInfo{
				{Path: "/a/Makefile", Size: 500, PhysicalSize: 500, Dir: "/a", Ext: ""},
			},
			want: map[string]ExtStats{"(no ext)": {Size: 500, PhysicalSize: 500, Count: 1}},
		},
		{
			name: "sparse file physical size aggregation",
			files: []internal.FileInfo{
				{Path: "/a/big.raw", Size: 1000, PhysicalSize: 100, Dir: "/a", Ext: ".raw"},
				{Path: "/a/small.raw", Size: 200, PhysicalSize: 200, Dir: "/a", Ext: ".raw"},
			},
			want: map[string]ExtStats{".raw": {Size: 1200, PhysicalSize: 300, Count: 2}},
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
				if gotStats.PhysicalSize != wantStats.PhysicalSize {
					t.Errorf("ext %s: physical size got %d, want %d", ext, gotStats.PhysicalSize, wantStats.PhysicalSize)
				}
				if gotStats.Count != wantStats.Count {
					t.Errorf("ext %s: count got %d, want %d", ext, gotStats.Count, wantStats.Count)
				}
			}
		})
	}
}
