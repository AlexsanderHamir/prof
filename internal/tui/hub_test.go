package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHubSelectFirstItemWithEnter(t *testing.T) {
	m := newHubModel()
	var tm tea.Model = m
	tm, _ = tm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	hm := tm.(*hubModel)
	if hm.result != MainCollect {
		t.Fatalf("want MainCollect, got %v", hm.result)
	}
}

func TestHubMoveDownAndSelectCompare(t *testing.T) {
	m := newHubModel()
	var tm tea.Model = m
	tm, _ = tm.Update(tea.KeyMsg{Type: tea.KeyDown})
	hm := tm.(*hubModel)
	if hm.cursor != 1 {
		t.Fatalf("cursor want 1, got %d", hm.cursor)
	}
	tm, _ = tm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	hm = tm.(*hubModel)
	if hm.result != MainCompare {
		t.Fatalf("want MainCompare, got %v", hm.result)
	}
}

func TestHubQuitWithQ(t *testing.T) {
	m := newHubModel()
	var tm tea.Model = m
	tm, _ = tm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	hm := tm.(*hubModel)
	if hm.result != MainQuit {
		t.Fatalf("want MainQuit, got %v", hm.result)
	}
}

func TestHubToggleHelp(t *testing.T) {
	m := newHubModel()
	var tm tea.Model = m
	tm, _ = tm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	hm := tm.(*hubModel)
	if !hm.showHelp {
		t.Fatal("expected showHelp true")
	}
	tm, _ = tm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	hm = tm.(*hubModel)
	if hm.showHelp {
		t.Fatal("expected showHelp false")
	}
}

func TestHubViewNonEmpty(t *testing.T) {
	m := newHubModel()
	v := m.View()
	if v == "" {
		t.Fatal("empty view")
	}
}
