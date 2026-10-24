package statusbar

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestTheBarShowsTheExactCommandTheResultAndTheVersion(t *testing.T) {
	got := Render(theme.Default(), Info{Command: "smith run build -p kit", State: Passed, Result: "4/4 tasks", Version: "1.2.3"}, 100)
	for _, want := range []string{"$ smith run build -p kit", "passed", "4/4 tasks", "smith 1.2.3"} {
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

func TestALongHintEndsInAnEllipsis(t *testing.T) {
	got := Render(theme.Default(), Info{Hint: "Stop is unavailable: nothing is running and there is a lot more to say about it"}, 40)
	if !strings.Contains(got, "…") || pane.Width(got) != 40 {
		t.Fatalf("got %q (%d cells)", got, pane.Width(got))
	}
}

func TestAHintReplacesTheCommandWhileItIsSet(t *testing.T) {
	got := Render(theme.Default(), Info{Command: "smith test", Hint: "Stop is unavailable: nothing is running"}, 80)
	if !strings.Contains(got, "nothing is running") || strings.Contains(got, "smith test") {
		t.Fatal(got)
	}
}

func TestOnANarrowScreenTheRightSideGoesFirstAndTheWidthIsExact(t *testing.T) {
	in := Info{Command: "smith run lint test --affected", State: Failed, Result: "1 failed", Version: "1.2.3"}
	for _, w := range []int{80, 60, 40, 20, 5, 1} {
		if got := Render(theme.Default(), in, w); pane.Width(got) != w {
			t.Fatalf("width %d: got %d", w, pane.Width(got))
		}
	}
	if got := Render(theme.Default(), in, 50); strings.Contains(got, "smith 1.2.3") {
		t.Fatalf("version should be dropped at 50: %q", got)
	}
}

func TestTheBarLeavesTheKeysToTheFooterAndTheWorkspaceToTheHeader(t *testing.T) {
	idle := Render(theme.Default(), Info{Version: "1.2.3"}, 120)
	for _, want := range []string{"ready", "smith 1.2.3"} {
		if !strings.Contains(idle, want) {
			t.Fatalf("missing %q in %q", want, idle)
		}
	}
	for _, stale := range []string{"ctrl+p", "f1", "q quit", "y copy"} {
		if strings.Contains(idle, stale) {
			t.Fatalf("the key hints live in the footer, not the bar: %q", idle)
		}
	}
}

func TestTheBarShowsWhenTheMouseIsOff(t *testing.T) {
	if got := Render(theme.Default(), Info{MouseOff: true}, 140); !strings.Contains(got, "mouse off") {
		t.Fatal(got)
	}
}
