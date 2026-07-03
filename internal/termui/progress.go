package termui

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"
)

// Phase identifies a user-visible collect pipeline step.
type Phase string

const (
	// PhasePrepare covers setup and prelude warnings before benchmarks run.
	PhasePrepare Phase = "prepare"
	// PhaseRunBenchmark covers go test, bench text write, and profile binary move.
	PhaseRunBenchmark Phase = "run_benchmark"
	// PhaseCollectProfiles covers pprof text and PNG generation.
	PhaseCollectProfiles Phase = "collect_profiles"
	// PhaseCollectFunctionProfiles covers parser extraction and per-function pprof lists.
	PhaseCollectFunctionProfiles Phase = "collect_function_profiles"
)

const defaultTermWidth = 120

// Progress describes a long-running step for user-facing labels.
type Progress struct {
	Phase  Phase
	Label  string // benchmark name
	Index  int    // 1-based benchmark index
	Total  int    // benchmark count
	Detail string // optional detail, e.g. "count=5" or "cpu, memory"
}

// WithPhase returns a copy of p with the given phase.
func (p Progress) WithPhase(phase Phase) Progress {
	p.Phase = phase
	return p
}

// WithDetail returns a copy of p with the given detail suffix.
func (p Progress) WithDetail(detail string) Progress {
	p.Detail = detail
	return p
}

// Session drives TTY progress UI for a collect run. Nil is non-interactive.
//
// Interactive layout:
//
//	✓ Preparing
//	    warning: …
//
//	Benchmark 1/2 · BenchmarkStringProcessor
//	  ✓ 0) Run benchmark (count=5)
//	  ✓ 1) Collect profiles (cpu, memory)
//	      warning: …
//	  ✓ 2) Collect per-function text profiles
//
// The step line updates in place while running (spinner), then becomes ✓ when done.
type Session struct {
	w           io.Writer
	fd          int
	interactive bool

	mu                sync.Mutex
	stageActive       bool
	headerLocked      bool
	warningCount      int
	runningLabel      string
	warnPrefix        string
	lastFrame         string
	spinnerStop       chan struct{}
	spinnerDone       sync.WaitGroup
	benchmarksStarted int
}

// NewSession reports whether w/fd is an interactive terminal.
func NewSession(w io.Writer, fd int) *Session {
	return &Session{
		w:           w,
		fd:          fd,
		interactive: term.IsTerminal(fd),
	}
}

// Interactive is true when the session can show spinners and styled lines.
func (s *Session) Interactive() bool {
	if s == nil {
		return false
	}
	return s.interactive
}

func formatProgressLabel(p Progress, running bool) string {
	linePrefix, _ := phasePrefixes(p.Phase)
	var b strings.Builder
	b.WriteString(linePrefix)
	switch p.Phase {
	case PhasePrepare:
		b.WriteString("Preparing")
	case PhaseRunBenchmark:
		b.WriteString("0) Run benchmark")
		if p.Detail != "" {
			fmt.Fprintf(&b, " (%s)", p.Detail)
		}
	case PhaseCollectProfiles:
		b.WriteString("1) Collect profiles")
		if p.Detail != "" {
			fmt.Fprintf(&b, " (%s)", p.Detail)
		}
	case PhaseCollectFunctionProfiles:
		b.WriteString("2) Collect per-function text profiles")
	}
	if running {
		b.WriteString("…")
	}
	return b.String()
}

func phasePrefixes(phase Phase) (linePrefix, warnPrefix string) {
	if phase == PhasePrepare {
		return "", "    "
	}
	return "  ", "      "
}

