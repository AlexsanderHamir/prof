package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetScannerMissingFile(t *testing.T) {
	_, _, err := GetScanner(filepath.Join(t.TempDir(), "nope.prof"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetScannerReadsFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.txt")
	if err := os.WriteFile(p, []byte("line1\nline2\n"), 0o600); err != nil {
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

func TestPrintConfigurationWithAndWithoutFilters(_ *testing.T) {
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
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module cfg\n\ngo 1.24.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfgJSON := `{"function_collection_filter":{"B":{"include_prefixes":["x"]}}}`
	if err := os.WriteFile(filepath.Join(root, ConfigFilename), []byte(cfgJSON), 0o600); err != nil {
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
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module mc\n\ngo 1.24.3\n"), 0o600); err != nil {
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
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module badjson\n\ngo 1.24.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ConfigFilename), []byte("{not json"), 0o600); err != nil {
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
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module tmpl\n\ngo 1.24.3\n"), 0o600); err != nil {
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
