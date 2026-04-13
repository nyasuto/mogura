package formatter

import (
	"embed"
	"encoding/json"
	"html/template"
	"io"

	"mogura/internal/analyzer"
)

//go:embed templates/report.html
var templateFS embed.FS

var reportTemplate = template.Must(template.ParseFS(templateFS, "templates/report.html"))

type htmlData struct {
	DataJSON template.JS
}

func FormatHTML(result analyzer.Result, w io.Writer) error {
	report := buildReport(result)
	b, err := json.Marshal(report)
	if err != nil {
		return err
	}

	return reportTemplate.Execute(w, htmlData{
		DataJSON: template.JS(b),
	})
}
