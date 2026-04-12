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

func TestPrune(t *testing.T) {
	tree := DirNode{
		Name: "/root", Size: 1000, FileCount: 10,
		Children: []DirNode{
			{
				Name: "a", Size: 600, FileCount: 6,
				Children: []DirNode{
					{
						Name: "deep", Size: 400, FileCount: 4,
						Children: []DirNode{
							{Name: "deeper", Size: 200, FileCount: 2},
						},
					},
					{Name: "leaf", Size: 200, FileCount: 2},
				},
			},
			{Name: "b", Size: 400, FileCount: 4},
		},
	}

	tests := []struct {
		name           string
		depth          int
		wantRootKids   int
		wantSize       int64
		wantFileCount  int
		wantGrandKids  int
	}{
		{
			name:          "depth 0 removes all children",
			depth:         0,
			wantRootKids:  0,
			wantSize:      1000,
			wantFileCount: 10,
		},
		{
			name:          "depth 1 keeps direct children only",
			depth:         1,
			wantRootKids:  2,
			wantSize:      1000,
			wantFileCount: 10,
			wantGrandKids: 0,
		},
		{
			name:           "depth 2 keeps grandchildren",
			depth:          2,
			wantRootKids:   2,
			wantSize:       1000,
			wantFileCount:  10,
			wantGrandKids:  2,
		},
		{
			name:          "depth exceeding tree depth returns full tree",
			depth:         100,
			wantRootKids:  2,
			wantSize:      1000,
			wantFileCount: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Prune(tree, tt.depth)
			if got.Size != tt.wantSize {
				t.Errorf("Size = %d, want %d", got.Size, tt.wantSize)
			}
			if got.FileCount != tt.wantFileCount {
				t.Errorf("FileCount = %d, want %d", got.FileCount, tt.wantFileCount)
			}
			if len(got.Children) != tt.wantRootKids {
				t.Errorf("Children count = %d, want %d", len(got.Children), tt.wantRootKids)
			}
			if tt.wantGrandKids > 0 && len(got.Children) > 0 {
				grandKids := 0
				for _, c := range got.Children {
					grandKids += len(c.Children)
				}
				if grandKids != tt.wantGrandKids {
					t.Errorf("Grandchildren count = %d, want %d", grandKids, tt.wantGrandKids)
				}
			}
		})
	}
}

func TestPrunePreservesOriginal(t *testing.T) {
	tree := DirNode{
		Name: "/root", Size: 100, FileCount: 2,
		Children: []DirNode{
			{Name: "a", Size: 60, FileCount: 1},
			{Name: "b", Size: 40, FileCount: 1},
		},
	}

	pruned := Prune(tree, 0)
	if len(pruned.Children) != 0 {
		t.Errorf("pruned should have no children, got %d", len(pruned.Children))
	}
	if len(tree.Children) != 2 {
		t.Errorf("original tree should still have 2 children, got %d", len(tree.Children))
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
