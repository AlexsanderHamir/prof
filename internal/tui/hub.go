package tui

import (
	"fmt"
	"strings"

	"github.com/AlexsanderHamir/prof/internal/workspace"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MainAction is the user's choice from the hub; the CLI dispatches to engines.
type MainAction int

const (
	// MainNone means no selection (should not be returned from RunMainMenu on success).
	MainNone MainAction = iota
	// MainQuit means the user exited without running a workflow.
	MainQuit
	// MainCollect runs interactive benchmark collection.
	MainCollect
	// MainCompare runs interactive compare between two tags.
	MainCompare
	// MainTools opens the tools submenu (benchstat, qcachegrind).
	MainTools
	// MainSetup is a deprecated alias for MainConfig.
	MainSetup
	// MainConfig creates prof.json when missing.
	MainConfig
	// MainDocs prints the documentation URL only.
	MainDocs
)

type mainItem struct {
	label  string
	action MainAction
}

// hubModel is the Bubble Tea model for the prof main menu.
type hubModel struct {
	cursor   int
	result   MainAction
	items    []mainItem
	showHelp bool
}

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	selStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	normalStyle = lipgloss.NewStyle()
	footerStyle = lipgloss.NewStyle().Faint(true).MarginTop(1)
	helpStyle   = lipgloss.NewStyle().Faint(true).MarginTop(1).PaddingLeft(2)
)

func newHubModel() *hubModel {
	return &hubModel{
		cursor: 0,
		result: MainNone,
		items: []mainItem{
			{"Run benchmarks and save profiles (pick a name for this run)", MainCollect},
			{"Built-in regression check — compare two runs, function by function", MainCompare},
			{fmt.Sprintf("External tools — %s or %s", workspace.ToolNameBenchstat, workspace.ToolNameQcachegrind), MainTools},
			{"Create prof.json — starter config + prof.json.example docs", MainConfig},
			{"Help — print link to online documentation", MainDocs},
			{"Quit", MainQuit},
		},
	}
}

// RunMainMenu runs the full-screen hub until the user selects an action or quits.
func RunMainMenu() (MainAction, error) {
	p := tea.NewProgram(newHubModel(), tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return MainNone, err
	}
	fm, ok := final.(*hubModel)
	if !ok {
		return MainNone, fmt.Errorf("internal error: unexpected model type %T", final)
	}
	if fm.result == MainNone {
		return MainQuit, nil
	}
	return fm.result, nil
}

func (m *hubModel) Init() tea.Cmd {
	return nil
}

func (m *hubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
	case "ctrl+c", "q":
		m.result = MainQuit
		return m, tea.Quit

	case "esc":
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		m.result = MainQuit
		return m, tea.Quit

	case "?":
		m.showHelp = !m.showHelp
		return m, nil

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "down", "j":
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
		return m, nil

	case "enter":
		m.result = m.items[m.cursor].action
		return m, tea.Quit
	}

	return m, nil
}

func (m *hubModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Prof — main menu"))
	b.WriteString("\n\n")
	for i, it := range m.items {
		prefix := "  "
		style := normalStyle
		if i == m.cursor {
			prefix = "▸ "
			style = selStyle
		}
		b.WriteString(style.Render(prefix + it.label))
		b.WriteString("\n")
	}

	if m.showHelp {
		b.WriteString(helpStyle.Render(
			"Save profiles: runs benchmarks and stores output under bench/<name>/.\n" +
				"Built-in regression check: prof compares two saved runs and lists which functions got slower or faster; can fail the run using limits in prof.json (for CI).\n" +
				fmt.Sprintf("External tools: run %s or %s (separate programs, not prof's built-in regression check).\n", workspace.ToolNameBenchstat, workspace.ToolNameQcachegrind) +
				"Create prof.json: writes valid prof.json and prof.json.example (commented reference) beside go.mod.\n" +
				"Press ? again to hide this help.",
		))
		b.WriteString("\n")
	}

	b.WriteString(footerStyle.Render("↑/↓ move · enter select · ? help · esc/q quit"))
	return b.String()
}
