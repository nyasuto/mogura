package formatter

import "strings"

func RenderBar(value, maxValue, width int) string {
	if width <= 0 || maxValue <= 0 {
		return strings.Repeat("░", width)
	}
	filled := value * width / maxValue
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}
