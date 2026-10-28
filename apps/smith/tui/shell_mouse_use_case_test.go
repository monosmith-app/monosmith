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
	x, y := l.findAfter("Doctor", 0) // the label row of the tab boxes, under the header
	if y != 2 {
		t.Fatalf("the tab labels are on row %d", y)
	}
	l.send(click(x+1, y))
	l.has("2 failed", "✗ go.work lists every module", "F fix all")
}

func TestClickingAFooterHintPressesItsKey(t *testing.T) {
	l := newLab(t, true, 160, 34)
	x, y := l.find("t test")
	l.send(click(x+2, y))
	if l.lastStarted() != "run test -p orders" {
		t.Fatalf("started %q", l.lastStarted())
	}
	l = newLab(t, true, 160, 34)
	x, y = l.find("? help")
	l.send(click(x+1, y))
	l.has("Keyboard shortcuts")
	l.keys("esc")
	x, y = l.find("m mouse")
	l.send(click(x+1, y))
	if !l.sh().mouseOff {
		t.Fatal("clicking the m hint turns the mouse off")
	}
}

func TestHoveringAFooterHintHighlightsIt(t *testing.T) {
	l := newLab(t, true, 160, 34)
	before := l.sh().View()
	x, y := l.find("l lint")
	l.send(motion(x+2, y))
	if l.sh().View() == before {
		t.Fatal("hover should change how the hint is drawn")
	}
	l.send(motion(159, 20))
	if l.sh().View() != before {
		t.Fatal("moving away should restore it")
	}
}

func TestAHintThatStandsForSeveralKeysIsNotClickable(t *testing.T) {
	l := newLab(t, true, 160, 34)
	x, y := l.find("1-7 views")
	l.send(click(x+1, y))
	if l.sh().tab != "projects" || len(l.started) != 0 {
		t.Fatal("nothing happens")
	}
}

func TestClickingARowSelectsItAndTheDetailFollows(t *testing.T) {
	l := newLab(t, true, 140, 34)
	x, y := l.find("● web")
	l.send(click(x+2, y))
	l.has("Depends on", "kit", "affected", "yes")
	if l.sh().list.SelectedName() != "web" {
		t.Fatalf("selected %q", l.sh().list.SelectedName())
	}
}

