package detail

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestAButtonActionShowsItsKeyAndHintAndEveryLineClicks(t *testing.T) {
	p := Panel{
		Actions: []Action{{ID: "fix", Label: "Fix", Shortcut: "f", Hint: "regenerate the pipelines", Button: true}},
		After:   []Section{{Heading: "How to fix", Items: []string{"a", "b", "c"}, Collapse: 1, Tail: []string{"or in a terminal: smith doctor"}}},
	}
	lines, hits := p.View(theme.Default(), 40, "")
	text := pane.Strip(strings.Join(lines, "\n"))
	for _, want := range []string{"[ Fix  f ]", "regenerate the pipelines", "+2 more", "or in a terminal: smith doctor"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in\n%s", want, text)
		}
	}
	if len(hits) != 2 || hits[1] != "fix" {
		t.Fatalf("the button and its hint click the action: %v", hits)
	}
}
