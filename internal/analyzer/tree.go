package analyzer

import (
	"path/filepath"
	"sort"
	"strings"

	"mogura/internal"
)

type DirNode struct {
	Name             string    `json:"name"`
	Size             int64     `json:"size"`
	Children         []DirNode `json:"children,omitempty"`
	FileCount        int       `json:"file_count"`
	DominantCategory string    `json:"dominant_category"`
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
		catSizes  map[Category]int64
	}

	dirs := make(map[string]*dirInfo)

	ensureDir := func(path string) *dirInfo {
		if d, ok := dirs[path]; ok {
			return d
		}
		d := &dirInfo{children: make(map[string]bool), catSizes: make(map[Category]int64)}
		dirs[path] = d
		return d
	}

	ensureDir(root)

	for _, f := range files {
		d := ensureDir(f.Dir)
		d.size += f.Size
		d.fileCount++
		cat := ClassifyExt(f.Ext)
		d.catSizes[cat] += f.Size

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
			for cat, sz := range dirs[child].catSizes {
				d.catSizes[cat] += sz
			}
		}

		var maxCat Category
		var maxSize int64
		for cat, sz := range d.catSizes {
			if sz > maxSize {
				maxSize = sz
				maxCat = cat
			}
		}
		node.DominantCategory = string(maxCat)

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
			Name:             node.Name,
			Size:             node.Size,
			FileCount:        node.FileCount,
			DominantCategory: node.DominantCategory,
		}
	}
	pruned := DirNode{
		Name:             node.Name,
		Size:             node.Size,
		FileCount:        node.FileCount,
		DominantCategory: node.DominantCategory,
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
