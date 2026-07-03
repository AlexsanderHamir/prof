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
type Session struct {
	w           io.Writer
	fd          int
	interactive bool

	mu              sync.Mutex
	stageActive     bool
	headerFrozen    bool
	warningCount    int
	completedStages int
	runningLabel    string
	lastFrame       string
	spinnerStop     chan struct{}
	spinnerDone     sync.WaitGroup
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
	var b strings.Builder
	switch p.Phase {
	case PhasePrepare:
		b.WriteString("Preparing")
	case PhaseRunBenchmark:
		if p.Total > 1 {
			fmt.Fprintf(&b, "Running benchmark %d/%d: ", p.Index, p.Total)
		} else {
			b.WriteString("Running benchmark: ")
		}
		b.WriteString(p.Label)
		if p.Detail != "" {
			fmt.Fprintf(&b, " (%s)", p.Detail)
		}
	case PhaseCollectProfiles:
		b.WriteString("Collecting profiles for ")
		b.WriteString(p.Label)
		if p.Detail != "" {
			fmt.Fprintf(&b, " (%s)", p.Detail)
		}
	case PhaseCollectFunctionProfiles:
		b.WriteString("Collecting function profiles for ")
		b.WriteString(p.Label)
	}
	if running {
		b.WriteString("…")
	}
	return b.String()
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
	if s.completedStages > 0 {
		fmt.Fprintln(s.w)
	}
	s.stageActive = true
	s.headerFrozen = false
	s.warningCount = 0
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
	s.completedStages++
	s.mu.Unlock()

	return fnErr
}

func (s *Session) startSpinnerLocked() {
	stop := make(chan struct{})
	s.spinnerStop = stop
	s.spinnerDone.Add(1)

	frames := dotSpinner.Frames
	s.lastFrame = frames[0]
	s.paintSpinnerFrameLocked(s.lastFrame)

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
				if !s.headerFrozen {
					frame := frames[i%len(frames)]
					s.lastFrame = frame
					s.paintSpinnerFrameLocked(frame)
				}
				s.mu.Unlock()
				i++
			}
		}
	}()
}

func (s *Session) paintSpinnerFrameLocked(frame string) {
	frameStyled := spinnerFrameStyle.Render(frame)
	labelStyled := LabelStyle.Render(s.runningLabel)
	fmt.Fprint(s.w, ansi.EraseEntireLine+"\r"+frameStyled+" "+labelStyled)
}

func (s *Session) freezeHeaderLocked() {
	if s.headerFrozen {
		return
	}
	s.headerFrozen = true
	frameStyled := spinnerFrameStyle.Render(s.lastFrame)
	labelStyled := LabelStyle.Render(s.runningLabel)
	fmt.Fprint(s.w, ansi.EraseEntireLine+"\r"+frameStyled+" "+labelStyled+"\n")
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

	if s.warningCount == 0 {
		fmt.Fprint(s.w, ansi.EraseEntireLine+"\r"+line+"\n")
		return
	}

	if !s.headerFrozen {
		s.freezeHeaderLocked()
	}

	linesUp := s.warningCount
	fmt.Fprint(s.w, ansi.CursorUp(linesUp))
	fmt.Fprint(s.w, ansi.EraseEntireLine+"\r"+line)
	fmt.Fprint(s.w, ansi.CursorDown(linesUp))
	fmt.Fprint(s.w, "\n")
}

func formatWarningLine(msg string) string {
	return WarningPrefixStyle.Render("    warning: ") + WarningStyle.Render(msg)
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

	s.freezeHeaderLocked()
	fmt.Fprintln(s.w, formatWarningLine(msg))
	s.warningCount++
}

// Success writes a completion message (styled on TTY, slog otherwise).
func (s *Session) Success(msg string) {
	if s == nil || !s.interactive {
		slog.Info(msg)
		return
	}
	fmt.Fprintln(s.w)
	fmt.Fprintln(s.w, SuccessStyle.Render(msg))
}

// newSessionForTest builds a session with a forced interactive flag (tests only).
func newSessionForTest(w io.Writer, interactive bool) *Session {
	return &Session{w: w, interactive: interactive}
}

var spinnerFrameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(AccentColor))
