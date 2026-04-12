package analyzer

type DirNode struct {
	Name      string
	Size      int64
	Children  []DirNode
	FileCount int
}
