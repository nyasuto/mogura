package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"mogura/internal"
)

func sortFilesByPath(files []internal.FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
}

func fileSet(files []internal.FileInfo) map[string]int64 {
	m := make(map[string]int64, len(files))
	for _, f := range files {
		m[f.Path] = f.Size
	}
	return m
}

func createTree(t *testing.T, base string, structure map[string]string) {
	t.Helper()
	for rel, content := range structure {
		abs := filepath.Join(base, rel)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestParallelBasic(t *testing.T) {
	base := t.TempDir()
	createTree(t, base, map[string]string{
		"a.txt":        "hello",
		"sub/b.go":     "package main",
		"sub/deep/c.c": "int main(){}",
	})

	files, err := Scan(base, ScanOpts{Workers: 4})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files))
	}

	set := fileSet(files)
	for _, name := range []string{"a.txt", "b.go", "c.c"} {
		found := false
		for p := range set {
			if filepath.Base(p) == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find %s in results", name)
		}
	}
}

func TestParallelSymlinkSkip(t *testing.T) {
	base := t.TempDir()
	createTree(t, base, map[string]string{
		"real.txt": "data",
	})
	if err := os.Symlink(filepath.Join(base, "real.txt"), filepath.Join(base, "link.txt")); err != nil {
		t.Fatal(err)
	}

	files, err := Scan(base, ScanOpts{Workers: 4})
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		if filepath.Base(f.Path) == "link.txt" {
			t.Error("symlink should be skipped")
		}
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}
}

func TestParallelPermissionError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping permission test as root")
	}

	base := t.TempDir()
	createTree(t, base, map[string]string{
		"ok.txt": "ok",
	})
	noRead := filepath.Join(base, "noaccess")
	if err := os.MkdirAll(noRead, 0o000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(noRead, 0o755)

	files, err := Scan(base, ScanOpts{Workers: 4})
	if err != nil {
		t.Fatalf("Scan should not error on permission denied, got: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 accessible file, got %d", len(files))
	}
}

