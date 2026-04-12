package formatter

import (
	"strings"
	"testing"

	"mogura/internal/analyzer"
)

func TestRenderSummary(t *testing.T) {
	tests := []struct {
		name     string
		input    SummaryInput
		contains []string
	}{
		{
			name: "全フィールドが表示される",
			input: SummaryInput{
				TotalSize: 10 * 1024 * 1024 * 1024,
				Categories: map[analyzer.Category]analyzer.CategoryStats{
					analyzer.CategoryVideo:    {Size: 5 * 1024 * 1024 * 1024, Count: 10, Percent: 50.0},
					analyzer.CategoryImage:    {Size: 2 * 1024 * 1024 * 1024, Count: 100, Percent: 20.0},
					analyzer.CategoryCode:     {Size: 1 * 1024 * 1024 * 1024, Count: 500, Percent: 10.0},
					analyzer.CategoryDocument: {Size: 512 * 1024 * 1024, Count: 50, Percent: 5.0},
					analyzer.CategoryCache:    {Size: 256 * 1024 * 1024, Count: 200, Percent: 2.5},
					analyzer.CategoryArchive:  {Size: 128 * 1024 * 1024, Count: 5, Percent: 1.25},
				},
				WasteDirs: []analyzer.WasteDir{
					{Path: "/project/node_modules", Size: 300 * 1024 * 1024, Kind: "node_modules"},
					{Path: "/project/.cache", Size: 100 * 1024 * 1024, Kind: "cache"},
				},
				Stale: analyzer.StaleResult{
					TotalSize:  500 * 1024 * 1024,
					TotalFiles: 42,
				},
			},
			contains: []string{
				"総容量: 10.0 GB",
				"カテゴリ内訳:",
				"動画",
				"画像",
				"コード",
				"ドキュメント",
				"キャッシュ",
				"キャッシュ/ゴミ合計: 400.0 MB",
				"古いファイル合計: 500.0 MB (42 件)",
				"推定節約可能量: 900.0 MB",
			},
		},
		{
			name: "カテゴリが5件未満でも動作する",
			input: SummaryInput{
				TotalSize: 1024,
				Categories: map[analyzer.Category]analyzer.CategoryStats{
					analyzer.CategoryCode: {Size: 1024, Count: 1, Percent: 100.0},
				},
				Stale: analyzer.StaleResult{},
			},
			contains: []string{
				"総容量: 1.0 KB",
				"コード",
				"推定節約可能量: 0 B",
			},
		},
		{
			name: "カテゴリ上位5件のみ表示される",
			input: SummaryInput{
				TotalSize: 7 * 1024,
				Categories: map[analyzer.Category]analyzer.CategoryStats{
					analyzer.CategoryVideo:    {Size: 7 * 1024, Count: 1, Percent: 30.0},
					analyzer.CategoryImage:    {Size: 6 * 1024, Count: 1, Percent: 25.0},
					analyzer.CategoryCode:     {Size: 5 * 1024, Count: 1, Percent: 20.0},
					analyzer.CategoryDocument: {Size: 4 * 1024, Count: 1, Percent: 10.0},
					analyzer.CategoryCache:    {Size: 3 * 1024, Count: 1, Percent: 8.0},
					analyzer.CategoryArchive:  {Size: 2 * 1024, Count: 1, Percent: 5.0},
					analyzer.CategoryOther:    {Size: 1 * 1024, Count: 1, Percent: 2.0},
				},
				Stale: analyzer.StaleResult{},
			},
			contains: []string{
				"動画", "画像", "コード", "ドキュメント", "キャッシュ",
			},
		},
		{
			name:  "空の入力",
			input: SummaryInput{},
			contains: []string{
				"総容量: 0 B",
				"推定節約可能量: 0 B",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderSummary(tt.input)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("RenderSummary() missing %q\ngot:\n%s", want, got)
				}
			}
		})
	}
}

func TestRenderSummaryTopCategoriesLimit(t *testing.T) {
	input := SummaryInput{
		TotalSize: 7 * 1024,
		Categories: map[analyzer.Category]analyzer.CategoryStats{
			analyzer.CategoryVideo:    {Size: 7 * 1024, Count: 1, Percent: 30.0},
			analyzer.CategoryImage:    {Size: 6 * 1024, Count: 1, Percent: 25.0},
			analyzer.CategoryCode:     {Size: 5 * 1024, Count: 1, Percent: 20.0},
			analyzer.CategoryDocument: {Size: 4 * 1024, Count: 1, Percent: 10.0},
			analyzer.CategoryCache:    {Size: 3 * 1024, Count: 1, Percent: 8.0},
			analyzer.CategoryArchive:  {Size: 2 * 1024, Count: 1, Percent: 5.0},
			analyzer.CategoryOther:    {Size: 1 * 1024, Count: 1, Percent: 2.0},
		},
		Stale: analyzer.StaleResult{},
	}

	got := RenderSummary(input)
	if strings.Contains(got, "その他") {
		t.Errorf("7位のその他が表示されている\ngot:\n%s", got)
	}
	if strings.Contains(got, "アーカイブ") {
		t.Errorf("6位のアーカイブが表示されている\ngot:\n%s", got)
	}
}
