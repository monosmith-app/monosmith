package projectpage

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

func TestUsesAreTwoBoxesOnAWideDetailAndTwoRowsOnANarrowOne(t *testing.T) {
	wide := Uses(plain, []string{"core", "kit"}, []string{"web"}, 76)
	text := pane.Strip(strings.Join(wide, "\n"))
	if len(wide) != 5 || !strings.Contains(text, "╭─ Uses") || !strings.Contains(text, "╭─ Used by") || !strings.Contains(text, "core · kit") || !strings.Contains(text, "a change here tests them too") {
		t.Fatalf("wide:\n%s", text)
	}
	narrow := Uses(plain, nil, []string{"web"}, 50)
	if len(narrow) != 2 || !strings.Contains(pane.Strip(narrow[0]), "Uses") || !strings.Contains(pane.Strip(narrow[0]), "nothing") || !strings.Contains(pane.Strip(narrow[1]), "web") {
		t.Fatalf("narrow: %q", narrow)
	}
	for _, l := range append(wide, narrow...) {
		if w := pane.Width(l); w != 76 && w != 50 {
			t.Fatalf("a line is %d cells: %q", w, pane.Strip(l))
		}
	}
}
