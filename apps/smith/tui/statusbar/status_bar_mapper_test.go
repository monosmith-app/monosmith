package statusbar

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestTheBarShowsTheExactCommandTheResultTheBranchAndTheVersion(t *testing.T) {
	got := Render(theme.Default(), Info{Command: "smith run build -p kit", State: Passed, Result: "4/4 tasks", Branch: "main", Version: "1.2.3"}, 100)
	for _, want := range []string{"$ smith run build -p kit", "passed", "4/4 tasks", "main", "smith 1.2.3"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
	if pane.Width(got) != 100 {
		t.Fatalf("width %d", pane.Width(got))
	}
}

func TestARunningCommandShowsTheSpinnerAndTheWord(t *testing.T) {
	th := theme.Default()
	got := Render(th, Info{Command: "smith test", State: Running, Result: "1/4", Tick: 3}, 80)
	if !strings.Contains(got, th.Icons.Spinner[3]) || !strings.Contains(got, "running") {
		t.Fatal(got)
	}
}

func TestAHintReplacesTheCommandWhileItIsSet(t *testing.T) {
	got := Render(theme.Default(), Info{Command: "smith test", Hint: "Stop is unavailable: nothing is running"}, 80)
	if !strings.Contains(got, "nothing is running") || strings.Contains(got, "smith test") {
		t.Fatal(got)
	}
}

func TestOnANarrowScreenTheRightSideGoesFirstAndTheWidthIsExact(t *testing.T) {
	in := Info{Command: "smith run lint test --affected", State: Failed, Result: "1 failed", Branch: "feat/something-long", Version: "1.2.3"}
	for _, w := range []int{80, 60, 40, 20, 5, 1} {
		if got := Render(theme.Default(), in, w); pane.Width(got) != w {
			t.Fatalf("width %d: got %d", w, pane.Width(got))
		}
	}
	if got := Render(theme.Default(), in, 50); strings.Contains(got, "smith 1.2.3") {
		t.Fatalf("version should be dropped at 50: %q", got)
	}
}
