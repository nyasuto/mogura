package analyzer

type DirDiff struct {
	Path     string `json:"path"`
	PrevSize int64  `json:"prev_size"`
	CurrSize int64  `json:"curr_size"`
	Delta    int64  `json:"delta"`
}
