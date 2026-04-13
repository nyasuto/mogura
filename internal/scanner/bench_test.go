package scanner

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func createBenchTree(b *testing.B, root string, dirs, filesPerDir int) {
	b.Helper()
	for d := 0; d < dirs; d++ {
		dir := filepath.Join(root, "dir", string(rune('a'+d/26))+string(rune('a'+d%26)))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			b.Fatal(err)
		}
		for f := 0; f < filesPerDir; f++ {
			path := filepath.Join(dir, "file"+string(rune('a'+f/26))+string(rune('a'+f%26))+".txt")
			if err := os.WriteFile(path, []byte("bench"), 0o644); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkScanWorkers1(b *testing.B) {
	tmp := b.TempDir()
	createBenchTree(b, tmp, 200, 50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Scan(tmp, ScanOpts{Workers: 1})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkScanWorkersNumCPU(b *testing.B) {
	tmp := b.TempDir()
	createBenchTree(b, tmp, 200, 50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Scan(tmp, ScanOpts{Workers: runtime.NumCPU()})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkScanLargeFlat(b *testing.B) {
	tmp := b.TempDir()
	createBenchTree(b, tmp, 10, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Scan(tmp, ScanOpts{Workers: runtime.NumCPU()})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkScanDeepNested(b *testing.B) {
	tmp := b.TempDir()
	dir := tmp
	for depth := 0; depth < 50; depth++ {
		dir = filepath.Join(dir, "d")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			b.Fatal(err)
		}
		for f := 0; f < 20; f++ {
			path := filepath.Join(dir, "f"+string(rune('a'+f))+".dat")
			if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
				b.Fatal(err)
			}
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Scan(tmp, ScanOpts{Workers: runtime.NumCPU()})
		if err != nil {
			b.Fatal(err)
		}
	}
}
