package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScan(t *testing.T) {
	base := t.TempDir()

	// Create directory structure
	sub := filepath.Join(base, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path    string
		content string
	}{
		{filepath.Join(base, "a.txt"), "hello"},
		{filepath.Join(sub, "b.go"), "package main"},
	}

	for _, tt := range tests {
		if err := os.WriteFile(tt.path, []byte(tt.content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	files, err := Scan(base)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	found := map[string]bool{}
	for _, f := range files {
		found[filepath.Base(f.Path)] = true

		if f.Size <= 0 {
			t.Errorf("expected positive size for %s, got %d", f.Path, f.Size)
		}
	}

	for _, name := range []string{"a.txt", "b.go"} {
		if !found[name] {
			t.Errorf("expected to find %s in results", name)
		}
	}
}

func TestScanExt(t *testing.T) {
	base := t.TempDir()
	if err := os.WriteFile(filepath.Join(base, "test.TXT"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := Scan(base)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	if files[0].Ext != ".txt" {
		t.Errorf("expected lowercase ext .txt, got %s", files[0].Ext)
	}
}

func TestScanSymlink(t *testing.T) {
	base := t.TempDir()
	real := filepath.Join(base, "real.txt")
	link := filepath.Join(base, "link.txt")

	if err := os.WriteFile(real, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}

	files, err := Scan(base)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		if filepath.Base(f.Path) == "link.txt" {
			t.Error("symlink should be skipped")
		}
	}
}

func TestScanPermissionError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping permission test as root")
	}

	base := t.TempDir()
	noRead := filepath.Join(base, "noaccess")
	if err := os.MkdirAll(noRead, 0o000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(noRead, 0o755)

	if err := os.WriteFile(filepath.Join(base, "ok.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := Scan(base)
	if err != nil {
		t.Fatalf("Scan should not return error on permission denied, got: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("expected 1 accessible file, got %d", len(files))
	}
}

func TestScanExclude(t *testing.T) {
	base := t.TempDir()

	dirs := []string{
		filepath.Join(base, "src"),
		filepath.Join(base, "node_modules"),
		filepath.Join(base, "node_modules", "pkg"),
		filepath.Join(base, ".cache"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	fileMap := map[string]string{
		filepath.Join(base, "src", "main.go"):              "package main",
		filepath.Join(base, "node_modules", "pkg", "a.js"): "module.exports={}",
		filepath.Join(base, ".cache", "data.bin"):          "cached",
	}
	for p, content := range fileMap {
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name      string
		exclude   []string
		wantFiles int
	}{
		{"no exclude", nil, 3},
		{"exclude node_modules", []string{"node_modules"}, 2},
		{"exclude multiple", []string{"node_modules", ".cache"}, 1},
		{"exclude nonexistent", []string{"vendor"}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := Scan(base, ScanOpts{Exclude: tt.exclude})
			if err != nil {
				t.Fatalf("Scan() error: %v", err)
			}
			if len(files) != tt.wantFiles {
				names := make([]string, len(files))
				for i, f := range files {
					names[i] = f.Path
				}
				t.Errorf("expected %d files, got %d: %v", tt.wantFiles, len(files), names)
			}
		})
	}
}

func TestScanExcludeGlob(t *testing.T) {
	base := t.TempDir()

	dirs := []string{
		filepath.Join(base, "src"),
		filepath.Join(base, "build"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	fileMap := map[string]string{
		filepath.Join(base, "src", "main.go"):   "package main",
		filepath.Join(base, "src", "util.go"):   "package main",
		filepath.Join(base, "readme.txt"):       "hello",
		filepath.Join(base, "notes.tmp"):        "temp",
		filepath.Join(base, "build", "out.tmp"): "temp2",
		filepath.Join(base, "build", "app.exe"): "binary",
		filepath.Join(base, "data.log"):         "log",
		filepath.Join(base, "src", "debug.log"): "log2",
	}
	for p, content := range fileMap {
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name      string
		exclude   []string
		wantFiles int
	}{
		{"glob *.tmp", []string{"*.tmp"}, 6},
		{"glob *.log", []string{"*.log"}, 6},
		{"glob and exact", []string{"*.tmp", "build"}, 5},
		{"glob *.go", []string{"*.go"}, 6},
		{"multiple globs", []string{"*.tmp", "*.log"}, 4},
		{"glob with question mark", []string{"?.log"}, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := Scan(base, ScanOpts{Exclude: tt.exclude})
			if err != nil {
				t.Fatalf("Scan() error: %v", err)
			}
			if len(files) != tt.wantFiles {
				names := make([]string, len(files))
				for i, f := range files {
					names[i] = filepath.Base(f.Path)
				}
				t.Errorf("expected %d files, got %d: %v", tt.wantFiles, len(files), names)
			}
		})
	}
}

func TestIsGlobPattern(t *testing.T) {
	tests := []struct {
		pattern string
		want    bool
	}{
		{"node_modules", false},
		{".cache", false},
		{"*.tmp", true},
		{"file?.log", true},
		{"[abc].txt", true},
		{"vendor", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			if got := isGlobPattern(tt.pattern); got != tt.want {
				t.Errorf("isGlobPattern(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestScanOnProgress(t *testing.T) {
	base := t.TempDir()

	sub := filepath.Join(base, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, f := range []string{
		filepath.Join(base, "a.txt"),
		filepath.Join(base, "b.txt"),
		filepath.Join(sub, "c.txt"),
	} {
		if err := os.WriteFile(f, []byte("data"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	var calls []int
	var lastDir string
	files, err := Scan(base, ScanOpts{
		OnProgress: func(scanned int, currentDir string) {
			calls = append(calls, scanned)
			lastDir = currentDir
		},
	})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(calls) != len(files) {
		t.Errorf("expected %d progress calls, got %d", len(files), len(calls))
	}

	for i, v := range calls {
		if v != i+1 {
			t.Errorf("call %d: expected scanned=%d, got %d", i, i+1, v)
		}
	}

	if lastDir == "" {
		t.Error("expected lastDir to be set")
	}
}

func TestScanNonExistent(t *testing.T) {
	_, err := Scan("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}
