package analyzer

import (
	"testing"

	"mogura/internal"
)

func TestBuildTree(t *testing.T) {
	tests := []struct {
		name      string
		files     []internal.FileInfo
		wantName  string
		wantSize  int64
		wantCount int
		wantKids  int
	}{
		{
			name:      "empty input",
			files:     nil,
			wantName:  "",
			wantSize:  0,
			wantCount: 0,
			wantKids:  0,
		},
		{
			name: "single file",
			files: []internal.FileInfo{
				{Path: "/a/b.txt", Size: 100, Dir: "/a"},
			},
			wantName:  "/a",
			wantSize:  100,
			wantCount: 1,
			wantKids:  0,
		},
		{
			name: "multiple files same dir",
			files: []internal.FileInfo{
				{Path: "/a/b.txt", Size: 100, Dir: "/a"},
				{Path: "/a/c.txt", Size: 200, Dir: "/a"},
			},
			wantName:  "/a",
			wantSize:  300,
			wantCount: 2,
			wantKids:  0,
		},
		{
			name: "nested directories",
			files: []internal.FileInfo{
				{Path: "/a/b.txt", Size: 100, Dir: "/a"},
				{Path: "/a/sub/c.txt", Size: 200, Dir: "/a/sub"},
			},
			wantName:  "/a",
			wantSize:  300,
			wantCount: 2,
			wantKids:  1,
		},
		{
			name: "multiple sibling dirs",
			files: []internal.FileInfo{
				{Path: "/root/a/x.txt", Size: 100, Dir: "/root/a"},
				{Path: "/root/b/y.txt", Size: 300, Dir: "/root/b"},
				{Path: "/root/c/z.txt", Size: 50, Dir: "/root/c"},
			},
			wantName:  "/root",
			wantSize:  450,
			wantCount: 3,
			wantKids:  3,
		},
		{
			name: "deep nesting",
			files: []internal.FileInfo{
				{Path: "/a/b/c/d.txt", Size: 500, Dir: "/a/b/c"},
				{Path: "/a/b/e.txt", Size: 100, Dir: "/a/b"},
			},
			wantName:  "/a/b",
			wantSize:  600,
			wantCount: 2,
			wantKids:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildTree(tt.files)
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Size != tt.wantSize {
				t.Errorf("Size = %d, want %d", got.Size, tt.wantSize)
			}
			if got.FileCount != tt.wantCount {
				t.Errorf("FileCount = %d, want %d", got.FileCount, tt.wantCount)
			}
			if len(got.Children) != tt.wantKids {
				t.Errorf("Children count = %d, want %d", len(got.Children), tt.wantKids)
			}
		})
	}
}

func TestBuildTreeChildrenSortedBySize(t *testing.T) {
	files := []internal.FileInfo{
		{Path: "/root/a/x.txt", Size: 100, Dir: "/root/a"},
		{Path: "/root/b/y.txt", Size: 300, Dir: "/root/b"},
		{Path: "/root/c/z.txt", Size: 50, Dir: "/root/c"},
	}

	got := BuildTree(files)
	if len(got.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(got.Children))
	}
	if got.Children[0].Name != "b" {
		t.Errorf("first child = %q, want %q", got.Children[0].Name, "b")
	}
	if got.Children[1].Name != "a" {
		t.Errorf("second child = %q, want %q", got.Children[1].Name, "a")
	}
	if got.Children[2].Name != "c" {
		t.Errorf("third child = %q, want %q", got.Children[2].Name, "c")
	}
}

func TestBuildTreeSizeAggregation(t *testing.T) {
	files := []internal.FileInfo{
		{Path: "/a/x.txt", Size: 100, Dir: "/a"},
		{Path: "/a/sub1/y.txt", Size: 200, Dir: "/a/sub1"},
		{Path: "/a/sub1/deep/z.txt", Size: 300, Dir: "/a/sub1/deep"},
		{Path: "/a/sub2/w.txt", Size: 50, Dir: "/a/sub2"},
	}

	got := BuildTree(files)
	if got.Size != 650 {
		t.Errorf("root Size = %d, want 650", got.Size)
	}
	if got.FileCount != 4 {
		t.Errorf("root FileCount = %d, want 4", got.FileCount)
	}

	// sub1 should have size 500 (200 + 300) and filecount 2
	var sub1 DirNode
	for _, c := range got.Children {
		if c.Name == "sub1" {
			sub1 = c
			break
		}
	}
	if sub1.Size != 500 {
		t.Errorf("sub1 Size = %d, want 500", sub1.Size)
	}
	if sub1.FileCount != 2 {
		t.Errorf("sub1 FileCount = %d, want 2", sub1.FileCount)
	}
}
