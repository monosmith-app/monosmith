package fixwizard

import (
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestTheStepListNamesDoneAndComingSteps(t *testing.T) {
	steps := []protocol.FixStep{{Description: "install"}, {Description: "refresh lock"}, {Description: "run checks"}}
	got := strings.Join(stepListRows(theme.Default(), steps, 2, 60), "\n")
	for _, want := range []string{"step 1: install", "step 3: run checks"} {
		if !strings.Contains(got, want) {
			t.Fatalf("lacks %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "refresh lock") {
		t.Fatalf("the running step belongs to its own line:\n%s", got)
	}
	if rows := stepListRows(theme.Default(), make([]protocol.FixStep, 3), 1, 60); rows != nil {
		t.Fatalf("unnamed steps give no list: %v", rows)
	}
}

func TestElapsedShowsOnlyOnceItMatters(t *testing.T) {
	for d, want := range map[time.Duration]string{time.Second: "", 12 * time.Second: "Running for 12s", 75 * time.Second: "Running for 1m 15s"} {
		if got := elapsedText(d); got != want {
			t.Fatalf("%v: %q, want %q", d, got, want)
		}
	}
}

func TestApplyingTellsWhatKeysDoAndHowLongItRuns(t *testing.T) {
	w := New(question(), show).Place(100, 30)
	w, _ = press(t, w, "1", "1", "enter")
	w.started = time.Now().Add(-20 * time.Second)
	got := text(w)
	for _, want := range []string{"Running for 20s", "Other keys do nothing until it is done"} {
		if !strings.Contains(got, want) {
			t.Fatalf("lacks %q:\n%s", want, got)
		}
	}
}
