package formatter

import (
	"fmt"
	"io"
	"strings"

	"mogura/internal"
	"mogura/internal/analyzer"
)

func FormatTree(result analyzer.Result, w io.Writer) {
	fmt.Fprint(w, RenderTree(result.DirTree))
}

func RenderTree(node analyzer.DirNode) string {
	var b strings.Builder
	rootSize := node.Size
	b.WriteString(fmt.Sprintf("%s %s (100.0%%)\n", node.Name, internal.FormatSize(node.Size)))
	renderChildren(&b, node.Children, "", rootSize)
	return b.String()
}

func renderChildren(b *strings.Builder, children []analyzer.DirNode, prefix string, rootSize int64) {
	visible := filterVisible(children, rootSize)
	for i, child := range visible {
		isLast := i == len(visible)-1

		connector := "├── "
		if isLast {
			connector = "└── "
		}

		percent := 0.0
		if rootSize > 0 {
			percent = float64(child.Size) / float64(rootSize) * 100
		}

		b.WriteString(fmt.Sprintf("%s%s%s %s (%.1f%%)\n", prefix, connector, child.Name, internal.FormatSize(child.Size), percent))

		if len(child.Children) > 0 {
			childPrefix := prefix + "│   "
			if isLast {
				childPrefix = prefix + "    "
			}
			renderChildren(b, child.Children, childPrefix, rootSize)
		}
	}
}

func filterVisible(children []analyzer.DirNode, rootSize int64) []analyzer.DirNode {
	if rootSize <= 0 {
		return children
	}
	var result []analyzer.DirNode
	for _, child := range children {
		percent := float64(child.Size) / float64(rootSize) * 100
		if percent >= 1.0 {
			result = append(result, child)
		}
	}
	return result
}