func TestClickingATaskActionOfTheDetailRunsItOnTheProject(t *testing.T) {
	l := newLab(t, true, 140, 34)
	x, y := l.find("● web")
	l.send(click(x+2, y))
	if len(l.started) != 0 {
		t.Fatal("selecting a project must not run anything")
	}
	tx, ty := l.find("Run test")
	l.send(click(tx+2, ty))
	if l.lastStarted() != "run test -p web" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestDoubleClickingAProjectOpensItsDetailOnANarrowScreen(t *testing.T) {
	l := newLab(t, true, 90, 30)
	x, y := l.find("○ kit")
	l.hasNot("Used by")
	l.send(click(x+2, y))
	l.hasNot("Used by")
	l.advance(2 * time.Second)
	l.doubleClick(x+2, y)
	l.has("Used by", "Run test")
	// A click on an action of the detail does it.
	ax, ay := l.find("Run test")
	l.send(click(ax+2, ay))
	if l.lastStarted() != "run test -p kit" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestRightClickOpensNoMenu(t *testing.T) {
	l := newLab(t, true, 140, 34)
	before := l.view()
	x, y := l.find("● web")
	l.send(rightClick(x+2, y))
	l.send(rightClick(5, 4))
	if l.view() != before {
		t.Fatalf("a right click changed the screen: %s", l.view())
	}
	l.keys("ctrl+o")
	if l.view() != before {
		t.Fatal("ctrl+o is no longer bound")
	}
}

func TestStopIsOfferedOnlyWhileARunGoes(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("2")
	l.hasNot("[■ Stop")
}

func TestTheResultToastLeavesWithTheScreenItWasRaisedOn(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("b")
	if !l.sh().toast.Shown() {
		t.Skip("the run has not ended in this lab")
	}
	l.keys("1")
	if l.sh().toast.Shown() {
		t.Fatal("the toast stayed on another screen")
	}
}

func TestBuildAllRunsEveryProjectAndTheScopeIsStatedOnce(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.has("On: orders")
	l.hasNot("Build orders")
	l.keys("ctrl+p").typeText("Build all").keys("enter")
	if l.lastStarted() != "run build" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestClickingAnActionOfTheDetailRunsIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	x, y := l.findAfter("Run test", 3)
	l.send(click(x+3, y))
	if l.lastStarted() != "run test -p orders" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestClickingAProjectInTheTableSelectsIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	x, y := l.find("○ kit")
	l.send(click(x+2, y))
	l.has("Used by")
}

func TestClickingADoctorFindingOpensItsRemedy(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("d")
	l.has("How to fix", "go work use ./libs/x") // the detail beside the list tells the selected check
	x, y := l.find("✗ Docker registries")
	l.hasNot("write web's registry")
	l.send(click(x+3, y)) // selecting a check shows its fix in the detail
	l.has("Fix", "write web's registry")
}

func TestClickingTheRunButtonOfTheDialogRunsAndCancelCancels(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("6", "x")
	x, y := l.find("[ Cancel")
	l.send(click(x+2, y))
	if len(l.started) != 0 {
		t.Fatal("cancel must not run")
	}
	l.hasNot("The command that will run:")
	l.keys("x")
	x, y = l.find("[ Run")
	l.send(click(x+2, y))
	if l.lastStarted() != "release --yes" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestClickingOutsideTheDialogCancelsIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("6", "x")
	l.send(click(2, 30))
	l.hasNot("The command that will run:")
	if len(l.started) != 0 {
		t.Fatal("nothing should run")
	}
}

func TestHoveringTheDialogButtonsHighlightsThem(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("6", "x")
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
	lines = append(lines, "kit:build ✓ 1s")
	l.next = &scripted{lines: lines}
	l.keys("b")
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

func TestTheWheelScrollsTheProjectListAndNeverMovesTheSelection(t *testing.T) {
	l := newLab(t, true, 140, 34)
	before := l.sh().list.SelectedName()
	for range 3 {
		l.send(wheel(5, 10, 1))
	}
	l.send(wheel(5, 10, -1))
	if got := l.sh().list.SelectedName(); got != before {
		t.Fatalf("the wheel moved the selection from %q to %q", before, got)
	}
}

func TestClickingATaskInTheRunListShowsItsLog(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.next = failingRun()
	l.keys("b")
	l.has("main.go:12")
	x, y := l.find("✓ kit:build")
	l.send(click(x+3, y))
	l.has("Log: kit:build", "ok example.com/kit")
	l.hasNot("main.go:12")
	x, y = l.find("All output")
	l.send(click(x+3, y))
	l.has("Log: all output", "main.go:12", "ok example.com/kit")
}

func TestClickingAGuideOptionSelectsAndDoubleClickingChoosesIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("a")
	l.has("What are you adding?")
	x, y := l.find("A package")
	l.send(click(x+2, y))
	l.has("› A package")
	l.doubleClick(x+2, y)
	l.hasNot("What are you adding?")
}

func TestClickingAPaletteRowRunsIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("ctrl+p").typeText("doctor")
	x, y := l.findAfter("Doctor", 3)
	l.send(click(x+2, y))
	l.hasNot("Command palette")
	l.has("2 failed")
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

func TestEveryTabIsOnTheStripAtTheLeastWidthAndAClickOpensIt(t *testing.T) {
	l := newLab(t, true, 72, 30)
	l.hasNot("›", "‹")
	for _, name := range []string{"Home", "Projects", "Run", "Doctor", "Deps", "Release"} {
		l.has(name)
	}
	x, y := l.findAfter("Run", 0)
	l.send(click(x+1, y))
	if l.sh().tab != "run" {
		t.Fatalf("a click on Run opened %q", l.sh().tab)
	}
}
