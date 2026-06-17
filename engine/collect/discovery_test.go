package collect

import (
	"os"
	"path/filepath"
	"testing"

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
