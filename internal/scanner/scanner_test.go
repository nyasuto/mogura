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

func TestScanNonExistent(t *testing.T) {
	_, err := Scan("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}
