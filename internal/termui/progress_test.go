package termui

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestFormatProgressLabel_runBenchmark(t *testing.T) {
	t.Parallel()

	got := formatProgressLabel(Progress{
		Phase:  PhaseRunBenchmark,
		Label:  "BenchmarkFibonacci",
		Index:  2,
		Total:  2,
		Detail: "count=5",
	}, true)
	want := "Running benchmark 2/2: BenchmarkFibonacci (count=5)…"
	if got != want {
		t.Fatalf("formatProgressLabel() = %q, want %q", got, want)
	}
}

func TestFormatProgressLabel_collectProfiles(t *testing.T) {
	t.Parallel()

	got := formatProgressLabel(Progress{
		Phase:  PhaseCollectProfiles,
		Label:  "BenchmarkFibonacci",
		Detail: "cpu, memory",
	}, true)
	want := "Collecting profiles for BenchmarkFibonacci (cpu, memory)…"
	if got != want {
		t.Fatalf("formatProgressLabel() = %q, want %q", got, want)
	}
}

func TestFormatProgressLabel_collectFunctionProfiles(t *testing.T) {
	t.Parallel()

	got := formatProgressLabel(Progress{
		Phase: PhaseCollectFunctionProfiles,
		Label: "BenchmarkFibonacci",
	}, true)
	want := "Collecting function profiles for BenchmarkFibonacci…"
	if got != want {
		t.Fatalf("formatProgressLabel() = %q, want %q", got, want)
	}
}

func TestFormatProgressLabel_singleBenchmark(t *testing.T) {
	t.Parallel()

	got := formatProgressLabel(Progress{
		Phase: PhaseRunBenchmark,
		Label: "BenchmarkOnly",
		Total: 1,
		Index: 1,
	}, true)
	if !strings.Contains(got, "Running benchmark: BenchmarkOnly") {
		t.Fatalf("single benchmark label = %q", got)
	}
}

func TestSession_RunWhile_nonTTY(t *testing.T) {
	t.Parallel()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	session := NewSession(w, int(r.Fd()))
	if session.Interactive() {
		t.Fatal("expected non-interactive session for pipe fd")
	}

	var calls int
	runErr := errors.New("boom")
	err = session.RunWhile(Progress{Label: "B"}, func() error {
		calls++
		return runErr
	})
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
	if !errors.Is(err, runErr) {
		t.Fatalf("RunWhile() err = %v, want %v", err, runErr)
	}
}

func TestSession_RunWhile_nilTask(t *testing.T) {
	t.Parallel()

	session := NewSession(os.Stderr, int(os.Stderr.Fd()))
	err := session.RunWhile(Progress{}, nil)
	if err == nil {
		t.Fatal("expected error for nil task")
	}
}
