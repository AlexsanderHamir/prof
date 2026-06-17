package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/config"
)

func TestLoadFromFile_missing(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n\ngo 1.24\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	if _, err := config.LoadFromFile("missing.json"); err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveFilter_global(t *testing.T) {
	cfg := &config.Config{
		FunctionFilter: map[string]config.FunctionFilter{
			config.GlobalSign: {IncludePrefixes: []string{"p"}},
		},
	}
	got := config.ResolveFilter(cfg, "BenchmarkX")
	if len(got.IncludePrefixes) != 1 || got.IncludePrefixes[0] != "p" {
		t.Fatalf("got %+v", got)
	}
}

func TestResolveFilter_perBench(t *testing.T) {
	cfg := &config.Config{
		FunctionFilter: map[string]config.FunctionFilter{
			"BenchmarkX": {IgnoreFunctions: []string{"init"}},
		},
	}
	got := config.ResolveFilter(cfg, "BenchmarkX")
	if len(got.IgnoreFunctions) != 1 {
		t.Fatalf("got %+v", got)
	}
}

func TestResolveFilter_emptyConfig(t *testing.T) {
	got := config.ResolveFilter(nil, "x")
	if got.IncludePrefixes != nil || got.IgnoreFunctions != nil {
		t.Fatalf("got %+v", got)
	}
}
