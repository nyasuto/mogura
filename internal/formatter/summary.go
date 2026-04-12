package formatter

import (
	"fmt"
	"sort"
	"strings"

	"mogura/internal"
	"mogura/internal/analyzer"
)

type SummaryInput struct {
	TotalSize  int64
	Categories map[analyzer.Category]analyzer.CategoryStats
	WasteDirs  []analyzer.WasteDir
	Stale      analyzer.StaleResult
}

func RenderSummary(input SummaryInput) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("総容量: %s\n", internal.FormatSize(input.TotalSize)))
	b.WriteByte('\n')

	renderTopCategories(&b, input.Categories)
	renderWasteSummary(&b, input.WasteDirs)
	renderStaleSummary(&b, input.Stale)
	renderSavings(&b, input)

	return b.String()
}

func renderTopCategories(b *strings.Builder, cats map[analyzer.Category]analyzer.CategoryStats) {
	if len(cats) == 0 {
		return
	}

	type entry struct {
		cat  analyzer.Category
		stat analyzer.CategoryStats
	}
	entries := make([]entry, 0, len(cats))
	for cat, stat := range cats {
		entries = append(entries, entry{cat, stat})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].stat.Size > entries[j].stat.Size
	})

	limit := 5
	if len(entries) < limit {
		limit = len(entries)
	}

	b.WriteString("カテゴリ内訳:\n")
	for _, e := range entries[:limit] {
		b.WriteString(fmt.Sprintf("  %-12s %10s  %5.1f%%\n",
			string(e.cat), internal.FormatSize(e.stat.Size), e.stat.Percent))
	}
	b.WriteByte('\n')
}

func renderWasteSummary(b *strings.Builder, wasteDirs []analyzer.WasteDir) {
	var total int64
	for _, w := range wasteDirs {
		total += w.Size
	}
	b.WriteString(fmt.Sprintf("キャッシュ/ゴミ合計: %s\n", internal.FormatSize(total)))
}

func renderStaleSummary(b *strings.Builder, stale analyzer.StaleResult) {
	b.WriteString(fmt.Sprintf("古いファイル合計: %s (%d 件)\n",
		internal.FormatSize(stale.TotalSize), stale.TotalFiles))
}

func renderSavings(b *strings.Builder, input SummaryInput) {
	var wasteTotal int64
	for _, w := range input.WasteDirs {
		wasteTotal += w.Size
	}
	savings := wasteTotal + input.Stale.TotalSize

	b.WriteString(fmt.Sprintf("\n推定節約可能量: %s\n", internal.FormatSize(savings)))
}
