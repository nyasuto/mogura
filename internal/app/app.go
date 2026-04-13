package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mogura/internal"
	"mogura/internal/analyzer"
	"mogura/internal/formatter"
	"mogura/internal/scanner"
)

type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
	FormatTree OutputFormat = "tree"
	FormatHTML OutputFormat = "html"
)

type SizeMode string

const (
	SizeModeLogical  SizeMode = "logical"
	SizeModePhysical SizeMode = "physical"
)

type Config struct {
	TargetPath    string
	TopN          int
	Depth         int
	OutputFormat  OutputFormat
	Exclude       []string
	OlderThanDays int
	DiffPath      string
	MinSize       int64
	FilterExt     []string
	Quiet         bool
	OneFileSystem bool
	SizeMode      SizeMode
}

func ParseFlags(args []string) (Config, error) {
	fs := flag.NewFlagSet("mogura", flag.ContinueOnError)

	jsonFlag := fs.Bool("json", false, "JSON 形式で出力")
	treeFlag := fs.Bool("tree", false, "ツリー形式で出力")
	htmlFlag := fs.Bool("html", false, "HTML レポートを出力")
	depth := fs.Int("depth", 3, "ツリー表示の深さ")
	top := fs.Int("top", 20, "巨大ファイル表示件数")
	olderThan := fs.Int("older-than", 365, "古いファイルの判定日数")
	exclude := fs.String("exclude", "", "除外パターン（カンマ区切り: node_modules,.git,*.tmp）")
	diffPath := fs.String("diff", "", "前回の JSON レポートファイルと差分比較")
	minSizeStr := fs.String("min-size", "", "最小ファイルサイズ（例: 10M, 1G, 500K）")
	filterExt := fs.String("ext", "", "対象拡張子（カンマ区切り: mp4,mkv,avi）")
	quiet := fs.Bool("quiet", false, "進捗表示を抑制")
	oneFS := fs.Bool("x", false, "ファイルシステム境界を越えない")
	sizeMode := fs.String("size-mode", "logical", "サイズ表示モード（logical / physical）")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "usage: mogura [flags] <path>\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	if fs.NArg() < 1 {
		fs.Usage()
		return Config{}, fmt.Errorf("path argument is required")
	}

	format := FormatText
	switch {
	case *jsonFlag:
		format = FormatJSON
	case *treeFlag:
		format = FormatTree
	case *htmlFlag:
		format = FormatHTML
	}

	var excludes []string
	if *exclude != "" {
		for _, e := range strings.Split(*exclude, ",") {
			if t := strings.TrimSpace(e); t != "" {
				excludes = append(excludes, t)
			}
		}
	}

	var minSize int64
	if *minSizeStr != "" {
		var err error
		minSize, err = ParseHumanSize(*minSizeStr)
		if err != nil {
			return Config{}, fmt.Errorf("invalid -min-size value %q: %w", *minSizeStr, err)
		}
	}

	var filterExts []string
	if *filterExt != "" {
		for _, e := range strings.Split(*filterExt, ",") {
			if t := strings.TrimSpace(e); t != "" {
				filterExts = append(filterExts, t)
			}
		}
	}

	sm := SizeModeLogical
	if *sizeMode == "physical" {
		sm = SizeModePhysical
	}

	return Config{
		TargetPath:    fs.Arg(0),
		TopN:          *top,
		Depth:         *depth,
		OutputFormat:  format,
		Exclude:       excludes,
		OlderThanDays: *olderThan,
		DiffPath:      *diffPath,
		MinSize:       minSize,
		FilterExt:     filterExts,
		Quiet:         *quiet,
		OneFileSystem: *oneFS,
		SizeMode:      sm,
	}, nil
}

func ParseHumanSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size string")
	}

	s = strings.ToUpper(s)
	multiplier := int64(1)
	suffix := s[len(s)-1]
	switch suffix {
	case 'K':
		multiplier = 1024
		s = s[:len(s)-1]
	case 'M':
		multiplier = 1024 * 1024
		s = s[:len(s)-1]
	case 'G':
		multiplier = 1024 * 1024 * 1024
		s = s[:len(s)-1]
	case 'T':
		multiplier = 1024 * 1024 * 1024 * 1024
		s = s[:len(s)-1]
	default:
		if suffix < '0' || suffix > '9' {
			return 0, fmt.Errorf("unknown size suffix %q", string(suffix))
		}
	}

	s = strings.TrimRight(s, "bBiI")

	var val float64
	_, err := fmt.Sscanf(s, "%f", &val)
	if err != nil {
		return 0, fmt.Errorf("cannot parse number %q: %w", s, err)
	}
	if val < 0 {
		return 0, fmt.Errorf("size must be non-negative")
	}

	return int64(val * float64(multiplier)), nil
}

func shortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}

func FilterFiles(files []internal.FileInfo, minSize int64, filterExt []string) []internal.FileInfo {
	if minSize == 0 && len(filterExt) == 0 {
		return files
	}

	extSet := make(map[string]bool, len(filterExt))
	for _, e := range filterExt {
		ext := strings.ToLower(e)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		extSet[ext] = true
	}

	filtered := make([]internal.FileInfo, 0, len(files))
	for _, f := range files {
		if minSize > 0 && f.Size < minSize {
			continue
		}
		if len(extSet) > 0 && !extSet[strings.ToLower(filepath.Ext(f.Path))] {
			continue
		}
		filtered = append(filtered, f)
	}
	return filtered
}

func Run(cfg Config, stdout io.Writer, stderr io.Writer) error {
	scanOpts := scanner.ScanOpts{
		Exclude:       cfg.Exclude,
		OneFileSystem: cfg.OneFileSystem,
	}

	if !cfg.Quiet {
		lastUpdate := time.Time{}
		scanOpts.OnProgress = func(scanned int, currentDir string) {
			now := time.Now()
			if now.Sub(lastUpdate) < 500*time.Millisecond {
				return
			}
			lastUpdate = now
			dir := shortenPath(currentDir)
			fmt.Fprintf(stderr, "\rScanning... %d files (%s)", scanned, dir)
		}
	}

	files, err := scanner.Scan(cfg.TargetPath, scanOpts)
	if err != nil {
		return err
	}

	if !cfg.Quiet && scanOpts.OnProgress != nil {
		fmt.Fprintf(stderr, "\r\033[K")
	}

	files = FilterFiles(files, cfg.MinSize, cfg.FilterExt)

	now := time.Now()
	result := analyzer.Analyze(files, analyzer.AnalyzeOpts{
		TopN:          cfg.TopN,
		Depth:         cfg.Depth,
		OlderThanDays: cfg.OlderThanDays,
		Now:           now,
	})
	result.SizeMode = string(cfg.SizeMode)

	if cfg.DiffPath != "" {
		prev, err := analyzer.LoadPrevResult(cfg.DiffPath)
		if err != nil {
			return err
		}
		diffs := analyzer.ComputeDiff(prev, result)
		result.DiffSummary = diffs

		switch cfg.OutputFormat {
		case FormatJSON:
			return formatter.FormatJSON(result, stdout)
		case FormatHTML:
			return formatter.FormatHTML(result, stdout)
		default:
			fmt.Fprintln(stdout, "=== 差分 ===")
			formatter.FormatDiffTable(diffs, stdout, 20)
		}
		return nil
	}

	switch cfg.OutputFormat {
	case FormatJSON:
		return formatter.FormatJSON(result, stdout)
	case FormatTree:
		formatter.FormatTree(result, stdout)
	case FormatHTML:
		return formatter.FormatHTML(result, stdout)
	default:
		formatter.FormatTable(result, stdout)
	}

	return nil
}
