package analyzer

import (
	"mogura/internal"
	"testing"
	"time"
)

func TestAnalyze(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	oldTime := now.AddDate(-2, 0, 0)

	files := []internal.FileInfo{
		{Path: "/app/main.go", Dir: "/app", Ext: ".go", Size: 1000, PhysicalSize: 1000, ModTime: now},
		{Path: "/app/util.go", Dir: "/app", Ext: ".go", Size: 500, PhysicalSize: 500, ModTime: now},
		{Path: "/app/img/photo.jpg", Dir: "/app/img", Ext: ".jpg", Size: 5000, PhysicalSize: 4096, ModTime: oldTime},
		{Path: "/app/data/old.csv", Dir: "/app/data", Ext: ".csv", Size: 3000, PhysicalSize: 3000, ModTime: oldTime},
	}

	opts := AnalyzeOpts{
		TopN:          2,
		Depth:         1,
		OlderThanDays: 365,
		Now:           now,
	}

	result := Analyze(files, opts)

	t.Run("TotalSize", func(t *testing.T) {
		if result.TotalSize != 9500 {
			t.Errorf("TotalSize = %d, want 9500", result.TotalSize)
		}
	})

	t.Run("TotalPhysicalSize", func(t *testing.T) {
		if result.TotalPhysicalSize != 8596 {
			t.Errorf("TotalPhysicalSize = %d, want 8596", result.TotalPhysicalSize)
		}
	})

	t.Run("DirSizes", func(t *testing.T) {
		if result.DirSizes["/app"].Size != 1500 {
			t.Errorf("DirSizes[/app].Size = %d, want 1500", result.DirSizes["/app"].Size)
		}
		if result.DirSizes["/app"].PhysicalSize != 1500 {
			t.Errorf("DirSizes[/app].PhysicalSize = %d, want 1500", result.DirSizes["/app"].PhysicalSize)
		}
		if result.DirSizes["/app/img"].Size != 5000 {
			t.Errorf("DirSizes[/app/img].Size = %d, want 5000", result.DirSizes["/app/img"].Size)
		}
		if result.DirSizes["/app/img"].PhysicalSize != 4096 {
			t.Errorf("DirSizes[/app/img].PhysicalSize = %d, want 4096", result.DirSizes["/app/img"].PhysicalSize)
		}
	})

	t.Run("ExtStats", func(t *testing.T) {
		goStats, ok := result.ExtStats[".go"]
		if !ok {
			t.Fatal("ExtStats missing .go")
		}
		if goStats.Size != 1500 || goStats.Count != 2 {
			t.Errorf("ExtStats[.go] = {Size:%d, Count:%d}, want {1500, 2}", goStats.Size, goStats.Count)
		}
	})

	t.Run("CategoryStats", func(t *testing.T) {
		if _, ok := result.CategoryStats[CategoryCode]; !ok {
			t.Error("CategoryStats missing コード")
		}
		if _, ok := result.CategoryStats[CategoryImage]; !ok {
			t.Error("CategoryStats missing 画像")
		}
	})

	t.Run("TopFiles", func(t *testing.T) {
		if len(result.TopFiles) != 2 {
			t.Fatalf("TopFiles len = %d, want 2", len(result.TopFiles))
		}
		if result.TopFiles[0].Size != 5000 {
			t.Errorf("TopFiles[0].Size = %d, want 5000", result.TopFiles[0].Size)
		}
	})

	t.Run("DirTree", func(t *testing.T) {
		if result.DirTree.Size != 9500 {
			t.Errorf("DirTree.Size = %d, want 9500", result.DirTree.Size)
		}
	})

	t.Run("StaleSummary", func(t *testing.T) {
		if result.StaleSummary.TotalFiles != 2 {
			t.Errorf("StaleSummary.TotalFiles = %d, want 2", result.StaleSummary.TotalFiles)
		}
		if result.StaleSummary.TotalSize != 8000 {
			t.Errorf("StaleSummary.TotalSize = %d, want 8000", result.StaleSummary.TotalSize)
		}
	})

	t.Run("SavingsEstimate", func(t *testing.T) {
		// WasteDirs=0(no waste patterns), StaleSummary.TotalPhysicalSize=4096+3000=7096
		if result.SavingsEstimate != 7096 {
			t.Errorf("SavingsEstimate = %d, want 7096", result.SavingsEstimate)
		}
	})
}

func TestAnalyzeDefaults(t *testing.T) {
	files := []internal.FileInfo{
		{Path: "/a/f.txt", Dir: "/a", Ext: ".txt", Size: 100, ModTime: time.Now()},
	}

	result := Analyze(files, AnalyzeOpts{})

	if result.TotalSize != 100 {
		t.Errorf("TotalSize = %d, want 100", result.TotalSize)
	}
	if len(result.TopFiles) != 1 {
		t.Errorf("TopFiles len = %d, want 1", len(result.TopFiles))
	}
}

func TestAnalyzeEmpty(t *testing.T) {
	result := Analyze(nil, AnalyzeOpts{})

	if result.TotalSize != 0 {
		t.Errorf("TotalSize = %d, want 0", result.TotalSize)
	}
	if len(result.DirSizes) != 0 {
		t.Errorf("DirSizes len = %d, want 0", len(result.DirSizes))
	}
}

func TestSavingsEstimateAlwaysPhysical(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	oldTime := now.AddDate(-2, 0, 0)

	files := []internal.FileInfo{
		{Path: "/app/node_modules/pkg/index.js", Dir: "/app/node_modules/pkg", Ext: ".js", Size: 10000, PhysicalSize: 2000, ModTime: now},
		{Path: "/app/old.dat", Dir: "/app", Ext: ".dat", Size: 50000, PhysicalSize: 5000, ModTime: oldTime},
	}

	result := Analyze(files, AnalyzeOpts{Now: now, OlderThanDays: 365})

	wastePhysical := result.WasteDirs[0].PhysicalSize
	stalePhysical := result.StaleSummary.TotalPhysicalSize
	wantSavings := wastePhysical + stalePhysical

	if result.SavingsEstimate != wantSavings {
		t.Errorf("SavingsEstimate = %d, want %d (physical-based); logical waste=%d, logical stale=%d",
			result.SavingsEstimate, wantSavings,
			result.WasteDirs[0].Size, result.StaleSummary.TotalSize)
	}
	if result.SavingsEstimate == result.WasteDirs[0].Size+result.StaleSummary.TotalSize {
		t.Error("SavingsEstimate equals logical sizes; should use physical sizes")
	}
}

func TestAnalyzeWasteDirs(t *testing.T) {
	files := []internal.FileInfo{
		{Path: "/app/node_modules/pkg/index.js", Dir: "/app/node_modules/pkg", Ext: ".js", Size: 2000, PhysicalSize: 1500, ModTime: time.Now()},
		{Path: "/app/src/main.go", Dir: "/app/src", Ext: ".go", Size: 500, PhysicalSize: 500, ModTime: time.Now()},
	}

	result := Analyze(files, AnalyzeOpts{})

	if len(result.WasteDirs) != 1 {
		t.Fatalf("WasteDirs len = %d, want 1", len(result.WasteDirs))
	}
	if result.WasteDirs[0].Kind != "node_modules" {
		t.Errorf("WasteDirs[0].Kind = %q, want %q", result.WasteDirs[0].Kind, "node_modules")
	}
	if result.SavingsEstimate != 1500 {
		t.Errorf("SavingsEstimate = %d, want 1500 (physical size of waste)", result.SavingsEstimate)
	}
}
