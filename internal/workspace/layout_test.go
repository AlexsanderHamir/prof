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
		{"bin", l.Bin("BenchmarkFoo", "cpu"), filepath.Join(root, "bench", "v1", "bin", "BenchmarkFoo", "BenchmarkFoo_cpu.out")},
		{"text", l.Text("BenchmarkFoo", "cpu"), filepath.Join(root, "bench", "v1", "text", "BenchmarkFoo", "BenchmarkFoo_cpu.txt")},
		{"grouped", l.Grouped("BenchmarkFoo", "cpu"), filepath.Join(root, "bench", "v1", "text", "BenchmarkFoo", "BenchmarkFoo_cpu_grouped.txt")},
		{"functions", l.FunctionsDir("cpu", "BenchmarkFoo"), filepath.Join(root, "bench", "v1", "cpu_functions", "BenchmarkFoo")},
		{"bench text", l.BenchText("BenchmarkFoo"), filepath.Join(root, "bench", "v1", "text", "BenchmarkFoo", "BenchmarkFoo.txt")},
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

func TestTagLayout_ResolveBin_missing(t *testing.T) {
	t.Parallel()
	l := workspace.NewTagLayout(t.TempDir(), "t")
	if _, err := l.ResolveBin("B", "cpu"); err == nil {
		t.Fatal("expected error for missing bin")
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
