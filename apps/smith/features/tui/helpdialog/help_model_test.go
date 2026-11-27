package helpdialog

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
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
	if text := strings.Join(h.Lines(th), "\n"); !strings.Contains(text, "Ctrl+R") || !strings.Contains(text, "Keys") {
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

// scrollbarOnTheBorder fails unless the lines draw a scrollbar and every cell
// of its track is the last one inside the right border, as in every pane.
func scrollbarOnTheBorder(t *testing.T, lines []string) {
	t.Helper()
	track := []rune(theme.Default().Icons.Track)[0]
	bars := 0
	for _, l := range lines {
		r := []rune(pane.Strip(l))
		for x, c := range r {
			if c != track {
				continue
			}
			if x != len(r)-2 {
				t.Fatalf("the scrollbar is in column %d, not next to the border in column %d: %q", x, len(r)-2, string(r))
			}
			bars++
		}
	}
	if bars == 0 {
		t.Fatalf("the test needs a body that scrolls:\n%s", pane.Strip(strings.Join(lines, "\n")))
	}
}

func TestTheHelpScrollbarSitsNextToTheBorder(t *testing.T) {
	scrollbarOnTheBorder(t, sampleHelp().Lines(theme.Default()))
}

// scoped is a help with keys of the open view and keys of two menus.
func scoped(sw, sh int) Help {
	rows := []HelpRow{
		{Group: "Moving", Title: "Move in a list", Shortcut: "Up Down / j k"},
		{Group: "Run", Title: "Build all", Shortcut: "B"},
		{Group: "Run", Title: "Stop", Shortcut: "Ctrl+C", Disabled: "nothing is running"},
		{Group: "Settings view", Title: "Reset a setting to its default", Shortcut: "u", View: true},
	}
	return NewHelp(rows, []string{"Click selects"}).Titled("Settings").Place(sw, sh)
}

func TestTheHelpGroupsTheKeysByScopeTheOpenViewFirst(t *testing.T) {
	h := scoped(150, 44)
	text := pane.Strip(strings.Join(h.Lines(theme.Default()), "\n"))
	view, everywhere := strings.Index(text, "Settings view"), strings.Index(text, "Everywhere")
	if !strings.Contains(text, "╭─ Keys · Settings ") || view < 0 || everywhere < view {
		t.Fatalf("the title names the view and its keys come before the ones of everywhere:\n%s", text)
	}
	for _, want := range []string{"u  Reset a setting to its default", "Moving", "Run", "B              Build all", "Disabled: nothing is running",
		"Every action is a smith command", "Plain mode", "Mouse", "Find a command by name:", "Esc close"} {
		if !strings.Contains(text, want) {
			t.Fatalf("the help lacks %q:\n%s", want, text)
		}
	}
}

func TestTheFootStaysWhileTheKeysScroll(t *testing.T) {
	h := sampleHelp()
	th := theme.Default()
	total := h.Total(th)
	for _, key := range []string{"", "pgdown", "end"} {
		if key != "" {
			h, _ = h.Key(key, total)
		}
		lines := h.Lines(th)
		if foot := pane.Strip(lines[len(lines)-2]); !strings.Contains(foot, "Find a command by name:") || !strings.Contains(foot, "Esc close") {
			t.Fatalf("after %q the foot is not pinned over the border: %q", key, foot)
		}
	}
	if h.top == 0 {
		t.Fatal("end scrolled the keys")
	}
}

func TestTheNarrowHelpFitsItsPlace(t *testing.T) {
	for _, sz := range [][2]int{{80, 30}, {120, 40}, {150, 44}} {
		h := scoped(sz[0], sz[1]).Within(3, 2).Place(sz[0], sz[1])
		r := h.Rect()
		if r.X < 0 || r.Y < 3 || r.X+r.W > sz[0] || r.Y+r.H > sz[1]-2 {
			t.Fatalf("%v: the help at %+v leaves the screen or its margins", sz, r)
		}
		lines := h.Lines(theme.Default())
		if len(lines) != r.H {
			t.Fatalf("%v: %d lines for %d rows", sz, len(lines), r.H)
		}
		for i, l := range lines {
			if pane.Width(l) != r.W {
				t.Fatalf("%v: line %d is %d cells, want %d", sz, i, pane.Width(l), r.W)
			}
		}
	}
}
