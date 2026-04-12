package analyzer

type WasteDir struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
	Kind string `json:"kind"`
}

var wastePatterns = map[string]string{
	"node_modules":  "node_modules",
	".cache":        "cache",
	"__pycache__":   "cache",
	"DerivedData":   "build",
	".Trash":        "cache",
	"Caches":        "cache",
	".gradle":       "build",
	".cargo/registry": "build",
	".npm":          "cache",
	"target":        "build",
}
