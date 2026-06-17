package tracker_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/engine/tracker"
	"github.com/AlexsanderHamir/prof/internal/workspace"
)

const compareBenchName = "BenchmarkGenPool"

func installBenchTagsFromTestdata(t *testing.T) string {
	t.Helper()
	modRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(modRoot, "go.mod"), []byte("module trackfixture\n\ngo 1.24\n"), workspace.PermFile); err != nil {
		t.Fatal(err)
	}
	for _, tag := range []string{"tag1", "tag2"} {
		src := filepath.Join("testdata", tag)
		dst := filepath.Join(modRoot, workspace.MainDirOutput, tag)
		if err := copyTree(src, dst); err != nil {
			t.Fatalf("copy %s: %v", tag, err)
		}
	}
	t.Chdir(modRoot)
	return modRoot
}

func copyTree(src, dst string) error {
	return os.CopyFS(dst, os.DirFS(src))
}

func TestCompareFixture_autoTags(t *testing.T) {
	installBenchTagsFromTestdata(t)
	profileTypes := []string{"memory", "cpu", "mutex", "block"}

	for _, profileType := range profileTypes {
		t.Run(profileType, func(t *testing.T) {
			selections := tracker.Selections{
				Baseline:      "tag1",
				Current:       "tag2",
				BenchmarkName: compareBenchName,
				ProfileType:   profileType,
			}

			profileResult, err := tracker.CheckPerformanceDifferences(&selections)
			if err != nil {
				t.Fatal(err)
			}
			if profileResult == nil {
				t.Fatal("profileResult should not be nil")
			}
			if len(profileResult.FunctionChanges) == 0 {
				t.Fatal("expected function changes")
			}
			first := profileResult.FunctionChanges[0]
			if first == nil {
				t.Fatal("first report should not be nil")
			}
			if first.Report() == "" {
				t.Fatal("report is missing")
			}
		})
	}
}

func TestCompareFixture_manualPaths(t *testing.T) {
	modRoot := installBenchTagsFromTestdata(t)
	filePath1 := filepath.Join(modRoot, workspace.MainDirOutput, "tag1", "bin", compareBenchName, compareBenchName+"_cpu.out")
	filePath2 := filepath.Join(modRoot, workspace.MainDirOutput, "tag2", "bin", compareBenchName, compareBenchName+"_cpu.out")

	selections := tracker.Selections{
		Baseline: filePath1,
		Current:  filePath2,
		IsManual: true,
	}

	profileResult, err := tracker.CheckPerformanceDifferences(&selections)
	if err != nil {
		t.Fatal(err)
	}
	if profileResult == nil {
		t.Fatal("profileResult should not be nil")
	}
	if len(profileResult.FunctionChanges) == 0 {
		t.Fatal("expected function changes")
	}
	if profileResult.FunctionChanges[0].Report() == "" {
		t.Fatal("report is missing")
	}
}
