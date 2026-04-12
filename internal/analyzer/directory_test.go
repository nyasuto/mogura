package analyzer

import (
	"testing"

	"mogura/internal"
)

func TestAggregateByDir(t *testing.T) {
	tests := []struct {
		name  string
		files []internal.FileInfo
		want  map[string]int64
	}{
		{
			name:  "empty input",
			files: nil,
			want:  map[string]int64{},
		},
		{
			name: "single file",
			files: []internal.FileInfo{
				{Path: "/a/b.txt", Size: 100, Dir: "/a"},
			},
			want: map[string]int64{"/a": 100},
		},
		{
			name: "multiple files same dir",
			files: []internal.FileInfo{
				{Path: "/a/b.txt", Size: 100, Dir: "/a"},
				{Path: "/a/c.txt", Size: 200, Dir: "/a"},
			},
			want: map[string]int64{"/a": 300},
		},
		{
			name: "multiple dirs",
			files: []internal.FileInfo{
				{Path: "/a/b.txt", Size: 100, Dir: "/a"},
				{Path: "/x/y.txt", Size: 50, Dir: "/x"},
				{Path: "/a/c.txt", Size: 200, Dir: "/a"},
				{Path: "/x/z.txt", Size: 75, Dir: "/x"},
			},
			want: map[string]int64{"/a": 300, "/x": 125},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AggregateByDir(tt.files)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d dirs, want %d", len(got), len(tt.want))
			}
			for dir, wantSize := range tt.want {
				if got[dir] != wantSize {
					t.Errorf("dir %s: got %d, want %d", dir, got[dir], wantSize)
				}
			}
		})
	}
}
