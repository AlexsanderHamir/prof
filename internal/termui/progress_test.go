package termui

import (
	"bytes"
	"errors"
	"io"
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

func TestFormatProgressLabel_doneOmitsEllipsis(t *testing.T) {
	t.Parallel()

	got := formatProgressLabel(Progress{Phase: PhasePrepare}, false)
	if got != "Preparing" || strings.Contains(got, "…") {
		t.Fatalf("done label = %q, want Preparing without ellipsis", got)
	}
}

func TestSession_RunWhile_persistentDoneAndWarnings(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := newSessionForTest(&buf, true)
	err := s.RunWhile(Progress{Phase: PhasePrepare}, func() error {
		s.Warn("setup warn")
		return nil
	})
	if err != nil {
		t.Fatalf("RunWhile() err = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Preparing") {
		t.Fatalf("output missing stage label: %q", out)
	}
	if !strings.Contains(out, "warning:") {
		t.Fatalf("output missing warning prefix: %q", out)
	}
	if !warningBelowStageHeader(out, "Preparing") {
		t.Fatalf("warning should be on its own line below the stage header: %q", out)
	}
	// lipgloss wraps done marker; check plain check character is present.
	if !strings.Contains(out, "✓") {
		t.Fatalf("output missing done marker: %q", out)
	}
}

func TestSession_Warn_indentOutsideStageFallsBackToSlog(t *testing.T) {
	t.Parallel()

	s := newSessionForTest(io.Discard, true)
	// Should not panic when no stage is active.
	s.Warn("orphan")
}

func warningBelowStageHeader(out, stage string) bool {
	stageIdx := strings.Index(out, stage)
	if stageIdx < 0 {
		return false
	}
	warnIdx := strings.Index(out[stageIdx:], "warning:")
	if warnIdx < 0 {
		return false
	}
	return strings.Contains(out[stageIdx:stageIdx+warnIdx], "\n")
}

func TestSession_WarningsStayBelowHeaderWithMultipleWarns(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := newSessionForTest(&buf, true)
	err := s.RunWhile(Progress{
		Phase:  PhaseCollectProfiles,
		Label:  "BenchmarkFoo",
		Detail: "cpu, memory",
	}, func() error {
		s.Warn("first warn")
		s.Warn("second warn")
		return nil
	})
	if err != nil {
		t.Fatalf("RunWhile() err = %v", err)
	}
	out := buf.String()
	if !warningBelowStageHeader(out, "Collecting profiles for BenchmarkFoo") {
		t.Fatalf("warnings not below header: %q", out)
	}
	first := strings.Index(out, "first warn")
	second := strings.Index(out, "second warn")
	if first < 0 || second < 0 || second <= first {
		t.Fatalf("expected two warning lines in order: %q", out)
	}
}
