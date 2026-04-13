package formatter

import (
	"bytes"
	"strings"
	"testing"

	"mogura/internal/analyzer"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name  string
		table Table
		want  string
	}{
		{
			name: "empty header",
			table: Table{
				Header: Row{},
			},
			want: "",
		},
		{
			name: "header only",
			table: Table{
				Header: Row{"Name", "Size"},
			},
			want: "Name  Size\n----------\n",
		},
		{
			name: "left aligned columns",
			table: Table{
				Header: Row{"Path", "Size"},
				Rows: []Row{
					{"foo", "1.0 KB"},
					{"barbaz", "2.0 MB"},
				},
			},
			want: "Path    Size  \n--------------\nfoo     1.0 KB\nbarbaz  2.0 MB\n",
		},
		{
			name: "right aligned size column",
			table: Table{
				Header:     Row{"Path", "Size"},
				Rows:       []Row{{"foo", "1.0 KB"}, {"barbaz", "2.0 MB"}},
				RightAlign: []bool{false, true},
			},
			want: "Path      Size\n--------------\nfoo     1.0 KB\nbarbaz  2.0 MB\n",
		},
		{
			name: "multiple right aligned columns",
			table: Table{
				Header:     Row{"Ext", "Size", "Count"},
				Rows:       []Row{{".go", "1.0 KB", "5"}, {".json", "200 B", "12"}},
				RightAlign: []bool{false, true, true},
			},
			want: "Ext      Size  Count\n--------------------\n.go    1.0 KB      5\n.json   200 B     12\n",
		},
		{
			name: "wide header narrow data",
			table: Table{
				Header:     Row{"Category", "Size"},
				Rows:       []Row{{"AB", "1 B"}},
				RightAlign: []bool{false, true},
			},
			want: "Category  Size\n--------------\nAB         1 B\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.table.Render()
			if got != tt.want {
				t.Errorf("Render() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestFormatDiffTable(t *testing.T) {
	tests := []struct {
		name  string
		diffs []analyzer.DirDiff
		limit int
		check func(t *testing.T, output string)
	}{
		{
			name:  "empty diffs",
			diffs: nil,
			limit: 0,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Path") {
					t.Error("expected header")
				}
			},
		},
		{
			name: "increase shows red plus",
			diffs: []analyzer.DirDiff{
				{Path: "/home/user", PrevSize: 1024, CurrSize: 2048, Delta: 1024},
			},
			limit: 0,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, colorRed) {
					t.Error("expected red color for increase")
				}
				if !strings.Contains(output, "+1.0 KB") {
					t.Error("expected +1.0 KB")
				}
			},
		},
		{
			name: "decrease shows green minus",
			diffs: []analyzer.DirDiff{
				{Path: "/tmp", PrevSize: 4096, CurrSize: 1024, Delta: -3072},
			},
			limit: 0,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, colorGreen) {
					t.Error("expected green color for decrease")
				}
				if !strings.Contains(output, "-3.0 KB") {
					t.Error("expected -3.0 KB")
				}
			},
		},
		{
			name: "zero delta no color",
			diffs: []analyzer.DirDiff{
				{Path: "/opt", PrevSize: 512, CurrSize: 512, Delta: 0},
			},
			limit: 0,
			check: func(t *testing.T, output string) {
				if strings.Contains(output, colorRed) || strings.Contains(output, colorGreen) {
					t.Error("expected no color for zero delta")
				}
				if !strings.Contains(output, "0 B") {
					t.Error("expected 0 B")
				}
			},
		},
		{
			name: "limit applied",
			diffs: []analyzer.DirDiff{
				{Path: "/a", PrevSize: 0, CurrSize: 100, Delta: 100},
				{Path: "/b", PrevSize: 0, CurrSize: 200, Delta: 200},
				{Path: "/c", PrevSize: 0, CurrSize: 300, Delta: 300},
			},
			limit: 2,
			check: func(t *testing.T, output string) {
				if strings.Contains(output, "/c") {
					t.Error("third entry should be excluded by limit")
				}
				if !strings.Contains(output, "/a") || !strings.Contains(output, "/b") {
					t.Error("first two entries should be present")
				}
			},
		},
		{
			name: "columns align despite ANSI codes",
			diffs: []analyzer.DirDiff{
				{Path: "/short", PrevSize: 1024, CurrSize: 2048, Delta: 1024},
				{Path: "/very/long/path", PrevSize: 1048576, CurrSize: 524288, Delta: -524288},
			},
			limit: 0,
			check: func(t *testing.T, output string) {
				lines := strings.Split(output, "\n")
				if len(lines) < 4 {
					t.Fatalf("expected at least 4 lines, got %d", len(lines))
				}
				headerWidth := displayWidth(lines[0])
				for i := 2; i < len(lines); i++ {
					if lines[i] == "" {
						continue
					}
					rowWidth := displayWidth(lines[i])
					if rowWidth != headerWidth {
						t.Errorf("row %d display width %d != header width %d", i, rowWidth, headerWidth)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			FormatDiffTable(tt.diffs, &buf, tt.limit)
			tt.check(t, buf.String())
		})
	}
}
