package formatter

import (
	"fmt"
	"io"
	"strings"

	"mogura/internal"
	"mogura/internal/analyzer"
)

func FormatTree(result analyzer.Result, w io.Writer) {
	usePhysical := result.SizeMode == "physical"
	fmt.Fprint(w, RenderTree(result.DirTree, usePhysical))
}

func nodeSize(node analyzer.DirNode, usePhysical bool) int64 {
	if usePhysical {
		return node.PhysicalSize
	}
	return node.Size
}

func RenderTree(node analyzer.DirNode, usePhysical bool) string {
	var b strings.Builder
	rootSize := nodeSize(node, usePhysical)
	b.WriteString(fmt.Sprintf("%s %s (100.0%%)\n", node.Name, internal.FormatSize(rootSize)))
	renderChildren(&b, node.Children, "", rootSize, usePhysical)
	return b.String()
}

func renderChildren(b *strings.Builder, children []analyzer.DirNode, prefix string, rootSize int64, usePhysical bool) {
	visible := filterVisible(children, rootSize, usePhysical)
	for i, child := range visible {
		isLast := i == len(visible)-1

		connector := "├── "
		if isLast {
			connector = "└── "
		}

		childSize := nodeSize(child, usePhysical)
		percent := 0.0
		if rootSize > 0 {
			percent = float64(childSize) / float64(rootSize) * 100
		}

		b.WriteString(fmt.Sprintf("%s%s%s %s (%.1f%%)\n", prefix, connector, child.Name, internal.FormatSize(childSize), percent))

		if len(child.Children) > 0 {
			childPrefix := prefix + "│   "
			if isLast {
				childPrefix = prefix + "    "
			}
			renderChildren(b, child.Children, childPrefix, rootSize, usePhysical)
		}
	}
}

func filterVisible(children []analyzer.DirNode, rootSize int64, usePhysical bool) []analyzer.DirNode {
	if rootSize <= 0 {
		return children
	}
	var result []analyzer.DirNode
	for _, child := range children {
		childSize := nodeSize(child, usePhysical)
		percent := float64(childSize) / float64(rootSize) * 100
		if percent >= 1.0 {
			result = append(result, child)
		}
	}
	return result
}
