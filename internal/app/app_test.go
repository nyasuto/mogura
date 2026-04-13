package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mogura/internal"
)

func setupTmpDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	sub := filepath.Join(dir, "subdir")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		filepath.Join(dir, "hello.txt"): "hello world",
		filepath.Join(sub, "data.csv"):  "a,b,c\n1,2,3\n",
		filepath.Join(sub, "image.png"): "fake png data for testing",
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestRun_TextOutput(t *testing.T) {
	dir := setupTmpDir(t)
	cfg := Config{
		TargetPath:    dir,
		TopN:          5,
		Depth:         3,
		OutputFormat:  FormatText,
		OlderThanDays: 365,
	}

	var stdout, stderr bytes.Buffer
	if err := Run(cfg, &stdout, &stderr); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if stdout.Len() == 0 {
		t.Error("expected non-empty stdout output")
	}
}

func TestRun_JSONOutput(t *testing.T) {
	dir := setupTmpDir(t)
	cfg := Config{
		TargetPath:    dir,
		TopN:          5,
		Depth:         3,
		OutputFormat:  FormatJSON,
		OlderThanDays: 365,
	}

	var stdout, stderr bytes.Buffer
	if err := Run(cfg, &stdout, &stderr); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if stdout.Len() == 0 {
		t.Error("expected non-empty stdout output")
	}
}

func TestRun_TreeOutput(t *testing.T) {
	dir := setupTmpDir(t)
	cfg := Config{
		TargetPath:    dir,
		TopN:          5,
		Depth:         3,
		OutputFormat:  FormatTree,
		OlderThanDays: 365,
	}

	var stdout, stderr bytes.Buffer
	if err := Run(cfg, &stdout, &stderr); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if stdout.Len() == 0 {
		t.Error("expected non-empty stdout output")
	}
}

func TestRun_WithExclude(t *testing.T) {
	dir := setupTmpDir(t)
	cfg := Config{
		TargetPath:    dir,
		TopN:          5,
		Depth:         3,
		OutputFormat:  FormatText,
		Exclude:       []string{"subdir"},
		OlderThanDays: 365,
	}

	var stdout, stderr bytes.Buffer
	if err := Run(cfg, &stdout, &stderr); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	output := stdout.String()
	if strings.Contains(output, "data.csv") || strings.Contains(output, "image.png") {
		t.Error("expected subdir files to be excluded from output")
	}
}

func TestRun_InvalidPath(t *testing.T) {
	cfg := Config{
		TargetPath:    "/nonexistent/path/that/does/not/exist",
		TopN:          5,
		Depth:         3,
		OutputFormat:  FormatText,
		OlderThanDays: 365,
	}

	var stdout, stderr bytes.Buffer
	err := Run(cfg, &stdout, &stderr)
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestFilterFiles(t *testing.T) {
	now := time.Now()
	files := []internal.FileInfo{
		{Path: "/a/small.txt", Size: 100, Ext: ".txt", ModTime: now},
		{Path: "/a/big.txt", Size: 10000, Ext: ".txt", ModTime: now},
		{Path: "/a/photo.jpg", Size: 5000, Ext: ".jpg", ModTime: now},
		{Path: "/a/video.mp4", Size: 20000, Ext: ".mp4", ModTime: now},
		{Path: "/a/noext", Size: 300, Ext: "", ModTime: now},
	}

	tests := []struct {
		name      string
		minSize   int64
		filterExt []string
		wantCount int
		wantPaths []string
	}{
		{
			name:      "no filter",
			minSize:   0,
			filterExt: nil,
			wantCount: 5,
		},
		{
			name:      "min size only",
			minSize:   5000,
			filterExt: nil,
			wantCount: 3,
			wantPaths: []string{"/a/big.txt", "/a/photo.jpg", "/a/video.mp4"},
		},
		{
			name:      "ext filter only",
			minSize:   0,
			filterExt: []string{"txt"},
			wantCount: 2,
			wantPaths: []string{"/a/small.txt", "/a/big.txt"},
		},
		{
			name:      "ext filter with dot prefix",
			minSize:   0,
			filterExt: []string{".jpg", ".mp4"},
			wantCount: 2,
			wantPaths: []string{"/a/photo.jpg", "/a/video.mp4"},
		},
		{
			name:      "both filters",
			minSize:   10000,
			filterExt: []string{"txt", "mp4"},
			wantCount: 2,
			wantPaths: []string{"/a/big.txt", "/a/video.mp4"},
		},
		{
			name:      "no matches",
			minSize:   100000,
			filterExt: nil,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterFiles(files, tt.minSize, tt.filterExt)
			if len(got) != tt.wantCount {
				t.Errorf("got %d files, want %d", len(got), tt.wantCount)
			}
			if tt.wantPaths != nil {
				for i, want := range tt.wantPaths {
					if i >= len(got) {
						break
					}
					if got[i].Path != want {
						t.Errorf("got[%d].Path = %q, want %q", i, got[i].Path, want)
					}
				}
			}
		})
	}
}

