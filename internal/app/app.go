package app

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
