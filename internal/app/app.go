package app

import (
	"flag"
	"fmt"
	"io"
)

type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
	FormatTree OutputFormat = "tree"
	FormatHTML OutputFormat = "html"
)

type Config struct {
	TargetPath    string
	TopN          int
	Depth         int
	OutputFormat  OutputFormat
	Exclude       []string
	OlderThanDays int
}

func ParseFlags(args []string) (Config, error) {
	fs := flag.NewFlagSet("mogura", flag.ContinueOnError)

	jsonFlag := fs.Bool("json", false, "JSON 形式で出力")
	treeFlag := fs.Bool("tree", false, "ツリー形式で出力")
	depth := fs.Int("depth", 3, "ツリー表示の深さ")
	top := fs.Int("top", 20, "巨大ファイル表示件数")
	olderThan := fs.Int("older-than", 365, "古いファイルの判定日数")

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
	}

	return Config{
		TargetPath:    fs.Arg(0),
		TopN:          *top,
		Depth:         *depth,
		OutputFormat:  format,
		OlderThanDays: *olderThan,
	}, nil
}

func Run(cfg Config, stdout io.Writer, stderr io.Writer) error {
	return nil
}
