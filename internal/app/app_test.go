package app

import "testing"

type stubBenchmark struct{}

func (stubBenchmark) RunBenchmarks(_, _ []string, _ string, _ int, _ bool) error { return nil }
func (stubBenchmark) DiscoverBenchmarks(_ string) ([]string, error)             { return nil, nil }
func (stubBenchmark) SupportedProfiles() []string                               { return nil }

func TestWithDefaultsNilReceiver(t *testing.T) {
	var s *Services
	out := s.WithDefaults()
	if out == nil || out.Benchmark == nil || out.Collector == nil || out.Tracker == nil || out.Tools == nil || out.Setup == nil {
		t.Fatalf("expected all fields set: %#v", out)
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

func TestDefaultNonNil(t *testing.T) {
	d := Default()
	if d.Benchmark == nil {
		t.Fatal("Default() should populate services")
	}
}
