package workspace_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/workspace"
)

func TestTagLayout_paths(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "mod")
	l := workspace.NewTagLayout(root, "v1")

	cases := []struct {
		name string
		got  string
		want string
	}{
		{
			"profile binary",
			l.ProfileBinary("BenchmarkFoo", "cpu"),
			filepath.Join(root, workspace.MainDirOutput, "v1", "profiles", "BenchmarkFoo", "cpu.out"),
		},
		{
			"hotspot",
			l.Hotspot("BenchmarkFoo", "cpu"),
			filepath.Join(root, workspace.MainDirOutput, "v1", "hotspots", "BenchmarkFoo", "cpu.txt"),
		},
		{
			"call tree text",
			l.CallTreeText("BenchmarkFoo", "cpu"),
			filepath.Join(root, workspace.MainDirOutput, "v1", "call_trees", "BenchmarkFoo", "cpu.txt"),
		},
		{
			"source lines",
			l.SourceLinesDir("cpu", "BenchmarkFoo"),
			filepath.Join(root, workspace.MainDirOutput, "v1", "source_lines", "cpu", "BenchmarkFoo"),
		},
		{
			"measurement",
			l.Measurement("BenchmarkFoo"),
			filepath.Join(root, workspace.MainDirOutput, "v1", "measurements", "BenchmarkFoo", "run.txt"),
		},
		{
			"call graph",
			l.CallGraph("cpu", "BenchmarkFoo"),
			filepath.Join(root, workspace.MainDirOutput, "v1", "call_graphs", "cpu", "BenchmarkFoo", "cpu.png"),
		},
		{
			"data mapping",
			l.DataMapping("BenchmarkFoo"),
			filepath.Join(root, workspace.MainDirOutput, "v1", "data_mapping", "BenchmarkFoo", "map.json"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.got != tc.want {
				t.Fatalf("got %q want %q", tc.got, tc.want)
			}
		})
	}
}

func TestRelFromTagRoot(t *testing.T) {
	t.Parallel()
	tagRoot := filepath.Join(t.TempDir(), ".prof", "v1")
	abs := filepath.Join(tagRoot, "hotspots", "BenchmarkFoo", "cpu.txt")
	got, err := workspace.RelFromTagRoot(tagRoot, abs)
	if err != nil {
		t.Fatal(err)
	}
	want := "hotspots/BenchmarkFoo/cpu.txt"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestTagLayout_ResolveProfileBinary_missing(t *testing.T) {
	t.Parallel()
	l := workspace.NewTagLayout(t.TempDir(), "t")
	if _, err := l.ResolveProfileBinary("B", "cpu"); err == nil {
		t.Fatal("expected error for missing profile binary")
	}
}

func TestCleanOrCreateTag(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "tag")
	if err := workspace.CleanOrCreateTag(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "old.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := workspace.CleanOrCreateTag(dir); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty tag dir, got %d entries", len(entries))
	}
}

func TestCleanOrCreateTag_rejectsFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "notdir")
	if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := workspace.CleanOrCreateTag(p); err == nil {
		t.Fatal("expected error when path is a file")
	}
}

func TestStemNormalization_windowsPath(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows path normalization")
	}
	// layout uses filepath.Join; manual ingest normalizes in collect package.
	p := `C:\data\cpu.out`
	base := filepath.Base(strings.ReplaceAll(p, `\`, "/"))
	if base != "cpu.out" {
		t.Fatalf("base=%q", base)
	}
}
