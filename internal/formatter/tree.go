package formatter

import (
	"fmt"
	"strings"

	"mogura/internal"
	"mogura/internal/analyzer"
)

func RenderTree(node analyzer.DirNode) string {
	var b strings.Builder
	rootSize := node.Size
	b.WriteString(fmt.Sprintf("%s %s (100.0%%)\n", node.Name, internal.FormatSize(node.Size)))
	renderChildren(&b, node.Children, "", rootSize)
	return b.String()
}

func renderChildren(b *strings.Builder, children []analyzer.DirNode, prefix string, rootSize int64) {
	for i, child := range children {
		isLast := i == len(children)-1

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