// BeginBenchmark prints a section header before the three steps for one benchmark.
func (s *Session) BeginBenchmark(index, total int, name string) {
	if s == nil || !s.interactive {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Fprintln(s.w)
	s.benchmarksStarted++

	var title string
	if total > 1 {
		title = fmt.Sprintf("Benchmark %d/%d · %s", index, total, name)
	} else {
		title = name
	}
	fmt.Fprintln(s.w, BenchmarkTitleStyle.Render(title))
}

var dotSpinner = spinner.Dot

// RunWhile runs fn while showing a persistent spinner when the session is interactive.
func (s *Session) RunWhile(p Progress, fn func() error) error {
	if fn == nil {
		return errors.New("termui: nil task")
	}
	if s == nil || !s.interactive {
		return fn()
	}

	s.mu.Lock()
	s.stageActive = true
	s.headerLocked = false
	s.warningCount = 0
	_, s.warnPrefix = phasePrefixes(p.Phase)
	s.runningLabel = formatProgressLabel(p, true)
	s.startSpinnerLocked()
	s.mu.Unlock()

	fnErr := fn()

	s.mu.Lock()
	s.signalSpinnerStopLocked()
	doneLabel := formatProgressLabel(p, false)
	failed := fnErr != nil
	s.mu.Unlock()

	s.spinnerDone.Wait()

	s.mu.Lock()
	s.finishStageLocked(doneLabel, failed)
	s.stageActive = false
	s.mu.Unlock()

	return fnErr
}

func (s *Session) startSpinnerLocked() {
	stop := make(chan struct{})
	s.spinnerStop = stop
	s.spinnerDone.Add(1)

	frames := dotSpinner.Frames
	s.lastFrame = frames[0]
	s.paintSpinnerHeaderLocked(s.lastFrame)

	go func() {
		defer s.spinnerDone.Done()
		ticker := time.NewTicker(dotSpinner.FPS)
		defer ticker.Stop()

		i := 1
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				s.mu.Lock()
				if !s.headerLocked {
					frame := frames[i%len(frames)]
					s.lastFrame = frame
					s.paintSpinnerHeaderLocked(frame)
				}
				s.mu.Unlock()
				i++
			}
		}
	}()
}

func (s *Session) paintSpinnerHeaderLocked(frame string) {
	content := spinnerFrameStyle.Render(frame) + " " + LabelStyle.Render(s.runningLabel)
	s.overwriteLineLocked(content, false)
}

func (s *Session) lockHeaderLocked() {
	if s.headerLocked {
		return
	}
	s.headerLocked = true
	// Commit the spinner line (erase + newline) so warnings append below a stable header.
	content := spinnerFrameStyle.Render(s.lastFrame) + " " + LabelStyle.Render(s.runningLabel)
	s.overwriteLineLocked(content, true)
}

func (s *Session) signalSpinnerStopLocked() {
	if s.spinnerStop == nil {
		return
	}
	close(s.spinnerStop)
	s.spinnerStop = nil
}

func (s *Session) finishStageLocked(doneLabel string, failed bool) {
	mark := DoneStyle.Render("✓")
	if failed {
		mark = FailStyle.Render("✗")
	}
	line := mark + " " + doneLabel

	if s.warningCount > 0 {
		s.seekStageHeaderLocked()
		s.overwriteLineLocked(line, false)
		s.seekAfterStageBlockLocked()
		return
	}

	s.overwriteLineLocked(line, true)
}

func (s *Session) seekStageHeaderLocked() {
	// Header line, then warningCount warning lines below; cursor starts after the block.
	n := s.warningCount + 1
	if n > 0 {
		fmt.Fprint(s.w, ansi.CursorUp(n))
	}
}

func (s *Session) seekAfterStageBlockLocked() {
	n := s.warningCount + 1
	if n > 0 {
		fmt.Fprint(s.w, ansi.CursorDown(n))
		// Width-padded overwrite leaves the cursor at the terminal edge; reset column
		// so the next write starts at the left margin.
		fmt.Fprint(s.w, "\r")
	}
}

func (s *Session) termWidth() int {
	_, width, err := term.GetSize(s.fd)
	if err != nil || width <= 0 {
		return defaultTermWidth
	}
	return width
}

func (s *Session) overwriteLineLocked(content string, newline bool) {
	width := s.termWidth()
	visible := lipgloss.Width(content)
	padding := width - visible
	if padding < 0 {
		padding = 0
	}
	fmt.Fprint(s.w, ansi.EraseEntireLine+"\r"+content+strings.Repeat(" ", padding))
	if newline {
		fmt.Fprint(s.w, "\n")
	}
}

func formatWarningLine(prefix, msg string) string {
	return WarningPrefixStyle.Render(prefix+"warning: ") + WarningStyle.Render(msg)
}

// Warn writes a styled warning on its own line below the active stage header.
func (s *Session) Warn(msg string) {
	if s == nil || !s.interactive {
		slog.Warn(msg)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.stageActive {
		slog.Warn(msg)
		return
	}

	s.lockHeaderLocked()
	fmt.Fprintln(s.w, formatWarningLine(s.warnPrefix, msg))
	s.warningCount++
}

// Success writes a completion message (styled on TTY, slog otherwise).
func (s *Session) Success(msg string) {
	if s == nil || !s.interactive {
		slog.Info(msg)
		return
	}
	fmt.Fprintln(s.w, SuccessStyle.Render(msg))
}

// newSessionForTest builds a session with a forced interactive flag (tests only).
func newSessionForTest(w io.Writer) *Session {
	return &Session{w: w, interactive: true}
}

var spinnerFrameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(AccentColor))
