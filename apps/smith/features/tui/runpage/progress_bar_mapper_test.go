package runpage

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

func TestTheBarFillsItsShareAndSaysTheCounts(t *testing.T) {
	bar := pane.Strip(Bar(plain, 2, 4, false, 40))
	if pane.Width(bar) != 40 || !strings.HasSuffix(bar, "2/4 50%") {
		t.Fatalf("bar %q", bar)
	}
	full, empty := strings.Count(bar, "█"), strings.Count(bar, "░")
	if full == 0 || full+empty != 40-pane.Width("  2/4 50%") || full-empty > 1 || empty-full > 1 {
		t.Fatalf("half a bar is %d full and %d empty: %q", full, empty, bar)
	}
}

func TestABarWithNoTotalOrNoRoomStillSaysTheCounts(t *testing.T) {
	if got := pane.Strip(Bar(plain, 0, 0, false, 30)); !strings.HasSuffix(got, "0/0 0%") || strings.Contains(got, "█") {
		t.Fatalf("empty bar %q", got)
	}
	if got := strings.TrimSpace(pane.Strip(Bar(plain, 1, 3, true, 8))); got != "1/3 33%" {
		t.Fatalf("a bar with no room %q", got)
	}
}
