package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/engine/tracker"
)

type stubBenchmark struct{}

func (stubBenchmark) RunBenchmarks(_, _ []string, _ string, _ int, _, _, _ bool) error { return nil }
func (stubBenchmark) DiscoverBenchmarks(_ string) ([]string, error)                    { return nil, nil }
func (stubBenchmark) SupportedProfiles() []string                                      { return nil }

type stubCollector struct{}

func (stubCollector) RunCollector(_ []string, _ string, _ bool) error { return nil }

type stubTracker struct{}

func (stubTracker) RunTrackAuto(_ *tracker.Selections) error   { return nil }
func (stubTracker) RunTrackManual(_ *tracker.Selections) error { return nil }

type stubTools struct{}

func (stubTools) RunBenchStats(_, _, _ string) error  { return nil }
func (stubTools) RunQcacheGrind(_, _, _ string) error { return nil }

type stubSetup struct{}

func (stubSetup) CreateTemplate() error { return nil }

func TestWithDefaultsNilReceiver(t *testing.T) {
	var s *Services
	out := s.WithDefaults()
	if out == nil || out.Benchmark == nil || out.Collector == nil || out.Tracker == nil || out.Tools == nil || out.Setup == nil {
		t.Fatalf("expected all fields set: %#v", out)
	}
}

func TestWithDefaultsAllNilFieldsNonNilReceiver(t *testing.T) {
	s := &Services{}
	out := s.WithDefaults()
	if out.Benchmark == nil || out.Collector == nil || out.Tracker == nil || out.Tools == nil || out.Setup == nil {
		t.Fatalf("expected defaults for every slot: %#v", out)
	}
}

func TestWithDefaultsFillsOnlyNil(t *testing.T) {
	custom := stubBenchmark{}
	s := &Services{Benchmark: custom}
	out := s.WithDefaults()
	if out.Benchmark != custom {
		t.Fatal("should preserve custom Benchmark")
	}
	if out.Collector == nil || out.Tracker == nil {
		t.Fatal("should fill nil dependencies")
	}
}

func TestWithDefaultsPreservesAllNonNil(t *testing.T) {
	s := &Services{
		Benchmark: stubBenchmark{},
		Collector: stubCollector{},
		Tracker:   stubTracker{},
		Tools:     stubTools{},
		Setup:     stubSetup{},
	}
	out := s.WithDefaults()
	if out.Benchmark != s.Benchmark || out.Collector != s.Collector || out.Tracker != s.Tracker ||
		out.Tools != s.Tools || out.Setup != s.Setup {
		t.Fatal("WithDefaults replaced a non-nil dependency")
	}
}

func TestDefaultNonNil(t *testing.T) {
	d := Default()
	if d.Benchmark == nil {
		t.Fatal("Default() should populate services")
	}
}

func TestDefaultBenchmarkDelegates(t *testing.T) {
	d := Default()
	if d.Benchmark.RunBenchmarks(nil, nil, "", 0, false, false, false) == nil {
		t.Fatal("expected error for empty benchmarks")
	}
	if d.Benchmark.RunBenchmarks([]string{"B"}, nil, "", 0, false, false, false) == nil {
		t.Fatal("expected error for empty profiles")
	}
	if p := d.Benchmark.SupportedProfiles(); p == nil {
		t.Fatal("expected non-nil slice from SupportedProfiles")
	}
	tmp := t.TempDir()
	if _, err := d.Benchmark.DiscoverBenchmarks(tmp); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultCollectorRunCollectorEmptyFiles(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := Default().Collector.RunCollector(nil, "tag", false); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultTrackerInvalidFormat(t *testing.T) {
	d := Default()
	sel := &tracker.Selections{OutputFormat: "nope"}
	if d.Tracker.RunTrackAuto(sel) == nil {
		t.Fatal("expected invalid format error")
	}
	if d.Tracker.RunTrackManual(sel) == nil {
		t.Fatal("expected invalid format error")
	}
}

func TestDefaultToolsEarlyErrors(t *testing.T) {
	t.Chdir(t.TempDir())
	d := Default()
	if d.Tools.RunBenchStats("a", "b", "c") == nil {
		t.Fatal("expected error when bench dir missing")
	}
	if d.Tools.RunQcacheGrind("tag", "bench", "cpu") == nil {
		t.Fatal("expected error when profile binary missing")
	}
}

func TestDefaultSetupCreateTemplate(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module tmpmod\n\ngo 1.24.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	if err := Default().Setup.CreateTemplate(); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(root, "config_template.json")
	if _, err := os.Stat(p); err != nil {
		t.Fatal(err)
	}
}
