package detail

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func splitPanel() Panel {
	return Panel{Actions: []Action{
		{ID: "a", Label: "Alpha", Shortcut: "a", Hint: "does alpha, a long sentence that wraps over several lines of a narrow panel"},
		{ID: "b", Label: "Beta", Shortcut: "b"},
	}}
}

func TestAPinnedRunPartLeavesTheViewAndKeepsTheRuleAndTheButton(t *testing.T) {
	p := splitPanel()
	p.PinRun = true
	view, hits := p.View(theme.Default(), 30, "")
	for _, id := range hits {
		if id == RunID {
			t.Fatal("the pinned run button is still in the view")
		}
	}
	if strings.Contains(pane.Strip(strings.Join(view, "\n")), "does alpha") {
		t.Fatal("the pinned description is still in the view")
	}
	for _, limit := range []int{2, 3, 20} {
		lines, runHits := p.PinnedRun(theme.Default(), 30, "", limit)
		if len(lines) > limit || len(lines) < 2 {
			t.Fatalf("limit %d: %d lines", limit, len(lines))
		}
		if !strings.HasPrefix(strings.TrimSpace(pane.Strip(lines[0])), "──") || runHits[len(lines)-1] != RunID || !strings.Contains(pane.Strip(lines[len(lines)-1]), "Alpha (a)") {
			t.Fatalf("limit %d: the rule and the run button must stay:\n%s", limit, strings.Join(lines, "\n"))
		}
	}
	if lines, _ := p.PinnedRun(theme.Default(), 30, "", 3); !strings.HasSuffix(strings.TrimSpace(pane.Strip(lines[1])), "…") {
		t.Fatalf("a cut description ends with an ellipsis: %q", pane.Strip(lines[1]))
	}
}

func TestAPanelThatDoesNotSplitPinsNothing(t *testing.T) {
	p := Panel{Actions: []Action{{ID: "a", Label: "Alpha", Shortcut: "a", Hint: "does alpha"}}}
	if lines, _ := p.PinnedRun(theme.Default(), 30, "", 10); len(lines) != 0 {
		t.Fatalf("one action does not split: %q", lines)
	}
}
