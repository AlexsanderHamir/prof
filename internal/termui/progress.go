package termui

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Phase identifies a user-visible collect pipeline step.
type Phase string

const (
	// PhaseRunBenchmark covers go test, bench text write, and profile binary move.
	PhaseRunBenchmark Phase = "run_benchmark"
	// PhaseCollectProfiles covers pprof text and PNG generation.
	PhaseCollectProfiles Phase = "collect_profiles"
	// PhaseAnalyzeProfiles covers parser extraction and per-function pprof lists.
	PhaseAnalyzeProfiles Phase = "analyze_profiles"
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

// Session drives TTY progress UI for a collect run. Zero value is non-interactive.
type Session struct {
	w           io.Writer
	fd          int
	interactive bool
}

// NewSession reports whether w/fd is an interactive terminal.
func NewSession(w io.Writer, fd int) Session {
	return Session{
		w:           w,
		fd:          fd,
		interactive: term.IsTerminal(fd),
	}
}

// Interactive is true when the session can show spinners and styled lines.
func (s Session) Interactive() bool {
	return s.interactive
}

type workDoneMsg struct {
	err error
}

type spinnerModel struct {
	spinner spinner.Model
	label   string
	workErr error
}

func newSpinnerModel(label string) spinnerModel {
	sp := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(AccentColor))),
	)
	return spinnerModel{
		spinner: sp,
		label:   LabelStyle.Render(label),
	}
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case workDoneMsg:
		m.workErr = msg.err
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m spinnerModel) View() string {
	return m.spinner.View() + " " + m.label
}

func formatProgressLabel(p Progress) string {
	var b strings.Builder
	switch p.Phase {
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
	case PhaseAnalyzeProfiles:
		b.WriteString("Analyzing profiles for ")
		b.WriteString(p.Label)
	}
	b.WriteString("…")
	return b.String()
}

// RunWhile runs fn while showing a spinner when the session is interactive.
func (s Session) RunWhile(p Progress, fn func() error) error {
	if fn == nil {
		return errors.New("termui: nil task")
	}
	if !s.interactive {
		return fn()
	}

	label := formatProgressLabel(p)
	doneCh := make(chan struct{})

	model := newSpinnerModel(label)
	prog := tea.NewProgram(
		model,
		tea.WithOutput(s.w),
		tea.WithInput(nil),
		tea.WithoutSignalHandler(),
	)

	go func() {
		_, _ = prog.Run()
		close(doneCh)
	}()

	fnErr := fn()
	prog.Send(workDoneMsg{err: fnErr})
	<-doneCh

	return fnErr
}

// Warn writes a styled warning line on interactive terminals.
func (s Session) Warn(msg string) {
	if !s.interactive {
		slog.Warn(msg)
		return
	}
	fmt.Fprintln(s.w, WarningStyle.Render("warning: "+msg))
}

// Success writes a completion message (styled on TTY, slog otherwise).
func (s Session) Success(msg string) {
	if !s.interactive {
		slog.Info(msg)
		return
	}
	fmt.Fprintln(s.w, SuccessStyle.Render(msg))
}
