package devservers

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestChooser(t *testing.T) {
	var m tea.Model = newChooser([]string{"api", "web", "worker"})
	space := tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	for _, k := range []tea.KeyMsg{space, {Type: tea.KeyDown}, {Type: tea.KeyDown}, space, {Type: tea.KeyEnter}} {
		m, _ = m.Update(k)
	}
	if got := m.(chooser).Chosen(); len(got) != 2 || got[0] != "api" || got[1] != "worker" {
		t.Fatalf("chosen %v", got)
	}
	m = newChooser([]string{"api"})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.(chooser).Chosen() != nil {
		t.Fatal("cancelled starts nothing")
	}
}
