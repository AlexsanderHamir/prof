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
		Detail: "count=5",
	}, true)
	want := "  0) Run benchmark (count=5)…"
	if got != want {
		t.Fatalf("formatProgressLabel() = %q, want %q", got, want)
	}
}

func TestFormatProgressLabel_collectProfiles(t *testing.T) {
	t.Parallel()

	got := formatProgressLabel(Progress{
		Phase:  PhaseCollectProfiles,
		Detail: "cpu, memory",
	}, true)
	want := "  1) Collect profiles (cpu, memory)…"
	if got != want {
		t.Fatalf("formatProgressLabel() = %q, want %q", got, want)
	}
}

func TestFormatProgressLabel_collectFunctionProfiles(t *testing.T) {
	t.Parallel()

	got := formatProgressLabel(Progress{
		Phase: PhaseCollectFunctionProfiles,
	}, true)
	want := "  2) Collect per-function text profiles…"
	if got != want {
		t.Fatalf("formatProgressLabel() = %q, want %q", got, want)
	}
}

func TestFormatProgressLabel_prepareUnindented(t *testing.T) {
	t.Parallel()

	got := formatProgressLabel(Progress{Phase: PhasePrepare}, true)
	if got != "Preparing…" {
		t.Fatalf("prepare label = %q", got)
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
	s := newSessionForTest(&buf)
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
	if !strings.Contains(out, "✓") {
		t.Fatalf("output missing done marker: %q", out)
	}
}

func TestSession_PreparingThenBenchmark(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := newSessionForTest(&buf)
	err := s.RunWhile(Progress{Phase: PhasePrepare}, func() error {
		s.Warn("no prof.json")
		return nil
	})
	if err != nil {
		t.Fatalf("RunWhile() err = %v", err)
	}
	s.BeginBenchmark(1, 2, "BenchmarkFoo")
	out := buf.String()
	if !strings.Contains(out, "no prof.json") {
		t.Fatalf("missing prepare warning: %q", out)
	}
	if !strings.Contains(out, "Benchmark 1/2 · BenchmarkFoo") {
		t.Fatalf("missing benchmark header: %q", out)
	}
	// Finish resets column after width-padded overwrite so the benchmark header is left-aligned.
	if !strings.Contains(out, "\r") {
		t.Fatalf("expected column reset after warned prepare stage: %q", out)
	}
}

func TestPrintWarning_writesStyledLine(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	PrintWarning(&buf, ConfigureDetailPrefix, "oops")
	if !strings.Contains(buf.String(), "warning:") {
		t.Fatalf("output = %q", buf.String())
	}
}

func TestStepGap_writesBlankLine(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	StepGap(&buf)
	if buf.String() != "\n" {
		t.Fatalf("StepGap() = %q, want newline", buf.String())
	}
}

func TestPrintSection_writesTitleAndRule(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	PrintSection(&buf, -1, SurveySectionTitle)
	out := buf.String()
	if !strings.HasPrefix(out, "\n") {
		t.Fatalf("expected leading blank line: %q", out)
	}
	if !strings.Contains(out, SurveySectionTitle) {
		t.Fatalf("missing title: %q", out)
	}
	if !strings.Contains(out, "─") {
		t.Fatalf("missing rule: %q", out)
	}
}

func TestSession_BeginCollect_printsSectionBreak(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := newSessionForTest(&buf)
	s.BeginCollect()
	err := s.RunWhile(Progress{Phase: PhasePrepare}, func() error { return nil })
	if err != nil {
		t.Fatalf("RunWhile() err = %v", err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "\n") {
		t.Fatalf("expected leading blank line, got %q", out)
	}
	if !strings.Contains(out, CollectSectionTitle) {
		t.Fatalf("missing section title: %q", out)
	}
	if !strings.Contains(out, "─") {
		t.Fatalf("missing section rule: %q", out)
	}
	titleIdx := strings.Index(out, CollectSectionTitle)
	prepIdx := strings.Index(out, "Preparing")
	if titleIdx < 0 || prepIdx < 0 || prepIdx <= titleIdx {
		t.Fatalf("Preparing should follow section title: %q", out)
	}
}

func TestSession_BeginBenchmark_leadingSeparator(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := newSessionForTest(&buf)
	s.BeginBenchmark(1, 2, "BenchmarkFoo")
	if !strings.HasPrefix(buf.String(), "\n") {
		t.Fatalf("expected leading blank line before benchmark header, got %q", buf.String())
	}
}

func TestSession_BeginBenchmark_printsSectionHeader(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := newSessionForTest(&buf)
	s.BeginBenchmark(1, 2, "BenchmarkFoo")
	s.BeginBenchmark(2, 2, "BenchmarkBar")
	out := buf.String()
	if !strings.Contains(out, "Benchmark 1/2 · BenchmarkFoo") {
		t.Fatalf("missing first benchmark header: %q", out)
	}
	if !strings.Contains(out, "Benchmark 2/2 · BenchmarkBar") {
		t.Fatalf("missing second benchmark header: %q", out)
	}
}

func TestSession_Warn_indentOutsideStageFallsBackToSlog(t *testing.T) {
	t.Parallel()

	s := newSessionForTest(io.Discard)
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

func TestSession_RunWhile_failedStageShowsErrorDetail(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := newSessionForTest(&buf)
	runErr := errors.New("setup failed")
	err := s.RunWhile(Progress{Phase: PhasePrepare}, func() error {
		return runErr
	})
	if !errors.Is(err, runErr) {
		t.Fatalf("RunWhile() err = %v, want %v", err, runErr)
	}
	out := buf.String()
	if !strings.Contains(out, "✗") {
		t.Fatalf("output missing failure marker: %q", out)
	}
	if !strings.Contains(out, "error:") {
		t.Fatalf("output missing error detail: %q", out)
	}
	if !strings.Contains(out, "setup failed") {
		t.Fatalf("output missing error message: %q", out)
	}
	if !detailBelowStageHeader(out, "Preparing", "error:") {
		t.Fatalf("error should be below Preparing header: %q", out)
	}
}

func TestSession_RunWhile_failedAfterWarnings(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := newSessionForTest(&buf)
	err := s.RunWhile(Progress{Phase: PhaseCollectProfiles, Detail: "cpu"}, func() error {
		s.Warn("png skipped")
		return errors.New("profile text failed")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	out := buf.String()
	if !strings.Contains(out, "warning:") || !strings.Contains(out, "error:") {
		t.Fatalf("expected warning and error lines: %q", out)
	}
	if !strings.Contains(out, "✗") {
		t.Fatalf("expected failure marker: %q", out)
	}
}

func TestSession_RunWhile_stageErrorsPerPhase(t *testing.T) {
	t.Parallel()

	cases := []struct {
		phase Phase
		label string
	}{
		{PhaseRunBenchmark, "0) Run benchmark"},
		{PhaseCollectProfiles, "1) Collect profiles"},
		{PhaseCollectFunctionProfiles, "2) Collect per-function text profiles"},
	}
	for _, tc := range cases {
		t.Run(string(tc.phase), func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			s := newSessionForTest(&buf)
			err := s.RunWhile(Progress{Phase: tc.phase}, func() error {
				return errors.New("phase failed")
			})
			if err == nil {
				t.Fatal("expected error")
			}
			out := buf.String()
			if !strings.Contains(out, tc.label) {
				t.Fatalf("missing label %q in %q", tc.label, out)
			}
			if !detailBelowStageHeader(out, tc.label, "error:") {
				t.Fatalf("error not below header for %s: %q", tc.phase, out)
			}
		})
	}
}

func TestSession_RunWhile_stage2Warnings(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := newSessionForTest(&buf)
	err := s.RunWhile(Progress{Phase: PhaseCollectFunctionProfiles}, func() error {
		s.Warn("skipping Foo")
		return nil
	})
	if err != nil {
		t.Fatalf("RunWhile() err = %v", err)
	}
	out := buf.String()
	if !detailBelowStageHeader(out, "2) Collect per-function text profiles", "warning:") {
		t.Fatalf("stage 2 warning not below header: %q", out)
	}
}

func TestErrorWasStaged(t *testing.T) {
	t.Parallel()

	base := errors.New("boom")
	if ErrorWasStaged(base) {
		t.Fatal("expected false for plain error")
	}
	wrapped := StagedDisplay(base)
	if !ErrorWasStaged(wrapped) {
		t.Fatal("expected true for staged error")
	}
}

func TestSession_ErrorDisplayed(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := newSessionForTest(&buf)
	if s.ErrorDisplayed() {
		t.Fatal("expected false before run")
	}
	_ = s.RunWhile(Progress{Phase: PhasePrepare}, func() error {
		return errors.New("boom")
	})
	if !s.ErrorDisplayed() {
		t.Fatal("expected true after staged error")
	}
	_ = buf.String()
}

func detailBelowStageHeader(out, stage, prefix string) bool {
	stageIdx := strings.Index(out, stage)
	if stageIdx < 0 {
		return false
	}
	detailIdx := strings.Index(out[stageIdx:], prefix)
	if detailIdx < 0 {
		return false
	}
	return strings.Contains(out[stageIdx:stageIdx+detailIdx], "\n")
}

func TestSession_RunWhile_warnedSuccessShowsCountSuffix(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := newSessionForTest(&buf)
	err := s.RunWhile(Progress{Phase: PhaseCollectProfiles, Detail: "cpu, memory"}, func() error {
		s.Warn("first")
		s.Warn("second")
		return nil
	})
	if err != nil {
		t.Fatalf("RunWhile() err = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "2 warnings") {
		t.Fatalf("expected warning count suffix: %q", out)
	}
}

func TestSession_Warn_truncatesOnNarrowTerminal(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := newNarrowSessionForTest(&buf, 36)
	longMsg := strings.Repeat("x", 80)
	err := s.RunWhile(Progress{Phase: PhasePrepare}, func() error {
		s.Warn(longMsg)
		return nil
	})
	if err != nil {
		t.Fatalf("RunWhile() err = %v", err)
	}
	out := buf.String()
	if strings.Count(out, longMsg) > 0 {
		t.Fatalf("expected truncated message, got full repeat in %q", out)
	}
	if !strings.Contains(out, "✓") {
		t.Fatalf("expected success marker after truncate: %q", out)
	}
}

func TestSession_WarningsStayBelowHeaderWithMultipleWarns(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := newSessionForTest(&buf)
	err := s.RunWhile(Progress{
		Phase:  PhaseCollectProfiles,
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
	if !warningBelowStageHeader(out, "1) Collect profiles") {
		t.Fatalf("warnings not below header: %q", out)
	}
	first := strings.Index(out, "first warn")
	second := strings.Index(out, "second warn")
	if first < 0 || second < 0 || second <= first {
		t.Fatalf("expected two warning lines in order: %q", out)
	}
}
