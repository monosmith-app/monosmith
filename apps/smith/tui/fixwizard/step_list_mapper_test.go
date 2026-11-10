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

func TestApplyingShowsTheLastLineOfOutputAndElapsedAfterTwoSeconds(t *testing.T) {
	w := New(question(), show).Place(100, 30)
	w, _ = press(t, w, "1", "1", "enter")
	w.started = time.Now().Add(-1 * time.Second)
	if got := text(w); strings.Contains(got, "1s") {
		t.Fatalf("under two seconds shows no time: %s", got)
	}
	w.started = time.Now().Add(-75 * time.Second)
	got := text(w.Noted([]string{"added 12 packages"}))
	for _, want := range []string{"1m 15s", "added 12 packages"} {
		if !strings.Contains(got, want) {
			t.Fatalf("lacks %q: %s", want, got)
		}
	}
}

func TestApplyingTellsWhatKeysDoAndHowLongItRuns(t *testing.T) {
	w := New(question(), show).Place(100, 30)
	w, _ = press(t, w, "1", "1", "enter")
	w.started = time.Now().Add(-20 * time.Second)
	got := text(w)
	for _, want := range []string{"20s", "other keys wait", "l full log"} {
		if !strings.Contains(got, want) {
			t.Fatalf("lacks %q:\n%s", want, got)
		}
	}
}
