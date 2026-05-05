package benchmark

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestScanForBenchmarks_FindsBenchmark(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x_test.go")
	content := `package x

import "testing"

func BenchmarkHello(b *testing.B) {
	for i := 0; i < b.N; i++ {}
}
`
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	names, err := scanForBenchmarks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "BenchmarkHello" {
		t.Fatalf("got %v", names)
	}
}

func TestHandleDirectorySkipsDot(t *testing.T) {
	if err := handleDirectory(filepath.Join("a", ".git")); !errors.Is(err, filepath.SkipDir) {
		t.Fatalf("got %v", err)
	}
	if err := handleDirectory("normal"); err != nil {
		t.Fatalf("got %v", err)
	}
}
