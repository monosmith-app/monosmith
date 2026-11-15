package dependencies

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPickerStartsWithoutMajorsAndToggles(t *testing.T) {
	updates := []Update{{Name: "chalk", Level: "patch"}, {Name: "zod", Level: "minor"}, {Name: "jest", Level: "major"}}
	var m tea.Model = newPicker(updates)
	press := func(msgs ...tea.KeyMsg) {
		for _, msg := range msgs {
			m, _ = m.Update(msg)
		}
	}
	space := tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	press(tea.KeyMsg{Type: tea.KeyDown}, space, tea.KeyMsg{Type: tea.KeyDown}, space, tea.KeyMsg{Type: tea.KeyEnter})
	var names []string
	for _, u := range m.(picker).Chosen() {
		names = append(names, u.Name)
	}
	if len(names) != 2 || names[0] != "chalk" || names[1] != "jest" {
		t.Fatalf("chosen %v", names)
	}
	m = newPicker(updates)
	press(tea.KeyMsg{Type: tea.KeyEsc})
	if m.(picker).Chosen() != nil {
		t.Fatal("cancelled means nothing")
	}
}
