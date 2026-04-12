package analyzer

import (
	"path/filepath"
	"sort"
	"strings"

	"mogura/internal"
)

type DirNode struct {
	Name      string
	Size      int64
	Children  []DirNode
	FileCount int
}

func BuildTree(files []internal.FileInfo) DirNode {
	if len(files) == 0 {
		return DirNode{}
	}

	root := commonRoot(files)

	type dirInfo struct {
		size      int64
		fileCount int
		children  map[string]bool
	}

	dirs := make(map[string]*dirInfo)

	ensureDir := func(path string) *dirInfo {
		if d, ok := dirs[path]; ok {
			return d
		}
		d := &dirInfo{children: make(map[string]bool)}
		dirs[path] = d
		return d
	}

	ensureDir(root)

	for _, f := range files {
		d := ensureDir(f.Dir)
		d.size += f.Size
		d.fileCount++

		current := f.Dir
		for current != root {
			parent := filepath.Dir(current)
			pd := ensureDir(parent)
			pd.children[current] = true
			current = parent
		}
	}

	var build func(path string) DirNode
	build = func(path string) DirNode {
		name := filepath.Base(path)
		if path == root {
			name = root
		}

		d := dirs[path]
		node := DirNode{
			Name:      name,
			Size:      d.size,
			FileCount: d.fileCount,
		}

		for child := range d.children {
			childNode := build(child)
			node.Size += childNode.Size
			node.FileCount += childNode.FileCount
			node.Children = append(node.Children, childNode)
		}

		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Size > node.Children[j].Size
		})

		return node
	}

	return build(root)
}

func Prune(node DirNode, depth int) DirNode {
	if depth <= 0 {
		return DirNode{
			Name:      node.Name,
			Size:      node.Size,
			FileCount: node.FileCount,
		}
	}
	pruned := DirNode{
		Name:      node.Name,
		Size:      node.Size,
		FileCount: node.FileCount,
	}
	for _, child := range node.Children {
		pruned.Children = append(pruned.Children, Prune(child, depth-1))
	}
	return pruned
}

func commonRoot(files []internal.FileInfo) string {
	root := files[0].Dir
	for _, f := range files[1:] {
		for !strings.HasPrefix(f.Dir+string(filepath.Separator), root+string(filepath.Separator)) && f.Dir != root {
			root = filepath.Dir(root)
		}
	}
	return root
}
