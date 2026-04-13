package formatter

import "testing"

func TestRenderBar(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		maxValue int
		width    int
		want     string
	}{
		{
			name:     "full bar",
			value:    100,
			maxValue: 100,
			width:    10,
			want:     "██████████",
		},
		{
			name:     "empty bar",
			value:    0,
			maxValue: 100,
			width:    10,
			want:     "░░░░░░░░░░",
		},
		{
			name:     "half bar",
			value:    50,
			maxValue: 100,
			width:    10,
			want:     "█████░░░░░",
		},
		{
			name:     "quarter bar",
			value:    25,
			maxValue: 100,
			width:    8,
			want:     "██░░░░░░",
		},
		{
			name:     "value exceeds max",
			value:    200,
			maxValue: 100,
			width:    10,
			want:     "██████████",
		},
		{
			name:     "zero max value",
			value:    50,
			maxValue: 0,
			width:    10,
			want:     "░░░░░░░░░░",
		},
		{
			name:     "zero width",
			value:    50,
			maxValue: 100,
			width:    0,
			want:     "",
		},
		{
			name:     "one filled",
			value:    1,
			maxValue: 10,
			width:    10,
			want:     "█░░░░░░░░░",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderBar(tt.value, tt.maxValue, tt.width)
			if got != tt.want {
				t.Errorf("RenderBar(%d, %d, %d) = %q, want %q", tt.value, tt.maxValue, tt.width, got, tt.want)
			}
		})
	}
}
