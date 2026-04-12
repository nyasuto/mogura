package formatter

import (
	"testing"
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
