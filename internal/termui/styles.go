package termui

import "github.com/charmbracelet/lipgloss"

// AccentColor matches the prof ui hub selection highlight (internal/tui/hub.go).
const AccentColor = "39"

var (
	// LabelStyle styles in-progress task labels beside the spinner.
	LabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(AccentColor))
	// FaintStyle styles completion lines after a task finishes.
	FaintStyle = lipgloss.NewStyle().Faint(true)
)
