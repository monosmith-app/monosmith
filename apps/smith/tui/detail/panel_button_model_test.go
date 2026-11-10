package detail

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/muesli/termenv"
)

func TestAButtonActionShowsItsKeyAndHintAndEveryLineClicks(t *testing.T) {
	p := Panel{
		Actions: []Action{{ID: "fix", Label: "Fix", Shortcut: "f", Hint: "regenerate the pipelines"}},
		After:   []Section{{Heading: "How to fix", Items: []string{"a", "b", "c"}, Collapse: 1, Tail: []string{"or in a terminal: smith doctor"}}},
	}
	lines, hits := p.View(theme.Default(), 40, "")
	text := pane.Strip(strings.Join(lines, "\n"))
	for _, want := range []string{"Fix (f)", "regenerate the pipelines", "+2 more", "or in a terminal: smith doctor"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in\n%s", want, text)
		}
	}
	if len(hits) != 2 || hits[1] != "fix" {
		t.Fatalf("the button and its hint click the action: %v", hits)
	}
}

func TestAButtonActionLooksDifferentWhenHoveredOrDisabled(t *testing.T) {
	// Draw escape codes even though the test has no terminal, so the states can be told apart.
	defer lipgloss.SetColorProfile(lipgloss.ColorProfile())
	lipgloss.SetColorProfile(termenv.TrueColor)
	th := theme.Default()
	p := Panel{Actions: []Action{{ID: "fix", Label: "Fix", Shortcut: "f"}}}
	rest, _ := p.View(th, 40, "")
	hover, _ := p.View(th, 40, "fix")
	p.Actions[0].Disabled = true
	off, _ := p.View(th, 40, "fix")
	if rest[1] == hover[1] || off[1] == hover[1] || off[1] == rest[1] {
		t.Fatalf("rest, hover and disabled must differ: %q %q %q", rest[1], hover[1], off[1])
	}
}

func TestAPressedButtonIsActiveOnlyWhileThePointerIsOverIt(t *testing.T) {
	defer lipgloss.SetColorProfile(lipgloss.ColorProfile())
	lipgloss.SetColorProfile(termenv.TrueColor)
	th := theme.Default()
	p := Panel{Actions: []Action{{ID: "a", Label: "Alpha", Shortcut: "a"}, {ID: "b", Label: "Beta", Shortcut: "b"}}, Pressed: "a"}
	hover, _ := Panel{Actions: p.Actions}.View(th, 40, "a")
	active, _ := p.View(th, 40, "a")
	away, _ := p.View(th, 40, "b")
	rest, _ := Panel{Actions: p.Actions}.View(th, 40, "")
	if active[1] == hover[1] || active[1] == rest[1] {
		t.Fatalf("a held button differs from hover and rest: %q", active[1])
	}
	if away[1] != rest[1] {
		t.Fatalf("a press dragged off its button shows the button at rest: %q vs %q", away[1], rest[1])
	}
}

func TestEveryActionShowsItsKeyAndTheCursorRowLeadsWithTheMark(t *testing.T) {
	th := theme.Default()
	p := Panel{Actions: []Action{{ID: "a", Label: "Alpha", Shortcut: "a", Hint: "does alpha"}, {ID: "b", Label: "Beta", Shortcut: "b"}}, Cursor: "b"}
	lines, hits := p.View(th, 40, "")
	text := pane.Strip(strings.Join(lines, "\n"))
	for _, want := range []string{"Alpha (a)", "Beta (b)", "does alpha"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in\n%s", want, text)
		}
	}
	mark := strings.TrimRight(th.Icons.Pointer, " ")
	for row, id := range hits {
		line := pane.Strip(lines[row])
		if strings.Contains(line, "(") && strings.Contains(line, mark) != (id == "b") {
			t.Fatalf("only the cursor row has the mark %q: %q", mark, line)
		}
	}
}
