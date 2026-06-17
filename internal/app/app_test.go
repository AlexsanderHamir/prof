package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/engine/cursoragent"
)

type stubCollect struct{}

func (stubCollect) RunAuto(_ CollectAutoOptions) error            { return nil }
func (stubCollect) RunManual(_ CollectManualOptions) error        { return nil }
func (stubCollect) DiscoverBenchmarks(_ string) ([]string, error) { return nil, nil }
func (stubCollect) SupportedProfiles() []string                   { return nil }

type stubTracker struct{}

func (stubTracker) RunTrackAuto(_ TrackOptions) error   { return nil }
func (stubTracker) RunTrackManual(_ TrackOptions) error { return nil }

type stubTools struct{}

func (stubTools) RunBenchStats(_, _, _ string) error  { return nil }
func (stubTools) RunQcacheGrind(_, _, _ string) error { return nil }

func TestWithDefaultsFillsAgentWhenNil(t *testing.T) {
	s := &Services{}
	out := s.WithDefaults()
	if out.Agent == nil {
		t.Fatal("expected default Agent")
	}
	if _, err := out.Agent.Run(t.Context(), cursoragent.RunRequest{}, cursoragent.Options{}); err != nil {
		// cursor-agent may be missing on PATH; wiring must still return an error from the client, not panic.
		t.Logf("Agent.Run: %v", err)
	}
}

type stubSetup struct{}

func (stubSetup) CreateTemplate() error { return nil }

func TestWithDefaultsNilReceiver(t *testing.T) {
	var s *Services
	out := s.WithDefaults()
	if out == nil || out.Runner == nil || out.Collect == nil || out.Tracker == nil || out.Tools == nil || out.Setup == nil || out.Config == nil {
		t.Fatalf("expected all fields set: %#v", out)
	}
}

func TestWithDefaultsAllNilFieldsNonNilReceiver(t *testing.T) {
	s := &Services{}
	out := s.WithDefaults()
	if out.Collect == nil || out.Tracker == nil || out.Tools == nil || out.Setup == nil {
		t.Fatalf("expected defaults for every slot: %#v", out)
	}
}

func TestWithDefaultsFillsOnlyNil(t *testing.T) {
	custom := stubCollect{}
	s := &Services{Collect: custom}
	out := s.WithDefaults()
	if out.Collect != custom {
		t.Fatal("should preserve custom Collect")
	}
	if out.Tracker == nil {
		t.Fatal("should fill nil dependencies")
	}
}

func TestWithDefaultsPreservesAllNonNil(t *testing.T) {
	s := &Services{
		Collect: stubCollect{},
		Tracker: stubTracker{},
		Tools:   stubTools{},
		Setup:   stubSetup{},
	}
	out := s.WithDefaults()
	if out.Collect != s.Collect || out.Tracker != s.Tracker || out.Tools != s.Tools || out.Setup != s.Setup {
		t.Fatal("WithDefaults replaced a non-nil dependency")
	}
}

func TestDefaultNonNil(t *testing.T) {
	d := Default()
	if d.Collect == nil {
		t.Fatal("Default() should populate services")
	}
}

func TestDefaultCollectDelegates(t *testing.T) {
	d := Default()
	if d.Collect.RunAuto(CollectAutoOptions{}) == nil {
		t.Fatal("expected error for empty benchmarks")
	}
	if d.Collect.RunAuto(CollectAutoOptions{Benchmarks: []string{"B"}}) == nil {
		t.Fatal("expected error for empty profiles")
	}
	if p := d.Collect.SupportedProfiles(); p == nil {
		t.Fatal("expected non-nil slice from SupportedProfiles")
	}
	tmp := t.TempDir()
	if _, err := d.Collect.DiscoverBenchmarks(tmp); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultCollectRunManualEmptyFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module tmpmod\n\ngo 1.24.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	if err := Default().Collect.RunManual(CollectManualOptions{Tag: "tag"}); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultTrackerInvalidFormat(t *testing.T) {
	d := Default()
	opts := TrackOptions{OutputFormat: "nope"}
	if d.Tracker.RunTrackAuto(opts) == nil {
		t.Fatal("expected invalid format error")
	}
	if d.Tracker.RunTrackManual(opts) == nil {
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
	p := filepath.Join(root, "prof.json")
	if _, err := os.Stat(p); err != nil {
		t.Fatal(err)
	}
}
