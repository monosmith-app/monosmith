package progress

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestDrawShowsTheBarAndTheCounts(t *testing.T) {
	got := Draw(theme.Default(), 3, 8, 30, false)
	if !strings.Contains(got, "3/8 37%") || !strings.Contains(got, "█") || !strings.Contains(got, "░") {
		t.Fatalf("%q", got)
	}
	if w := len([]rune(got)); w != 30 {
		t.Fatalf("width %d: %q", w, got)
	}
}

func TestDrawFallsBackToTheCountsWhenThereIsNoRoom(t *testing.T) {
	if got := Draw(theme.Default(), 3, 8, 8, false); got != "3/8 37%" {
		t.Fatalf("%q", got)
	}
}

func TestDrawUsesASCIIWhenAsked(t *testing.T) {
	th := theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII})
	got := Draw(th, 4, 8, 24, true)
	if !strings.Contains(got, "#") || !strings.Contains(got, "-") || strings.Contains(got, "█") {
		t.Fatalf("%q", got)
	}
}
