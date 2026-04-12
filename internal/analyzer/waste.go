package analyzer

import (
	"mogura/internal"
	"path/filepath"
	"sort"
	"strings"
)

type WasteDir struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
	Kind string `json:"kind"`
}

var wastePatterns = map[string]string{
	"node_modules":    "node_modules",
	".cache":          "cache",
	"__pycache__":     "cache",
	"DerivedData":     "build",
	".Trash":          "cache",
	"Caches":          "cache",
	".gradle":         "build",
	".cargo/registry": "build",
	".npm":            "cache",
	"target":          "build",
}

func findWasteMatch(dir string) (wastePath string, kind string, found bool) {
	sep := string(filepath.Separator)
	normalized := dir + sep

	for pattern, k := range wastePatterns {
		target := sep + pattern + sep
		idx := strings.Index(normalized, target)
		if idx < 0 {
			continue
		}
		matched := dir[:idx+len(sep)+len(pattern)]
		if !found || len(matched) < len(wastePath) {
			wastePath = matched
			kind = k
			found = true
		}
	}
	return
}

const DefaultGitSizeThreshold int64 = 100 * 1024 * 1024 // 100MB

func DetectLargeGitDirs(files []internal.FileInfo, threshold int64) []WasteDir {
	gitSizes := make(map[string]int64)

	sep := string(filepath.Separator)
	target := sep + ".git" + sep

	for _, f := range files {
		normalized := f.Dir + sep
		idx := strings.Index(normalized, target)
		if idx < 0 {
			continue
		}
		gitPath := f.Dir[:idx+len(target)-1]
		gitSizes[gitPath] += f.Size
	}

	var result []WasteDir
	for path, size := range gitSizes {
		if size >= threshold {
			result = append(result, WasteDir{
				Path: path,
				Size: size,
				Kind: "git",
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Size > result[j].Size
	})

	return result
}

func DetectWaste(files []internal.FileInfo) []WasteDir {
	wasteSizes := make(map[string]int64)
	wasteKinds := make(map[string]string)

	for _, f := range files {
		if wastePath, kind, ok := findWasteMatch(f.Dir); ok {
			wasteSizes[wastePath] += f.Size
			wasteKinds[wastePath] = kind
		}
	}

	result := make([]WasteDir, 0, len(wasteSizes))
	for path, size := range wasteSizes {
		result = append(result, WasteDir{
			Path: path,
			Size: size,
			Kind: wasteKinds[path],
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Size > result[j].Size
	})

	return result
}
