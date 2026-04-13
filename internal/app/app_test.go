package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func setupTmpDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	sub := filepath.Join(dir, "subdir")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		filepath.Join(dir, "hello.txt"):    "hello world",
		filepath.Join(sub, "data.csv"):     "a,b,c\n1,2,3\n",
		filepath.Join(sub, "image.png"):    "fake png data for testing",
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
		})
	}
}