func TestParallelExclude(t *testing.T) {
	base := t.TempDir()
	createTree(t, base, map[string]string{
		"src/main.go":           "package main",
		"node_modules/pkg/a.js": "module.exports={}",
		".cache/data.bin":       "cached",
	})

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
			files, err := Scan(base, ScanOpts{Exclude: tt.exclude, Workers: 4})
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

func TestParallelExcludeGlob(t *testing.T) {
	base := t.TempDir()
	createTree(t, base, map[string]string{
		"src/main.go":   "package main",
		"src/util.go":   "package main",
		"readme.txt":    "hello",
		"notes.tmp":     "temp",
		"build/out.tmp": "temp2",
		"build/app.exe": "binary",
		"data.log":      "log",
		"src/debug.log": "log2",
	})

	tests := []struct {
		name      string
		exclude   []string
		wantFiles int
	}{
		{"glob *.tmp", []string{"*.tmp"}, 6},
		{"glob *.log", []string{"*.log"}, 6},
		{"glob and exact", []string{"*.tmp", "build"}, 5},
		{"multiple globs", []string{"*.tmp", "*.log"}, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := Scan(base, ScanOpts{Exclude: tt.exclude, Workers: 4})
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

func TestParallelOneFileSystem(t *testing.T) {
	base := t.TempDir()
	createTree(t, base, map[string]string{
		"a.txt":     "data",
		"sub/b.txt": "data",
	})

	filesWith, err := Scan(base, ScanOpts{OneFileSystem: true, Workers: 4})
	if err != nil {
		t.Fatal(err)
	}
	filesWithout, err := Scan(base, ScanOpts{OneFileSystem: false, Workers: 4})
	if err != nil {
		t.Fatal(err)
	}

	if len(filesWith) != len(filesWithout) {
		t.Errorf("same filesystem: expected equal counts, got %d vs %d", len(filesWith), len(filesWithout))
	}
	if len(filesWith) != 2 {
		t.Errorf("expected 2 files, got %d", len(filesWith))
	}
}

func TestParallelPhysicalSize(t *testing.T) {
	base := t.TempDir()

	sparse := filepath.Join(base, "sparse.raw")
	f, err := os.Create(sparse)
	if err != nil {
		t.Fatal(err)
	}
	const logicalSize = 1 << 30
	if err := f.Truncate(logicalSize); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	files, err := Scan(base, ScanOpts{Workers: 4})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	fi := files[0]
	if fi.Size != logicalSize {
		t.Errorf("expected logical size %d, got %d", logicalSize, fi.Size)
	}
	if fi.PhysicalSize >= fi.Size {
		t.Errorf("sparse file: expected PhysicalSize (%d) << Size (%d)", fi.PhysicalSize, fi.Size)
	}
}

func TestParallelOnProgress(t *testing.T) {
	base := t.TempDir()
	createTree(t, base, map[string]string{
		"a.txt":     "data",
		"b.txt":     "data",
		"sub/c.txt": "data",
	})

	var callCount int
	files, err := Scan(base, ScanOpts{
		Workers: 2,
		OnProgress: func(scanned int, currentDir string) {
			callCount++
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if callCount != len(files) {
		t.Errorf("expected %d progress calls, got %d", len(files), callCount)
	}
}

func TestParallelOrderIndependent(t *testing.T) {
	base := t.TempDir()
	createTree(t, base, map[string]string{
		"a/1.txt":     "a1",
		"a/2.txt":     "a2",
		"b/1.txt":     "b1",
		"b/2.txt":     "b2",
		"c/d/1.txt":   "cd1",
		"c/d/2.txt":   "cd2",
		"c/d/3.txt":   "cd3",
		"e/f/g/1.txt": "efg1",
		"root.txt":    "root",
	})

	runs := 10
	var reference map[string]int64

	for i := 0; i < runs; i++ {
		files, err := Scan(base, ScanOpts{Workers: 4})
		if err != nil {
			t.Fatalf("run %d: Scan() error: %v", i, err)
		}

		set := fileSet(files)
		if reference == nil {
			reference = set
			continue
		}

		if len(set) != len(reference) {
			t.Fatalf("run %d: file count %d != reference %d", i, len(set), len(reference))
		}
		for path, size := range reference {
			if got, ok := set[path]; !ok {
				t.Errorf("run %d: missing file %s", i, path)
			} else if got != size {
				t.Errorf("run %d: file %s size %d != reference %d", i, path, got, size)
			}
		}
	}
}

func TestParallelDeterministicResults(t *testing.T) {
	base := t.TempDir()

	tree := make(map[string]string)
	for _, dir := range []string{"a", "b", "c", "d", "e"} {
		for j := 0; j < 5; j++ {
			key := filepath.Join(dir, string(rune('0'+j))+".dat")
			tree[key] = "content"
		}
	}
	createTree(t, base, tree)

	files1, err := Scan(base, ScanOpts{Workers: 1})
	if err != nil {
		t.Fatal(err)
	}
	files8, err := Scan(base, ScanOpts{Workers: 8})
	if err != nil {
		t.Fatal(err)
	}

	sortFilesByPath(files1)
	sortFilesByPath(files8)

	if len(files1) != len(files8) {
		t.Fatalf("Workers=1 got %d files, Workers=8 got %d", len(files1), len(files8))
	}

	for i := range files1 {
		if files1[i].Path != files8[i].Path {
			t.Errorf("index %d: path mismatch %s vs %s", i, files1[i].Path, files8[i].Path)
		}
		if files1[i].Size != files8[i].Size {
			t.Errorf("index %d: size mismatch %d vs %d", i, files1[i].Size, files8[i].Size)
		}
		if files1[i].PhysicalSize != files8[i].PhysicalSize {
			t.Errorf("index %d: physical size mismatch %d vs %d", i, files1[i].PhysicalSize, files8[i].PhysicalSize)
		}
		if files1[i].Ext != files8[i].Ext {
			t.Errorf("index %d: ext mismatch %s vs %s", i, files1[i].Ext, files8[i].Ext)
		}
	}
}
