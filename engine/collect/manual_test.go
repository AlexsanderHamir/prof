package collect

import (
	"context"
	"testing"

	"github.com/AlexsanderHamir/prof/engine/tooling"
)

func TestManualBenchAndProfile(t *testing.T) {
	t.Parallel()
	cases := []struct {
		path        string
		wantBench   string
		wantProfile string
	}{
		{"cpu.out", "cpu", "cpu"},
		{"BenchmarkFoo_cpu.out", "BenchmarkFoo", "cpu"},
		{"mybench_memory.out", "mybench", "memory"},
		{`C:\data\block.out`, "block", "block"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()
			bench, profile := manualBenchAndProfile(tc.path)
			if bench != tc.wantBench || profile != tc.wantProfile {
				t.Fatalf("got bench=%q profile=%q want bench=%q profile=%q", bench, profile, tc.wantBench, tc.wantProfile)
			}
		})
	}
}

func TestRunAuto_validation(t *testing.T) {
	t.Parallel()
	if err := RunAuto(nil, AutoOptions{}); err == nil {
		t.Fatal("expected nil runner error")
	}
	if err := RunAuto(noopRunner{}, AutoOptions{}); err == nil {
		t.Fatal("expected empty benchmarks error")
	}
	if err := RunAuto(noopRunner{}, AutoOptions{Benchmarks: []string{"B"}}); err == nil {
		t.Fatal("expected empty profiles error")
	}
	if err := RunAuto(noopRunner{}, AutoOptions{Benchmarks: []string{"B"}, Profiles: []string{"cpu"}, Count: 0}); err == nil {
		t.Fatal("expected count error")
	}
}

func TestRunManual_validation(t *testing.T) {
	t.Parallel()
	if err := RunManual(nil, ManualOptions{Tag: "t"}); err == nil {
		t.Fatal("expected nil runner error")
	}
}

type noopRunner struct{}

func (noopRunner) Run(_ context.Context, _ []string, _ tooling.RunOpts) ([]byte, error) {
	return nil, nil
}
