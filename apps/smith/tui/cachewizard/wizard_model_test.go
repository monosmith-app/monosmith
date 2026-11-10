package cachewizard

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSecretPressMapsEnterEscAndLetters(t *testing.T) {
	for name, tc := range map[string]struct {
		msg  tea.KeyMsg
		want rune
	}{
		"enter":  {tea.KeyMsg{Type: tea.KeyEnter}, '\r'},
		"esc":    {tea.KeyMsg{Type: tea.KeyEsc}, 27},
		"ctrl+c": {tea.KeyMsg{Type: tea.KeyCtrlC}, 3},
		"letter": {tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}, 'h'},
		"other":  {tea.KeyMsg{Type: tea.KeyTab}, 1},
	} {
		if got := secretPress(tc.msg).Press; got != tc.want {
			t.Errorf("%s: %q, want %q", name, got, tc.want)
		}
	}
}
