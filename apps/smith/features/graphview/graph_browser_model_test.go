package graphview

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func press(b Browser, keys ...string) Browser {
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k {
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case "right":
			msg = tea.KeyMsg{Type: tea.KeyRight}
		case "left":
			msg = tea.KeyMsg{Type: tea.KeyLeft}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		next, _ := b.Update(msg)
		b = next.(Browser)
	}
	return b
}

func TestBrowserWalksTheGraph(t *testing.T) {
	b := NewBrowser(Nodes(chain(t), map[string]bool{"web": true}))
	if b.Selected() != "sdk" {
		t.Fatalf("starts at the first project: %s", b.Selected())
	}
	view := b.View()
	for _, want := range []string{"Projects", "Projects it needs", "  ui", "Projects that use it", "  web", "q quit"} {
		if !strings.Contains(view, want) {
			t.Errorf("view lacks %q:\n%s", want, view)
		}
	}
	if b = press(b, "right"); b.Selected() != "ui" {
		t.Fatalf("→ goes to what it depends on: %s", b.Selected())
	}
	if b = press(b, "d"); b.Selected() != "sdk" {
		t.Fatalf("d goes to what uses it: %s", b.Selected())
	}
	if b = press(b, "left", "left"); b.Selected() != "sdk" {
		t.Fatalf("← retraces: %s", b.Selected())
	}
	if b = press(b, "down", "down", "down", "down"); b.Selected() != "sdk" {
		t.Fatalf("↓ wraps: %s", b.Selected())
	}
	if _, cmd := b.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}); cmd == nil {
		t.Fatal("q quits")
	}
}
