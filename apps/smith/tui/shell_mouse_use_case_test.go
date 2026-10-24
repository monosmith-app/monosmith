package tui

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// findAfter is find, starting at screen row minRow.
func (l *lab) findAfter(text string, minRow int) (x, y int) {
	l.t.Helper()
	for row, line := range l.lines() {
		if row < minRow {
			continue
		}
		if i := strings.Index(line, text); i >= 0 {
			return len([]rune(line[:i])), row
		}
	}
	l.t.Fatalf("%q not on screen from row %d:\n%s", text, minRow, l.view())
	return 0, 0
}

func (l *lab) doubleClick(x, y int) {
	l.send(click(x, y))
	l.advance(120 * time.Millisecond)
	l.send(click(x, y))
}

func TestClickingATabSwitchesTheView(t *testing.T) {
	l := newLab(t, true, 140, 34)
	_, y := l.findAfter("Doctor", 1) // the tab row
	x, _ := l.findAfter("Doctor", 1)
	l.send(click(x+1, y))
	l.has("2 passed", "✖ go.work lists every module", "Re-check")
}

func TestClickingAMenuHeadingOpensItAndHoveringAnotherSwitchesMenus(t *testing.T) {
	l := newLab(t, true, 140, 34)
	x, _ := l.find("Run")
	l.send(click(x+1, 0))
	l.has("Run selected task", "Ctrl+R")
	rx, _ := l.find("Release")
	l.send(motion(rx+1, 0))
	l.has("Preview next release", "F2")
	l.hasNot("Run selected task")
	// A click on an entry runs it.
	ex, ey := l.findAfter("Preview next release", 1)
	l.send(click(ex+1, ey))
	if l.lastStarted() != "release --dry-run" {
		t.Fatalf("started %q", l.lastStarted())
	}
	l.hasNot("Preview next release  ")
}

func TestClickingOutsideAnOpenMenuClosesItAndRunsNothing(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+r")
	l.has("Run selected task")
	l.send(click(100, 25))
	l.hasNot("Run selected task")
	if len(l.started) != 0 {
		t.Fatal("nothing should run")
	}
	// Clicking the open heading again also closes it.
	x, _ := l.find("Run")
	l.send(click(x+1, 0))
	l.has("Run selected task")
	l.send(click(x+1, 0))
	l.hasNot("Run selected task")
}

func TestHoveringADisabledMenuEntryExplainsWhyInTheStatusBar(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+r")
	x, y := l.findAfter("Stop", 1)
	l.send(motion(x+1, y))
	l.has("Stop is unavailable: nothing is running")
	l.send(click(x+1, y))
	if len(l.started) != 0 {
		t.Fatal("a disabled entry must not run")
	}
	l.has("Run selected task") // the menu stays open
}

func TestClickingARowSelectsItAndTheInspectorFollows(t *testing.T) {
	l := newLab(t, true, 140, 34)
	x, y := l.find("● web")
	l.send(click(x+2, y))
	l.has("Depends on", "kit", "affected", "yes")
	if l.sh().focus.String() != "navigator" {
		t.Fatalf("focus %v", l.sh().focus)
	}
}

