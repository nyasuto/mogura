package analyzer

import (
	"math"
	"mogura/internal"
	"testing"
)

func TestAggregateByCategory(t *testing.T) {
	tests := []struct {
		name  string
		files []internal.FileInfo
		want  map[Category]CategoryStats
	}{
		{
			name:  "empty input",
			files: nil,
			want:  map[Category]CategoryStats{},
		},
		{
			name: "single category",
			files: []internal.FileInfo{
				{Path: "a.go", Size: 100, Ext: ".go"},
				{Path: "b.py", Size: 200, Ext: ".py"},
			},
			want: map[Category]CategoryStats{
				CategoryCode: {Size: 300, Count: 2, Percent: 100.0},
			},
		},
		{
			name: "multiple categories",
			files: []internal.FileInfo{
				{Path: "a.go", Size: 300, Ext: ".go"},
				{Path: "b.mp4", Size: 600, Ext: ".mp4"},
				{Path: "c.unknown", Size: 100, Ext: ".unknown"},
			},
			want: map[Category]CategoryStats{
				CategoryCode:  {Size: 300, Count: 1, Percent: 30.0},
				CategoryVideo: {Size: 600, Count: 1, Percent: 60.0},
				CategoryOther: {Size: 100, Count: 1, Percent: 10.0},
			},
		},
		{
			name: "no extension maps to other",
			files: []internal.FileInfo{
				{Path: "Makefile", Size: 500, Ext: ""},
			},
			want: map[Category]CategoryStats{
				CategoryOther: {Size: 500, Count: 1, Percent: 100.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AggregateByCategory(tt.files)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d categories, want %d", len(got), len(tt.want))
			}
			for cat, wantStats := range tt.want {
				gotStats, ok := got[cat]
				if !ok {
					t.Errorf("missing category %q", cat)
					continue
				}
				if gotStats.Size != wantStats.Size {
					t.Errorf("category %q: size = %d, want %d", cat, gotStats.Size, wantStats.Size)
				}
				if gotStats.Count != wantStats.Count {
					t.Errorf("category %q: count = %d, want %d", cat, gotStats.Count, wantStats.Count)
				}
				if math.Abs(gotStats.Percent-wantStats.Percent) > 0.01 {
					t.Errorf("category %q: percent = %.2f, want %.2f", cat, gotStats.Percent, wantStats.Percent)
				}
			}
		})
	}
}
