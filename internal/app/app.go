package app

import "io"

type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
	FormatTree OutputFormat = "tree"
	FormatHTML OutputFormat = "html"
)

type Config struct {
	TargetPath   string
	TopN         int
	Depth        int
	OutputFormat OutputFormat
	Exclude      []string
	OlderThanDays int
}

func Run(cfg Config, stdout io.Writer, stderr io.Writer) error {
	return nil
}