func TestDoubleClickingATaskRunsItAndASingleClickDoesNot(t *testing.T) {
	l := newLab(t, true, 140, 34)
	x, y := l.find("● web")
	l.send(click(x+2, y))
	l.keys("right") // open web's tasks
	tx, ty := l.find("› build")
	l.send(click(tx+2, ty))
	if len(l.started) != 0 {
		t.Fatal("a single click must not run")
	}
	l.advance(2 * time.Second)
	l.doubleClick(tx+2, ty)
	if l.lastStarted() != "run build -p web" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestDoubleClickingAProjectOpensItsTasks(t *testing.T) {
	l := newLab(t, true, 140, 34)
	x, y := l.find("○ kit")
	l.hasNot("› build")
	l.doubleClick(x+2, y)
	l.has("› build", "› test")
	l.doubleClick(x+2, y)
	l.hasNot("› build")
}

func TestRightClickOpensAContextMenuForTheRowAndAClickRunsAnEntry(t *testing.T) {
	l := newLab(t, true, 140, 34)
	x, y := l.find("● web")
	l.send(rightClick(x+2, y))
	l.has("Show or hide tasks", "Explain", "Command palette…", "Keyboard shortcuts")
	bx, by := l.findAfter("Test", y)
	l.send(click(bx+1, by))
	if l.lastStarted() != "run test -p web" {
		t.Fatalf("started %q", l.lastStarted())
	}
	l.hasNot("Show or hide tasks")
}

func TestTheContextMenuOfATaskRunsIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("shift+tab", "down", "right", "down") // web › build
	tx, ty := l.find("› build")
	l.send(rightClick(tx+2, ty))
	l.has("Run build")
	rx, ry := l.findAfter("Run build", ty)
	l.send(click(rx+1, ry))
	if l.lastStarted() != "run build -p web" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestEscClosesTheContextMenu(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.send(rightClick(5, 4))
	l.has("Command palette…")
	l.keys("esc")
	l.hasNot("Command palette…")
	l.keys("ctrl+o") // the keyboard equivalent
	l.has("Command palette…")
}

func TestClickingAButtonRunsItsCommandOnTheSelection(t *testing.T) {
	l := newLab(t, true, 140, 34)
	x, y := l.find("[▶ Test")
	l.send(click(x+3, y))
	if l.lastStarted() != "run test -p orders" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestStopIsOfferedOnlyWhileARunGoes(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+2")
	l.hasNot("[■ Stop")
}

func TestTheResultToastLeavesWithTheScreenItWasRaisedOn(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("ctrl+b")
	if !l.sh().toast.Shown() {
		t.Skip("the run has not ended in this lab")
	}
	l.keys("alt+1")
	if l.sh().toast.Shown() {
		t.Fatal("the toast stayed on another screen")
	}
}

func TestBuildAllRunsEveryProjectAndTheButtonsNameTheirScope(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.has("[▶ Build orders", "[▶ Build all")
	x, y := l.find("[▶ Build all")
	l.send(click(x+3, y))
	if l.lastStarted() != "run build" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestHoveringAButtonHighlightsIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	before := l.sh().View()
	x, y := l.find("[▶ Lint")
	l.send(motion(x+3, y))
	if l.sh().View() == before {
		t.Fatal("hover should change how the button is drawn")
	}
	l.send(motion(139, 20))
	if l.sh().View() != before {
		t.Fatal("moving away should restore it")
	}
}

func TestClickingAnInspectorActionRunsIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	x, y := l.findAfter("Typecheck", 3) // the inspector's action, not the button
	l.send(click(x+3, y))
	if l.lastStarted() != "run typecheck -p orders" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestClickingAProjectInTheTableSelectsIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	x, y := l.findAfter("kit", 5) // the table row (the navigator is further left)
	for x < 40 {
		nx, ny := l.findAfter("lib", y+1)
		x, y = nx, ny
	}
	l.send(click(x+1, y))
	l.has("Used by")
}

func TestClickingADoctorFindingOpensItsRemedy(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("ctrl+d")
	x, y := l.find("✖ go.work")
	l.send(click(x+3, y))
	l.has("remedy: go work use ./libs/x")
	x, y = l.find("✖ go.work")
	l.send(click(x+3, y))
	l.hasNot("remedy:")
}

func TestClickingTheRunButtonOfTheDialogRunsAndCancelCancels(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("f3")
	x, y := l.find("[ Cancel")
	l.send(click(x+2, y))
	if len(l.started) != 0 {
		t.Fatal("cancel must not run")
	}
	l.hasNot("The command that will run:")
	l.keys("f3")
	x, y = l.find("[ Run")
	l.send(click(x+2, y))
	if l.lastStarted() != "release --yes" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestClickingOutsideTheDialogCancelsIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("f3")
	l.send(click(2, 30))
	l.hasNot("The command that will run:")
	if len(l.started) != 0 {
		t.Fatal("nothing should run")
	}
}

func TestHoveringTheDialogButtonsHighlightsThem(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("f3")
	before := l.sh().View()
	x, y := l.find("[ Cancel")
	l.send(motion(x+2, y))
	if l.sh().View() == before {
		t.Fatal("hover should change the button")
	}
}

func TestTheWheelScrollsTheLogAndPausesIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	lines := []string{"kit:build ▶ x"}
	for i := 0; i < 150; i++ {
		lines = append(lines, "kit:build │ output line "+strconv.Itoa(i))
	}
	lines = append(lines, "kit:build ✔ 1s")
	l.next = &scripted{lines: lines}
	l.keys("ctrl+b", "esc")
	l.has("output line 149", "following")
	_, y := l.find("output line 149")
	l.send(wheel(60, y-3, -1))
	l.has("paused")
	l.hasNot("output line 149")
	for i := 0; i < 60; i++ {
		l.send(wheel(60, y-3, 1))
	}
	l.has("following", "output line 149")
}

func TestTheWheelScrollsTheNavigatorWithoutMovingTheSelection(t *testing.T) {
	l := newLab(t, true, 140, 9) // short: the tree does not fit
	l.hasNot("sdk python")
	l.send(wheel(5, 5, 1))
	l.send(wheel(5, 5, 1))
	l.send(wheel(5, 5, 1))
	l.has("sdk python")
	l.has("orders") // the inspector still shows the selected project
	l.send(wheel(5, 5, -1))
	l.send(wheel(5, 5, -1))
	l.send(wheel(5, 5, -1))
	l.has("▾ Apps")
}

func TestClickingATaskInTheRunListShowsItsLog(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.next = failingRun()
	l.keys("ctrl+b")
	l.has("main.go:12")
	x, y := l.find("✔ kit:build")
	l.send(click(x+3, y))
	l.has("Log: kit:build", "ok example.com/kit")
	l.hasNot("main.go:12")
	x, y = l.find("All output")
	l.send(click(x+3, y))
	l.has("Log: all output", "main.go:12", "ok example.com/kit")
}

func TestClickingAGuideOptionSelectsAndDoubleClickingChoosesIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+6")
	x, y := l.find("Add a project")
	l.send(click(x+2, y))
	l.has("› Add a project")
	l.hasNot("What kind?")
	l.doubleClick(x+2, y)
	l.has("What kind?")
}

func TestClickingAPaletteRowRunsIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("ctrl+p").typeText("doctor")
	x, y := l.findAfter("Doctor", 3)
	l.send(click(x+2, y))
	l.hasNot("Command palette")
	l.has("2 passed")
	// Clicking outside closes it.
	l.keys("ctrl+p")
	l.send(click(1, 30))
	l.hasNot("Command palette")
}

func TestClickingOutsideTheHelpClosesIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("f1")
	l.has("Keyboard shortcuts")
	l.send(click(1, 32))
	l.hasNot("Keyboard shortcuts")
}

func TestTheNavigatorTabWorksWithTheMouseAt79Columns(t *testing.T) {
	l := newLab(t, true, 79, 24)
	_, y := l.find("Workspace ")
	l.send(click(2, y+1))
	x, ty := l.find("○ kit")
	l.doubleClick(x+2, ty)
	l.has("› build")
	tx, tty := l.find("› build")
	l.advance(3 * time.Second)
	l.doubleClick(tx+2, tty)
	if l.lastStarted() != "run build -p kit" {
		t.Fatalf("started %q", l.lastStarted())
	}
}
