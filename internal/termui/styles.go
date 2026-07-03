package termui

import "github.com/charmbracelet/lipgloss"

// AccentColor matches the prof ui hub selection highlight (internal/tui/hub.go).
const AccentColor = "39"

var (
	// LabelStyle styles in-progress task labels beside the spinner.
	LabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(AccentColor))
	// FaintStyle styles secondary completion text.
	FaintStyle = lipgloss.NewStyle().Faint(true)
	// WarningStyle styles the warning message body on interactive terminals.
	WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	// WarningPrefixStyle styles the fixed warning prefix (indented, faint).
	WarningPrefixStyle = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("214"))
	// SuccessStyle styles the final success line on interactive terminals.
	SuccessStyle = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("42"))
	// DoneStyle styles the completion marker on finished stages.
	DoneStyle = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("42"))
	// FailStyle styles the failure marker on failed stages.
	FailStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	// BenchmarkTitleStyle styles the benchmark group header in the progress log.
	BenchmarkTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(AccentColor))
)
