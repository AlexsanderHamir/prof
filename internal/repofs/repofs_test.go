package repofs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/repofs"
)

func absEq(t *testing.T, a, b string) {
	t.Helper()
	aa, err := filepath.Abs(a)
	if err != nil {
		t.Fatal(err)
	}
	bb, err := filepath.Abs(b)
	if err != nil {
		t.Fatal(err)
	}
	if aa != bb {
		t.Fatalf("paths differ:\n  %q\n  %q", aa, bb)
	}
}

func TestFindGoModuleRoot_FromNestedDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module nest\n\ngo 1.24.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "pkg", "deep")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)
	got, err := repofs.FindGoModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	absEq(t, got, root)
}

func TestFindGoModuleRoot_NotFound(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	_, err := repofs.FindGoModuleRoot()
	if err == nil || !strings.Contains(err.Error(), "go.mod not found") {
		t.Fatalf("got %v", err)
	}
}

func TestCleanOrCreateTagCreatesMissing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "tag1")
	if err := repofs.CleanOrCreateTag(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		t.Fatal(err)
	}
}

func TestCleanOrCreateTagCleansExistingDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "tag2")
	if err := os.MkdirAll(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "nested", "f.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := repofs.CleanOrCreateTag(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("leftover: %v", entries)
	}
}

func TestCleanOrCreateTagRejectsFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "notdir")
	if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := repofs.CleanOrCreateTag(p, 0o755); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("got %v", err)
	}
}
