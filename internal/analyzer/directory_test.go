package analyzer

import (
	"testing"

	"mogura/internal"
)

func TestAggregateByDir(t *testing.T) {
	tests := []struct {
		name  string
		files []internal.FileInfo
		want  map[string]DirSizeInfo
	}{
		{
			name:  "empty input",
			files: nil,
			want:  map[string]DirSizeInfo{},
		},
		{
			name: "single file",
			files: []internal.FileInfo{
				{Path: "/a/b.txt", Size: 100, PhysicalSize: 80, Dir: "/a"},
			},
			want: map[string]DirSizeInfo{"/a": {Size: 100, PhysicalSize: 80}},
		},
		{
			name: "multiple files same dir",
			files: []internal.FileInfo{
				{Path: "/a/b.txt", Size: 100, PhysicalSize: 100, Dir: "/a"},
				{Path: "/a/c.txt", Size: 200, PhysicalSize: 150, Dir: "/a"},
			},
			want: map[string]DirSizeInfo{"/a": {Size: 300, PhysicalSize: 250}},
		},
		{
			name: "multiple dirs",
			files: []internal.FileInfo{
				{Path: "/a/b.txt", Size: 100, PhysicalSize: 100, Dir: "/a"},
				{Path: "/x/y.txt", Size: 50, PhysicalSize: 50, Dir: "/x"},
				{Path: "/a/c.txt", Size: 200, PhysicalSize: 200, Dir: "/a"},
				{Path: "/x/z.txt", Size: 75, PhysicalSize: 75, Dir: "/x"},
			},
			want: map[string]DirSizeInfo{"/a": {Size: 300, PhysicalSize: 300}, "/x": {Size: 125, PhysicalSize: 125}},
		},
		{
			name: "sparse file",
			files: []internal.FileInfo{
				{Path: "/d/big.raw", Size: 1000000, PhysicalSize: 4096, Dir: "/d"},
			},
			want: map[string]DirSizeInfo{"/d": {Size: 1000000, PhysicalSize: 4096}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AggregateByDir(tt.files)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d dirs, want %d", len(got), len(tt.want))
			}
			for dir, wantInfo := range tt.want {
				gotInfo := got[dir]
				if gotInfo.Size != wantInfo.Size {
					t.Errorf("dir %s: Size got %d, want %d", dir, gotInfo.Size, wantInfo.Size)
				}
				if gotInfo.PhysicalSize != wantInfo.PhysicalSize {
					t.Errorf("dir %s: PhysicalSize got %d, want %d", dir, gotInfo.PhysicalSize, wantInfo.PhysicalSize)
				}
			}
		})
	}
}
