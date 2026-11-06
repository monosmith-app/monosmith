package dialog

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func sampleHelp() Help {
	var rows []HelpRow
	for i := 0; i < 40; i++ {
		rows = append(rows, HelpRow{Group: "Run", Title: "Command", Shortcut: "Ctrl+R"})
	}
	rows = append(rows, HelpRow{Group: "Help", Title: "Keyboard shortcuts", Shortcut: "F1"})
	return NewHelp(rows, []string{"Click selects, double-click runs"}).Place(100, 30)
}

func TestHelpListsShortcutsAndScrolls(t *testing.T) {
	h := sampleHelp()
	th := theme.Default()
	if text := strings.Join(h.Lines(th), "\n"); !strings.Contains(text, "Ctrl+R") || !strings.Contains(text, "Keyboard shortcuts") {
		t.Fatalf("\n%s", text)
	}
	total := h.Total(th)
	h, _ = h.Key("end", total)
	if text := strings.Join(h.Lines(th), "\n"); !strings.Contains(text, "double-click runs") {
		t.Fatalf("the end shows the mouse lines:\n%s", text)
	}
	h, _ = h.Key("home", total)
	if h.top != 0 {
		t.Fatal("home")
	}
}

func TestHelpClosesOnEscF1AndAnOutsideClick(t *testing.T) {
	h := sampleHelp()
	for _, k := range []string{"esc", "f1", "?"} {
		if _, o := h.Key(k, 10); !o.Done {
			t.Fatalf("%s should close", k)
		}
	}
	if !h.Click(0, 0).Done || h.Click(h.rect.X+5, h.rect.Y+5).Done {
		t.Fatal("outside closes, inside does not")
	}
}
