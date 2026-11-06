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
		Actions: []Action{{ID: "fix", Label: "Fix", Shortcut: "f", Hint: "regenerate the pipelines", Button: true}},
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
	p := Panel{Actions: []Action{{ID: "fix", Label: "Fix", Shortcut: "f", Button: true}}}
	rest, _ := p.View(th, 40, "")
	hover, _ := p.View(th, 40, "fix")
	p.Actions[0].Disabled = true
	off, _ := p.View(th, 40, "fix")
	if rest[1] == hover[1] || off[1] == hover[1] || off[1] == rest[1] {
		t.Fatalf("rest, hover and disabled must differ: %q %q %q", rest[1], hover[1], off[1])
	}
}
