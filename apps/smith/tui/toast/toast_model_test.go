package toast

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestAToastShowsItsTextWithASymbolAndExpires(t *testing.T) {
	to := New("3 projects installed", Success)
	lines := to.Lines(theme.Default(), 60)
	if len(lines) != 3 || !strings.Contains(lines[1], "3 projects installed") || !strings.Contains(lines[1], "✓") {
		t.Fatalf("%q", lines)
	}
	for _, l := range lines {
		if pane.Width(l) != pane.Width(lines[0]) {
			t.Fatalf("ragged box: %q", lines)
		}
	}
	for i := 0; i < Lifetime; i++ {
		to = to.Tick()
	}
	if to.Shown() || to.Lines(theme.Default(), 60) != nil {
		t.Fatal("expired")
	}
}

func TestAToastCanBeDismissedAndAZeroToastIsHidden(t *testing.T) {
	if (Toast{}).Shown() || New("x", Info).Dismiss().Shown() {
		t.Fatal("hidden")
	}
}

func TestALongMessageIsCutToTheRoom(t *testing.T) {
	lines := New(strings.Repeat("long ", 40), Error).Lines(theme.Default(), 30)
	if pane.Width(lines[0]) > 30 {
		t.Fatalf("%d", pane.Width(lines[0]))
	}
}
