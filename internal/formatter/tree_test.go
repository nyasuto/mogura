package formatter

import (
	"strings"
	"testing"

	"mogura/internal/analyzer"
)

func TestRenderTree(t *testing.T) {
	tests := []struct {
		name     string
		node     analyzer.DirNode
		contains []string
	}{
		{
			name: "single node no children",
			node: analyzer.DirNode{
				Name: "root",
				Size: 1024,
			},
			contains: []string{"root 1.0 KB (100.0%)"},
		},
		{
			name: "one level children",
			node: analyzer.DirNode{
				Name: "project",
				Size: 2048,
				Children: []analyzer.DirNode{
					{Name: "src", Size: 1536},
					{Name: "docs", Size: 512},
				},
			},
			contains: []string{
				"project 2.0 KB (100.0%)",
				"├── src 1.5 KB (75.0%)",
				"└── docs 512 B (25.0%)",
			},
		},
		{
			name: "nested children",
			node: analyzer.DirNode{
				Name: "root",
				Size: 4096,
				Children: []analyzer.DirNode{
					{
						Name: "a",
						Size: 3072,
						Children: []analyzer.DirNode{
							{Name: "a1", Size: 2048},
							{Name: "a2", Size: 1024},
						},
					},
					{Name: "b", Size: 1024},
				},
			},
			contains: []string{
				"root 4.0 KB (100.0%)",
				"├── a 3.0 KB (75.0%)",
				"│   ├── a1 2.0 KB (50.0%)",
				"│   └── a2 1.0 KB (25.0%)",
				"└── b 1.0 KB (25.0%)",
			},
		},
		{
			name: "zero size root",
			node: analyzer.DirNode{
				Name: "empty",
				Size: 0,
			},
			contains: []string{"empty 0 B (100.0%)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderTree(tt.node)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("RenderTree() missing expected line %q\ngot:\n%s", want, got)
				}
			}
		})
	}
}

func TestRenderTreeOmitsSmallNodes(t *testing.T) {
	tests := []struct {
		name       string
		node       analyzer.DirNode
		contains   []string
		notContain []string
	}{
		{
			name: "omit children below 1 percent",
			node: analyzer.DirNode{
				Name: "root",
				Size: 10000,
				Children: []analyzer.DirNode{
					{Name: "big", Size: 9000},
					{Name: "medium", Size: 900},
					{Name: "tiny", Size: 99},
					{Name: "micro", Size: 1},
				},
			},
			contains: []string{
				"root 9.8 KB (100.0%)",
				"big 8.8 KB (90.0%)",
				"medium 900 B (9.0%)",
			},
			notContain: []string{
				"tiny",
				"micro",
			},
		},
		{
			name: "exactly 1 percent is shown",
			node: analyzer.DirNode{
				Name: "root",
				Size: 10000,
				Children: []analyzer.DirNode{
					{Name: "big", Size: 9900},
					{Name: "borderline", Size: 100},
				},
			},
			contains: []string{
				"big",
				"borderline",
			},
			notContain: []string{},
		},
		{
			name: "nested small nodes omitted",
			node: analyzer.DirNode{
				Name: "root",
				Size: 10000,
				Children: []analyzer.DirNode{
					{
						Name: "dir",
						Size: 10000,
						Children: []analyzer.DirNode{
							{Name: "large", Size: 9950},
							{Name: "small", Size: 50},
						},
					},
				},
			},
			contains:   []string{"large"},
			notContain: []string{"small"},
		},
		{
			name: "last connector adjusts after filtering",
			node: analyzer.DirNode{
				Name: "root",
				Size: 10000,
				Children: []analyzer.DirNode{
					{Name: "big", Size: 9900},
					{Name: "tiny", Size: 50},
					{Name: "also_tiny", Size: 50},
				},
			},
			contains: []string{
				"└── big",
			},
			notContain: []string{
				"├── big",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderTree(tt.node)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("RenderTree() missing expected %q\ngot:\n%s", want, got)
				}
			}
			for _, notWant := range tt.notContain {
				if strings.Contains(got, notWant) {
					t.Errorf("RenderTree() should not contain %q\ngot:\n%s", notWant, got)
				}
			}
		})
	}
}

func TestRenderTreeStructure(t *testing.T) {
	node := analyzer.DirNode{
		Name: "root",
		Size: 4096,
		Children: []analyzer.DirNode{
			{
				Name: "a",
				Size: 3072,
				Children: []analyzer.DirNode{
					{Name: "a1", Size: 2048},
					{Name: "a2", Size: 1024},
				},
			},
			{Name: "b", Size: 1024},
		},
	}

	got := RenderTree(node)
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")

	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d:\n%s", len(lines), got)
	}

	if !strings.HasPrefix(lines[1], "├") {
		t.Errorf("line 1 should start with ├, got: %s", lines[1])
	}
	if !strings.HasPrefix(lines[4], "└") {
		t.Errorf("last line should start with └, got: %s", lines[4])
	}
}
