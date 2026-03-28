package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module nest\n\ngo 1.24.3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "pkg", "deep")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)
	got, err := FindGoModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	absEq(t, got, root)
}

func TestFindGoModuleRoot_NotFound(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	_, err := FindGoModuleRoot()
	if err == nil || !strings.Contains(err.Error(), "go.mod not found") {
		t.Fatalf("got %v", err)
	}
}

func TestGetScannerMissingFile(t *testing.T) {
	_, _, err := GetScanner(filepath.Join(t.TempDir(), "nope.prof"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetScannerReadsFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.txt")
	if err := os.WriteFile(p, []byte("line1\nline2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sc, f, err := GetScanner(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if !sc.Scan() || sc.Text() != "line1" {
		t.Fatal()
	}
}

func TestCleanOrCreateTagCreatesMissing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "tag1")
	if err := CleanOrCreateTag(dir); err != nil {
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
	if err := os.WriteFile(filepath.Join(dir, "nested", "f.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CleanOrCreateTag(dir); err != nil {
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
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CleanOrCreateTag(p); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("got %v", err)
	}
}

func TestPrintConfigurationWithAndWithoutFilters(t *testing.T) {
	PrintConfiguration(&BenchArgs{
		Benchmarks: []string{"B"},
		Profiles:   []string{"cpu"},
		Tag:        "t",
		Count:      1,
	}, nil)
	PrintConfiguration(&BenchArgs{Tag: "t2"}, map[string]FunctionFilter{
		"Bench": {IncludePrefixes: []string{"p"}, IgnoreFunctions: []string{"init"}},
	})
}

func TestLoadFromFileSuccess(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module cfg\n\ngo 1.24.3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfgJSON := `{"function_collection_filter":{"B":{"include_prefixes":["x"]}}}`
	if err := os.WriteFile(filepath.Join(root, ConfigFilename), []byte(cfgJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	cfg, err := LoadFromFile(ConfigFilename)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.FunctionFilter == nil || cfg.FunctionFilter["B"].IncludePrefixes[0] != "x" {
		t.Fatalf("%+v", cfg)
	}
}

func TestLoadFromFileModuleRootFailure(t *testing.T) {
	t.Chdir(t.TempDir())
	_, err := LoadFromFile(ConfigFilename)
	if err == nil || !strings.Contains(err.Error(), "module root") {
		t.Fatalf("got %v", err)
	}
}

func TestLoadFromFileMissingConfig(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module mc\n\ngo 1.24.3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	_, err := LoadFromFile(ConfigFilename)
	if err == nil || !strings.Contains(err.Error(), "read config") {
		t.Fatalf("got %v", err)
	}
}

func TestLoadFromFileInvalidJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module badjson\n\ngo 1.24.3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ConfigFilename), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	_, err := LoadFromFile(ConfigFilename)
	if err == nil || !strings.Contains(err.Error(), "parse config") {
		t.Fatalf("got %v", err)
	}
}

func TestCreateTemplateWritesFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module tmpl\n\ngo 1.24.3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	if err := CreateTemplate(); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(root, "config_template.json")
	if _, err := os.Stat(p); err != nil {
		t.Fatal(err)
	}
}
