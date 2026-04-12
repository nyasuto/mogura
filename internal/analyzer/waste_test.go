package analyzer

import (
	"testing"

	"mogura/internal"
)

func TestDetectWaste(t *testing.T) {
	tests := []struct {
		name  string
		files []internal.FileInfo
		want  []WasteDir
	}{
		{
			name:  "empty input",
			files: nil,
			want:  []WasteDir{},
		},
		{
			name: "no waste directories",
			files: []internal.FileInfo{
				{Path: "/home/user/src/main.go", Size: 1000, Dir: "/home/user/src"},
			},
			want: []WasteDir{},
		},
		{
			name: "node_modules detected",
			files: []internal.FileInfo{
				{Path: "/project/node_modules/express/index.js", Size: 500, Dir: "/project/node_modules/express"},
				{Path: "/project/node_modules/lodash/lodash.js", Size: 300, Dir: "/project/node_modules/lodash"},
			},
			want: []WasteDir{
				{Path: "/project/node_modules", Size: 800, Kind: "node_modules"},
			},
		},
		{
			name: "multiple waste directories sorted by size",
			files: []internal.FileInfo{
				{Path: "/project/node_modules/a.js", Size: 100, Dir: "/project/node_modules"},
				{Path: "/project/.cache/b.dat", Size: 500, Dir: "/project/.cache"},
				{Path: "/project/__pycache__/c.pyc", Size: 200, Dir: "/project/__pycache__"},
			},
			want: []WasteDir{
				{Path: "/project/.cache", Size: 500, Kind: "cache"},
				{Path: "/project/__pycache__", Size: 200, Kind: "cache"},
				{Path: "/project/node_modules", Size: 100, Kind: "node_modules"},
			},
		},
		{
			name: "nested waste uses shallowest match",
			files: []internal.FileInfo{
				{Path: "/a/node_modules/b/node_modules/c.js", Size: 100, Dir: "/a/node_modules/b/node_modules"},
			},
			want: []WasteDir{
				{Path: "/a/node_modules", Size: 100, Kind: "node_modules"},
			},
		},
		{
			name: "cargo registry multi-component pattern",
			files: []internal.FileInfo{
				{Path: "/home/user/.cargo/registry/cache/pkg.crate", Size: 2000, Dir: "/home/user/.cargo/registry/cache"},
			},
			want: []WasteDir{
				{Path: "/home/user/.cargo/registry", Size: 2000, Kind: "build"},
			},
		},
		{
			name: "multiple separate node_modules",
			files: []internal.FileInfo{
				{Path: "/a/node_modules/x.js", Size: 300, Dir: "/a/node_modules"},
				{Path: "/b/node_modules/y.js", Size: 700, Dir: "/b/node_modules"},
			},
			want: []WasteDir{
				{Path: "/b/node_modules", Size: 700, Kind: "node_modules"},
				{Path: "/a/node_modules", Size: 300, Kind: "node_modules"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectWaste(tt.files)
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d waste dirs, want %d: %v", len(got), len(tt.want), got)
			}
			for i, w := range tt.want {
				if got[i].Path != w.Path {
					t.Errorf("[%d] path: got %q, want %q", i, got[i].Path, w.Path)
				}
				if got[i].Size != w.Size {
					t.Errorf("[%d] size: got %d, want %d", i, got[i].Size, w.Size)
				}
				if got[i].Kind != w.Kind {
					t.Errorf("[%d] kind: got %q, want %q", i, got[i].Kind, w.Kind)
				}
			}
		})
	}
}
