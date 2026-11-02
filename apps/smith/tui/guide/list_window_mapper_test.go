package guide

import (
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// longList is a form with one multiple choice of 25 options and a glossary.
func longList(t *testing.T) Model {
	t.Helper()
	var opts []Option
	for i := range 25 {
		opts = append(opts, Option{Label: fmt.Sprintf("framework: opt%02d", i), Value: fmt.Sprintf("opt%02d", i), Hint: "a choice"})
	}
	step := Step{Key: "c", Prompt: "Build it from choices?", Multi: true, Optional: true, Glossary: true, Options: opts,
		Note: "stack = the set of tools.\nframework = how the app is written.\nlint = the tool that checks."}
	m, ok := NewModel([]Action{{Title: "Add", Steps: []Step{step}}}).Pick("Add")
	if !ok {
		t.Fatal("no action")
	}
	return m
}

func windowText(m Model, w, h int) (string, map[int]int) {
	lines, at := m.Window(theme.Default(), w, h)
	for i := range lines {
		lines[i] = pane.Strip(lines[i])
	}
	return strings.Join(lines, "\n"), at
}

func optionRows(text string) int { return strings.Count(text, "framework: opt") }

var screens = []struct {
	name string
	w, h int
}{
	{"110x32", 100, 22},
	{"80x24", 72, 14},
}

func TestTheWheelScrollsTheListAndTheSelectionStays(t *testing.T) {
	for _, sc := range screens {
		t.Run(sc.name, func(t *testing.T) {
			m := longList(t)
			first, _ := windowText(m, sc.w, sc.h)
			if !strings.Contains(first, "↓ ") || strings.Contains(first, "↑ ") {
				t.Fatalf("at the top only rows below are hidden:\n%s", first)
			}
			m = m.Wheel(6)
			text, _ := windowText(m, sc.w, sc.h)
			if m.cursor != 0 {
				t.Fatalf("the wheel moved the selection to %d", m.cursor)
			}
			if strings.Contains(text, "opt00") || !strings.Contains(text, "↑ ") || !strings.Contains(text, "↓ ") {
				t.Fatalf("the list did not scroll, or has no indicators:\n%s", text)
			}
		})
	}
}

func TestKeysMoveTheSelectionAndKeepItInView(t *testing.T) {
	for _, sc := range screens {
		t.Run(sc.name, func(t *testing.T) {
			m := longList(t)
			windowText(m, sc.w, sc.h)
			for range 12 {
				m = press(m, "down")
				windowText(m, sc.w, sc.h)
			}
			text, _ := windowText(m, sc.w, sc.h)
			if m.cursor != 12 || !strings.Contains(text, "opt12") {
				t.Fatalf("cursor %d, row 12 not in view:\n%s", m.cursor, text)
			}
			m = press(m, "end")
			text, _ = windowText(m, sc.w, sc.h)
			if m.cursor != 24 || !strings.Contains(text, "opt24") || strings.Contains(text, "↓ ") {
				t.Fatalf("End: cursor %d:\n%s", m.cursor, text)
			}
			m = press(m, "pgup")
			text, _ = windowText(m, sc.w, sc.h)
			if m.cursor >= 24 || m.cursor < 24-sc.h || !strings.Contains(text, fmt.Sprintf("opt%02d", m.cursor)) {
				t.Fatalf("PgUp: cursor %d:\n%s", m.cursor, text)
			}
			m = press(m, "home")
			text, _ = windowText(m, sc.w, sc.h)
			if m.cursor != 0 || !strings.Contains(text, "opt00") || strings.Contains(text, "↑ ") {
				t.Fatalf("Home: cursor %d:\n%s", m.cursor, text)
			}
			m = press(m, "pgdown")
			if m.cursor < 3 {
				t.Fatalf("PgDn moved to %d", m.cursor)
			}
		})
	}
}

func TestTheListAlwaysKeepsEightRowsAndTheGlossaryFoldsOnAShortScreen(t *testing.T) {
	for _, sc := range screens {
		t.Run(sc.name, func(t *testing.T) {
			text, _ := windowText(longList(t), sc.w, sc.h)
			if got := optionRows(text); got < 8 {
				t.Fatalf("the list got %d rows:\n%s", got, text)
			}
			// The pane decides: a short pane folds the words, a tall one keeps them.
			folded := strings.Contains(text, wordsHint)
			if (sc.h < 24) != folded || (folded && strings.Contains(text, "framework = ")) {
				t.Fatalf("a pane of %d rows: folded %v:\n%s", sc.h, folded, text)
			}
		})
	}
	// A tall screen shows the words, and "?" folds them again.
	m := longList(t)
	text, _ := windowText(m, 100, 30)
	if !strings.Contains(text, "framework = how the app") || optionRows(text) < 8 {
		t.Fatalf("a tall screen shows the words:\n%s", text)
	}
	// A short one opens them with ? and the list still keeps its rows.
	m = press(longList(t), "?")
	text, _ = windowText(m, 100, 22)
	if !strings.Contains(text, "framework = how the app") || optionRows(text) < 8 {
		t.Fatalf("? opens the words:\n%s", text)
	}
}

func TestAClickRowMapsToTheOptionUnderItAfterAScroll(t *testing.T) {
	m := longList(t).Wheel(5)
	text, at := windowText(m, 100, 22)
	lines := strings.Split(text, "\n")
	for row, opt := range at {
		if !strings.Contains(lines[row], fmt.Sprintf("opt%02d", opt)) {
			t.Fatalf("row %d is option %d but reads %q", row, opt, lines[row])
		}
	}
	if len(at) == 0 {
		t.Fatal("no option rows")
	}
}
