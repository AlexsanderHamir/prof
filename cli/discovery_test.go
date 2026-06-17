package cli

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/workspace"
)

func writeTempGoMod(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module discmod\n\ngo 1.24.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestDiscoverAvailableTagsMissingBench(t *testing.T) {
	root := t.TempDir()
	writeTempGoMod(t, root)
	t.Chdir(root)
	tags, err := discoverAvailableTags()
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 0 {
		t.Fatalf("got %v", tags)
	}
}

func TestDiscoverAvailableTagsWithDirs(t *testing.T) {
	root := t.TempDir()
	writeTempGoMod(t, root)
	bench := filepath.Join(root, workspace.MainDirOutput)
	if err := os.MkdirAll(filepath.Join(bench, "alpha"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(bench, "beta"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	tags, err := discoverAvailableTags()
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(tags)
	if !slices.Equal(tags, []string{"alpha", "beta"}) {
		t.Fatalf("got %v", tags)
	}
}

func TestDiscoverAvailableBenchmarksMissingDir(t *testing.T) {
	root := t.TempDir()
	writeTempGoMod(t, root)
	t.Chdir(root)
	got, err := discoverAvailableBenchmarks("any")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatal(got)
	}
}

func TestDiscoverAvailableBenchmarksWithSubdirs(t *testing.T) {
	root := t.TempDir()
	writeTempGoMod(t, root)
	textDir := filepath.Join(root, workspace.MainDirOutput, "v1", workspace.ProfileTextDir)
	if err := os.MkdirAll(filepath.Join(textDir, "B1"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(textDir, "B2"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	got, err := discoverAvailableBenchmarks("v1")
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(got)
	if !slices.Equal(got, []string{"B1", "B2"}) {
		t.Fatalf("got %v", got)
	}
}

func TestDiscoverAvailableProfilesParsesTxt(t *testing.T) {
	root := t.TempDir()
	writeTempGoMod(t, root)
	benchName := "BenchmarkGenPool"
	dir := filepath.Join(root, workspace.MainDirOutput, "t1", workspace.ProfileTextDir, benchName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		benchName + "_cpu.txt",
		benchName + "_memory.txt",
		benchName + "_mutex.txt",
		benchName + "_block.txt",
		"other.txt",
		"BenchmarkOther_cpu.txt",
		benchName + "_trace.txt",
		"skip_dir",
	} {
		path := filepath.Join(dir, name)
		if name == "skip_dir" {
			if err := os.Mkdir(path, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(root)
	got, err := discoverAvailableProfiles("t1", benchName)
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(got)
	if !slices.Equal(got, []string{"block", "cpu", "memory", "mutex"}) {
		t.Fatalf("got %v", got)
	}
}
