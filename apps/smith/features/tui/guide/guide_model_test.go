package guide

import (
	tea "github.com/charmbracelet/bubbletea"
)

func keys(m tea.Model, ks ...string) tea.Model {
	for _, k := range ks {
		m, _ = m.Update(keyMsg(k))
	}
	return m
}

func keyMsg(k string) tea.KeyMsg {
	switch k {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "space":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func menu() Model {
	return NewModel(Actions([]string{"go", "node"}, []string{"go", "node", "python"}, []string{"go", "node"})[1:])
}