func TestParseHumanSize(t *testing.T) {
	tests := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{"100", 100, false},
		{"10K", 10 * 1024, false},
		{"10k", 10 * 1024, false},
		{"5M", 5 * 1024 * 1024, false},
		{"1G", 1024 * 1024 * 1024, false},
		{"2T", 2 * 1024 * 1024 * 1024 * 1024, false},
		{"1.5M", int64(1.5 * 1024 * 1024), false},
		{"", 0, true},
		{"abc", 0, true},
		{"-5M", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseHumanSize(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ParseHumanSize(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    Config
		wantErr bool
	}{
		{
			name: "basic path only",
			args: []string{"/tmp"},
			want: Config{
				TargetPath:    "/tmp",
				TopN:          20,
				Depth:         3,
				OutputFormat:  FormatText,
				OlderThanDays: 365,
			},
		},
		{
			name: "json flag",
			args: []string{"--json", "/tmp"},
			want: Config{
				TargetPath:    "/tmp",
				TopN:          20,
				Depth:         3,
				OutputFormat:  FormatJSON,
				OlderThanDays: 365,
			},
		},
		{
			name: "tree flag with depth",
			args: []string{"--tree", "--depth", "5", "/tmp"},
			want: Config{
				TargetPath:    "/tmp",
				TopN:          20,
				Depth:         5,
				OutputFormat:  FormatTree,
				OlderThanDays: 365,
			},
		},
		{
			name: "exclude flag",
			args: []string{"--exclude", "node_modules,.git,*.tmp", "/tmp"},
			want: Config{
				TargetPath:    "/tmp",
				TopN:          20,
				Depth:         3,
				OutputFormat:  FormatText,
				Exclude:       []string{"node_modules", ".git", "*.tmp"},
				OlderThanDays: 365,
			},
		},
		{
			name: "min-size and ext flags",
			args: []string{"--min-size", "10M", "--ext", "mp4,mkv", "/tmp"},
			want: Config{
				TargetPath:    "/tmp",
				TopN:          20,
				Depth:         3,
				OutputFormat:  FormatText,
				OlderThanDays: 365,
				MinSize:       10 * 1024 * 1024,
				FilterExt:     []string{"mp4", "mkv"},
			},
		},
		{
			name:    "invalid min-size",
			args:    []string{"--min-size", "abc", "/tmp"},
			wantErr: true,
		},
		{
			name:    "no path",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFlags(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.TargetPath != tt.want.TargetPath {
				t.Errorf("TargetPath = %q, want %q", got.TargetPath, tt.want.TargetPath)
			}
			if got.TopN != tt.want.TopN {
				t.Errorf("TopN = %d, want %d", got.TopN, tt.want.TopN)
			}
			if got.Depth != tt.want.Depth {
				t.Errorf("Depth = %d, want %d", got.Depth, tt.want.Depth)
			}
			if got.OutputFormat != tt.want.OutputFormat {
				t.Errorf("OutputFormat = %q, want %q", got.OutputFormat, tt.want.OutputFormat)
			}
			if got.OlderThanDays != tt.want.OlderThanDays {
				t.Errorf("OlderThanDays = %d, want %d", got.OlderThanDays, tt.want.OlderThanDays)
			}
			if got.MinSize != tt.want.MinSize {
				t.Errorf("MinSize = %d, want %d", got.MinSize, tt.want.MinSize)
			}
			if len(got.FilterExt) != len(tt.want.FilterExt) {
				t.Errorf("FilterExt len = %d, want %d", len(got.FilterExt), len(tt.want.FilterExt))
			} else {
				for i := range got.FilterExt {
					if got.FilterExt[i] != tt.want.FilterExt[i] {
						t.Errorf("FilterExt[%d] = %q, want %q", i, got.FilterExt[i], tt.want.FilterExt[i])
					}
				}
			}
			if len(got.Exclude) != len(tt.want.Exclude) {
				t.Errorf("Exclude len = %d, want %d", len(got.Exclude), len(tt.want.Exclude))
			} else {
				for i := range got.Exclude {
					if got.Exclude[i] != tt.want.Exclude[i] {
						t.Errorf("Exclude[%d] = %q, want %q", i, got.Exclude[i], tt.want.Exclude[i])
					}
				}
			}
		})
	}
}
