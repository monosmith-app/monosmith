package fixwizard

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

func applyingWizard(t *testing.T) Wizard {
	t.Helper()
	w := New(question(), show).Place(100, 30)
	w, _ = press(t, w, "1", "1", "enter")
	return w.Progress(1, 2, "npm install")
}

func numbered(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("log line %03d", i)
	}
	return out
}

func TestApplyingShowsTheLastSixLinesInRowsThatKeepTheHeight(t *testing.T) {
	w := applyingWizard(t)
	empty := w.Rect().H
	w = w.Noted([]string{"one", "two", "three", "four", "five", "six", "seven"})
	got := text(w)
	if strings.Contains(got, "one") || !strings.Contains(got, "two") || !strings.Contains(got, "seven") {
		t.Fatalf("expected the last six lines:\n%s", got)
	}
	if w.Rect().H != empty {
		t.Fatalf("height changed from %d to %d", empty, w.Rect().H)
	}
}

func TestLOpensTheFullLogAndEscClosesItBeforeItStopsTheStep(t *testing.T) {
	w := applyingWizard(t).Logged(numbered(80))
	w, out := press(t, w, "l")
	if !w.LogOpen() || out != (Outcome{}) {
		t.Fatalf("l did not open the log: %+v", out)
	}
	if got := text(w); !strings.Contains(got, "full log") || !strings.Contains(got, "Close log (l)") {
		t.Fatalf("log view:\n%s", got)
	}
	w, out = press(t, w, "esc")
	if w.LogOpen() || out.Stop {
		t.Fatalf("esc must close the log first: open=%v out=%+v", w.LogOpen(), out)
	}
	if _, out = press(t, w, "esc"); !out.Stop {
		t.Fatal("a second esc stops the step")
	}
	w, _ = press(t, w, "l")
	if _, out = press(t, w, "ctrl+c"); !out.Stop {
		t.Fatal("ctrl+c stops the step even with the log open")
	}
}

func TestTheFullLogFollowsTheNewestLineUntilTheReaderScrollsUp(t *testing.T) {
	w := applyingWizard(t).Logged(numbered(80))
	w, _ = press(t, w, "l")
	if got := text(w); !strings.Contains(got, "log line 079") {
		t.Fatalf("log opens on the newest line:\n%s", got)
	}
	w, _ = press(t, w, "pgup")
	top := w.top
	w = w.Logged(numbered(90))
	if w.top != top {
		t.Fatalf("scrolled view moved from %d to %d", top, w.top)
	}
	w, _ = press(t, w, "end")
	w = w.Logged(numbered(95))
	if w.top != w.bottom() {
		t.Fatal("back at the end, the view follows again")
	}
}

func TestAClickOnTheLogHintOpensTheLogAndTheCloseButtonClosesIt(t *testing.T) {
	w := applyingWizard(t).Logged(numbered(3))
	r := w.Rect()
	y := r.Y + len(w.bodyRows())
	if got := pane.Strip(w.Lines(plain(), 0)[y-r.Y]); !strings.Contains(got, "l full log") {
		t.Fatalf("hint row is not at y=%d: %q", y, got)
	}
	if w2, _ := w.Click(r.X+3, y); w2.LogOpen() {
		t.Fatal("a click on the stop words must not open the log")
	}
	w, _ = w.Click(r.X+2+ansi.StringWidth(applyingHint)-2, y)
	if !w.LogOpen() {
		t.Fatal("a click on the hint did not open the log")
	}
	r = w.Rect()
	w, _ = w.Click(r.X+w.buttonSpans(r.W - 4)[0][0], r.Y+r.H-2)
	if w.LogOpen() {
		t.Fatal("the Close log button did not close the log")
	}
}
