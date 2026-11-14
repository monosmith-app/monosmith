package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

// runLab is a lab with a build run open in the run panel.
func runLab(t *testing.T) *lab {
	t.Helper()
	l := newLab(t, true, 140, 36)
	l.next = failingRun()
	l.keys("b")
	if !l.sh().runShown() {
		t.Fatal("no run panel")
	}
	return l
}

func TestTheKeyMakesTheRunFillTheTerminalUnderATopBarAndEscGoesBack(t *testing.T) {
	l := runLab(t)
	l.keys("Z")
	if !l.sh().fullRun() {
		t.Fatal("Z did not make the run full screen")
	}
	rows := strings.Split(l.view(), "\n")
	if len(rows) != 36 {
		t.Fatalf("the screen has %d rows, want 36", len(rows))
	}
	if !strings.Contains(rows[0], "[-]") && !strings.Contains(rows[0], " - ") {
		t.Errorf("the top bar has no button to go back: %q", rows[0])
	}
	view := l.view()
	for _, gone := range []string{"Projects (2)", "palette (:)", "quit (q)"} {
		if strings.Contains(view, gone) {
			t.Errorf("%q is still on the screen in full screen:\n%s", gone, view)
		}
	}
	if !strings.Contains(view, "orders:build") {
		t.Errorf("the run is not shown:\n%s", view)
	}
	l.keys("esc")
	if l.sh().fullRun() || !strings.Contains(l.view(), "Projects (2)") {
		t.Error("Esc did not give the screen back")
	}
}

func TestTheEnlargeButtonAndTheTopBarButtonSwitchTheView(t *testing.T) {
	l := runLab(t)
	sh := l.sh()
	in := sh.detailInner()
	hit, ok := sh.sideContent(in.W, in.H).hits[0]
	if !ok || len(hit.spans) != 1 || hit.spans[0].id != "x:run-fullscreen" {
		t.Fatalf("the run header has no enlarge button: %+v", hit)
	}
	x := in.X + hit.spans[0].from + 1
	l.send(click(x, in.Y)).send(letGo(x, in.Y))
	if !l.sh().fullRun() {
		t.Fatalf("the + button did not enlarge the run:\n%s", l.view())
	}
	first, _ := l.sh().fullBarSpan()
	l.send(click(first+1, 0)).send(letGo(first+1, 0))
	if l.sh().fullRun() {
		t.Fatal("the - button of the top bar did not give the screen back")
	}
}

func TestADoubleClickOnTheOutputGoesFullScreenAndBackButNotOnAButton(t *testing.T) {
	l := runLab(t)
	sh := l.sh()
	in := sh.detailInner()
	// A row of the log, well under the buttons.
	x, y := in.X+4, in.Y+in.H-2
	l.send(click(x, y)).send(letGo(x, y))
	l.advance(120 * time.Millisecond)
	l.send(click(x, y)).send(letGo(x, y))
	if !l.sh().fullRun() {
		t.Fatalf("a double click on the output did not enlarge it:\n%s", l.view())
	}
	l.advance(2 * time.Second)
	l.send(click(5, 12)).send(letGo(5, 12))
	l.advance(120 * time.Millisecond)
	l.send(click(5, 12)).send(letGo(5, 12))
	if l.sh().fullRun() {
		t.Fatal("a double click on the output did not go back")
	}
}

func TestFullScreenEndsWhenTheRunClosesAndTheViewHasNoSplitOfTheOldLayout(t *testing.T) {
	l := runLab(t)
	l.keys("Z")
	l.keys("3")
	if l.sh().fullRun() {
		t.Error("full screen stayed open on another tab")
	}
	l.keys("2")
	if l.sh().runFull {
		t.Error("the flag outlived the tab: the next run would open full screen")
	}
	if strings.Contains(pane.Strip(l.view()), "Esc or - goes back") {
		t.Error("the top bar stayed on screen")
	}
}
