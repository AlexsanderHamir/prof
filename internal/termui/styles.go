package termui

import "github.com/charmbracelet/lipgloss"

// AccentColor matches the prof ui hub selection highlight (internal/tui/hub.go).
const AccentColor = "39"

var (
	// LabelStyle styles in-progress task labels beside the spinner.
	LabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(AccentColor))
	// FaintStyle styles secondary completion text.
	FaintStyle = lipgloss.NewStyle().Faint(true)
	// WarningStyle styles user-facing warnings on interactive terminals.
	WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	// SuccessStyle styles the final success line on interactive terminals.
	SuccessStyle = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("42"))
)
