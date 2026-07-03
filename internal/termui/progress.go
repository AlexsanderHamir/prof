package termui

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
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

	mu            sync.Mutex
	stageActive   bool
	stageWarnings atomic.Int32
	spinnerStop   chan struct{}
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

	runningLabel := formatProgressLabel(p, true)
	doneLabel := formatProgressLabel(p, false)

	s.mu.Lock()
	s.stageActive = true
	s.stageWarnings.Store(0)
	s.startSpinner(runningLabel)
	s.mu.Unlock()

	fnErr := fn()

	s.mu.Lock()
	s.finishStage(doneLabel, fnErr != nil)
	s.stageActive = false
	s.stageWarnings.Store(0)
	s.mu.Unlock()

	return fnErr
}

func (s *Session) startSpinner(label string) {
	stop := make(chan struct{})
	s.spinnerStop = stop

	styled := LabelStyle.Render(label)
	frameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(AccentColor))
	frames := dotSpinner.Frames

	go func() {
		ticker := time.NewTicker(dotSpinner.FPS)
		defer ticker.Stop()

		i := 0
		writeHeader := func() {
			frame := frameStyle.Render(frames[i%len(frames)])
			linesUp := int(s.stageWarnings.Load())
			if linesUp > 0 {
				fmt.Fprintf(s.w, "\033[%dF", linesUp)
			}
			fmt.Fprintf(s.w, "\r%s %s", frame, styled)
			if linesUp > 0 {
				fmt.Fprintf(s.w, "\033[%dE", linesUp)
			}
		}

		writeHeader()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				i++
				writeHeader()
			}
		}
	}()
}

func (s *Session) finishStage(doneLabel string, failed bool) {
	if s.spinnerStop != nil {
		close(s.spinnerStop)
		s.spinnerStop = nil
	}

	mark := DoneStyle.Render("✓")
	if failed {
		mark = FailStyle.Render("✗")
	}

	linesUp := int(s.stageWarnings.Load())
	if linesUp > 0 {
		fmt.Fprintf(s.w, "\033[%dF", linesUp)
	}
	fmt.Fprintf(s.w, "\r%s %s\n", mark, doneLabel)
}

// Warn writes a styled warning indented under the active stage on interactive terminals.
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

	fmt.Fprintln(s.w, WarningStyle.Render("    warning: "+msg))
	s.stageWarnings.Add(1)
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
func newSessionForTest(w io.Writer, interactive bool) *Session {
	return &Session{w: w, interactive: interactive}
}
