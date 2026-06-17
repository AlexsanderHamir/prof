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

func TestResolveCollectionFilter_defaults(t *testing.T) {
	cfg := &config.Config{
		Collection: config.Collection{
			Defaults: config.FunctionFilter{IncludePrefixes: []string{"p"}},
		},
	}
	got := config.ResolveCollectionFilter(cfg, config.CollectionTargetAuto("BenchmarkX"))
	if len(got.IncludePrefixes) != 1 || got.IncludePrefixes[0] != "p" {
		t.Fatalf("got %+v", got)
	}
}

func TestResolveCollectionFilter_perBenchOverrides(t *testing.T) {
	cfg := &config.Config{
		Collection: config.Collection{
			Defaults: config.FunctionFilter{IncludePrefixes: []string{"default"}},
			Benchmarks: map[string]config.FunctionFilter{
				"BenchmarkX": {IgnoreFunctions: []string{"init"}},
			},
		},
	}
	got := config.ResolveCollectionFilter(cfg, config.CollectionTargetAuto("BenchmarkX"))
	if len(got.IncludePrefixes) != 1 || got.IncludePrefixes[0] != "default" {
		t.Fatalf("expected inherited prefix, got %+v", got)
	}
	if len(got.IgnoreFunctions) != 1 || got.IgnoreFunctions[0] != "init" {
		t.Fatalf("got %+v", got)
	}
}

func TestResolveCollectionFilter_manualProfile(t *testing.T) {
	cfg := &config.Config{
		Collection: config.Collection{
			ManualProfiles: map[string]config.FunctionFilter{
				"BenchmarkX_cpu": {IncludePrefixes: []string{"pkg"}},
			},
		},
	}
	got := config.ResolveCollectionFilter(cfg, config.CollectionTargetManual("BenchmarkX_cpu"))
	if len(got.IncludePrefixes) != 1 || got.IncludePrefixes[0] != "pkg" {
		t.Fatalf("got %+v", got)
	}
}

func TestResolveCollectionFilter_emptyConfig(t *testing.T) {
	got := config.ResolveCollectionFilter(nil, config.CollectionTargetAuto("x"))
	if got.IncludePrefixes != nil || got.IgnoreFunctions != nil {
		t.Fatalf("got %+v", got)
	}
}

func TestResolveTrackPolicy_merge(t *testing.T) {
	cfg := &config.Config{
		Track: config.Track{
			Defaults: config.TrackPolicy{
				IgnorePrefixes:       []string{"runtime."},
				MaxRegressionPercent: 15,
			},
			Benchmarks: map[string]config.TrackPolicy{
				"BenchmarkX": {MaxRegressionPercent: 10},
			},
		},
	}
	got := config.ResolveTrackPolicy(cfg, "BenchmarkX")
	if got.MaxRegressionPercent != 10 {
		t.Fatalf("got %v", got.MaxRegressionPercent)
	}
	if len(got.IgnorePrefixes) != 1 {
		t.Fatalf("expected inherited ignore prefix, got %+v", got)
	}
}

func TestSaveRoundTrip(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/foo\n\ngo 1.24\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)

	cfg := config.Default()
	cfg.Collection.Benchmarks = map[string]config.FunctionFilter{
		"BenchmarkFoo": {IgnoreFunctions: []string{"init"}},
	}
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Version != config.CurrentVersion {
		t.Fatalf("version %d", loaded.Version)
	}
	if len(loaded.Collection.Benchmarks["BenchmarkFoo"].IgnoreFunctions) != 1 {
		t.Fatalf("got %+v", loaded.Collection.Benchmarks)
	}
}

func TestValidate_rejectsNegativeThreshold(t *testing.T) {
	cfg := &config.Config{
		Version: config.CurrentVersion,
		Track: config.Track{
			Defaults: config.TrackPolicy{MaxRegressionPercent: -1},
		},
	}
	if err := config.Validate(cfg); err == nil {
		t.Fatal("expected error")
	}
}
