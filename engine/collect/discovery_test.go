package collect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/testpaths"
	"github.com/AlexsanderHamir/prof/internal/workspace"
)

func TestScanForBenchmarks_skipsNestedModule(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module rootmod\n"), workspace.PermFile); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "root_bench_test.go"), []byte("func BenchmarkRoot(b *testing.B) {}\n"), workspace.PermFile); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "tests", "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "go.mod"), []byte("module nested\n"), workspace.PermFile); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "bench_test.go"), []byte("func BenchmarkNested(b *testing.B) {}\n"), workspace.PermFile); err != nil {
		t.Fatal(err)
	}

	names, err := scanForBenchmarks(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "BenchmarkRoot" {
		t.Fatalf("got %v want [BenchmarkRoot]", names)
	}
}

func TestScanForBenchmarks_findsBenchmarksDir(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module rootmod\n"), workspace.PermFile); err != nil {
		t.Fatal(err)
	}
	benchDir := filepath.Join(root, "benchmarks")
	if err := os.MkdirAll(benchDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(benchDir, "benchmark_test.go"), []byte("func BenchmarkExample(b *testing.B) {}\n"), workspace.PermFile); err != nil {
		t.Fatal(err)
	}

	names, err := scanForBenchmarks(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "BenchmarkExample" {
		t.Fatalf("got %v want [BenchmarkExample]", names)
	}
}

func TestDiscoverBenchmarks_repoBenchmarksFixture(t *testing.T) {
	root, err := testpaths.ModuleRoot()
	if err != nil {
		t.Skip(err)
	}
	fixture := filepath.Join(root, "benchmarks", "benchmark_test.go")
	if _, err := os.Stat(fixture); err != nil {
		t.Skip("benchmarks fixture not present in this module")
	}

	names, err := DiscoverBenchmarks(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"BenchmarkStringProcessor",
		"BenchmarkFibonacci",
		"BenchmarkMatrixMultiplication",
		"BenchmarkDataGeneration",
	}
	if len(names) != len(want) {
		t.Fatalf("got %d benchmarks %v, want %d %v", len(names), names, len(want), want)
	}
	seen := make(map[string]struct{}, len(names))
	for _, n := range names {
		seen[n] = struct{}{}
	}
	for _, n := range want {
		if _, ok := seen[n]; !ok {
			t.Fatalf("missing %s in %v", n, names)
		}
	}
}
