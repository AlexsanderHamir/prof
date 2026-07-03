package termui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Progress describes a long-running step for user-facing labels.
type Progress struct {
	Label  string // benchmark name
	Index  int    // 1-based step index
	Total  int    // total steps in the run
	Detail string // optional detail, e.g. "count=5"
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
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(AccentColor))),
	)
	return spinnerModel{
		spinner: s,
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

// formatProgressLabel builds the human-readable label shown beside the spinner.
func formatProgressLabel(p Progress) string {
	var b strings.Builder
	if p.Total > 1 {
		fmt.Fprintf(&b, "Running benchmark %d/%d: ", p.Index, p.Total)
	} else {
		b.WriteString("Running benchmark: ")
	}
	b.WriteString(p.Label)
	if p.Detail != "" {
		fmt.Fprintf(&b, " (%s)", p.Detail)
	}
	b.WriteString("…")
	return b.String()
}

// RunWhile runs fn while showing a spinner on w when fd is a terminal.
// When fd is not a terminal, fn runs with no UI overhead.
func RunWhile(w io.Writer, fd int, p Progress, fn func() error) error {
	if fn == nil {
		return fmt.Errorf("termui: nil task")
	}
	if !term.IsTerminal(fd) {
		return fn()
	}

	label := formatProgressLabel(p)
	doneCh := make(chan struct{})

	model := newSpinnerModel(label)
	prog := tea.NewProgram(
		model,
		tea.WithOutput(w),
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

// DoneLine writes a single faint completion line to w when fd is a terminal.
func DoneLine(w io.Writer, fd int, p Progress) {
	if !term.IsTerminal(fd) {
		return
	}
	fmt.Fprintln(w, FaintStyle.Render(p.Label+" finished"))
}
