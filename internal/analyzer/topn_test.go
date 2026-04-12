package analyzer

import (
	"mogura/internal"
	"testing"
)

func TestTopNFiles(t *testing.T) {
	files := []internal.FileInfo{
		{Path: "/a.txt", Size: 100},
		{Path: "/b.txt", Size: 500},
		{Path: "/c.txt", Size: 300},
		{Path: "/d.txt", Size: 200},
		{Path: "/e.txt", Size: 400},
	}

	tests := []struct {
		name      string
		n         int
		wantLen   int
		wantFirst string
		wantLast  string
	}{
		{"top 3", 3, 3, "/b.txt", "/c.txt"},
		{"top 10 (more than available)", 10, 5, "/b.txt", "/a.txt"},
		{"default (n=0)", 0, 5, "/b.txt", "/a.txt"},
		{"top 1", 1, 1, "/b.txt", "/b.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TopNFiles(files, tt.n)
			if len(got) != tt.wantLen {
				t.Errorf("len = %d, want %d", len(got), tt.wantLen)
			}
			if got[0].Path != tt.wantFirst {
				t.Errorf("first = %s, want %s", got[0].Path, tt.wantFirst)
			}
			if got[len(got)-1].Path != tt.wantLast {
				t.Errorf("last = %s, want %s", got[len(got)-1].Path, tt.wantLast)
			}
		})
	}
}

func TestTopNFilesEmpty(t *testing.T) {
	got := TopNFiles(nil, 5)
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestTopNFilesDoesNotMutateInput(t *testing.T) {
	files := []internal.FileInfo{
		{Path: "/a.txt", Size: 100},
		{Path: "/b.txt", Size: 500},
	}
	TopNFiles(files, 1)
	if files[0].Path != "/a.txt" {
		t.Error("input slice was mutated")
	}
}
