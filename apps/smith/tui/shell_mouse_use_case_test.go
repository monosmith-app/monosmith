package tui

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
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
	x, y := l.findAfter("Doctor", 0) // the one row of tabs, the first on a wide screen
	if y != 0 {
		t.Fatalf("the tab labels are on row %d", y)
	}
	l.send(click(x+1, y))
	l.has("2 failed", "× go.work lists every module", "fix all (F)")
}

func TestClickingAFooterHintPressesItsKey(t *testing.T) {
	l := newLab(t, true, 260, 34)
	x, y := l.find("test orders (t)")
	l.send(click(x+2, y))
	if l.lastStarted() != "run test -p orders" {
		t.Fatalf("started %q", l.lastStarted())
	}
	l = newLab(t, true, 160, 34)
	x, y = l.find("help (?)")
	l.send(click(x+1, y))
	l.has("Keyboard shortcuts")
	l.keys("esc")
	x, y = l.find("mouse (m)")
	l.send(click(x+1, y))
	if !l.sh().mouseOff {
		t.Fatal("clicking the m hint turns the mouse off")
	}
}

func TestHoveringAFooterHintHighlightsIt(t *testing.T) {
	l := newLab(t, true, 160, 34)
	before := l.sh().View()
	x, y := l.find("lint orders (l)")
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
	x, y := l.find("1-6 views")
	l.send(click(x+1, y))
	if l.sh().tab != "projects" || len(l.started) != 0 {
		t.Fatal("nothing happens")
	}
}

func TestClickingARowSelectsItAndTheDetailFollows(t *testing.T) {
	l := newLab(t, true, 140, 34)
	x, y := l.find("● web")
	l.send(click(x+2, y))
	l.has("dependencies", "kit", "affected", "yes")
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
	tx, ty := l.find("Test app")
	l.send(click(tx+2, ty)).send(letGo(tx+2, ty))
	if l.lastStarted() != "run test -p web" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestDoubleClickingAProjectOpensItsDetailOnANarrowScreen(t *testing.T) {
	l := newLab(t, true, 90, 30)
	x, y := l.find("○ kit")
	l.hasNot("Test library")
	l.send(click(x+2, y))
	l.hasNot("Test library")
	l.advance(2 * time.Second)
	l.doubleClick(x+2, y)
	l.has("dependants", "Test library")
	// A click on an action of the detail does it.
	ax, ay := l.find("Test library")
	l.send(click(ax+2, ay)).send(letGo(ax+2, ay))
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
	l.has("4 projects")
	l.hasNot("Build orders")
	l.keys("ctrl+p").typeText("Build all").keys("enter")
	if l.lastStarted() != "run build" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestClickingAnActionOfTheDetailRunsIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	x, y := l.findAfter("Test app", 3)
	l.send(click(x+3, y)).send(letGo(x+3, y))
	if l.lastStarted() != "run test -p orders" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestClickingAProjectInTheTableSelectsIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	x, y := l.find("○ kit")
	l.send(click(x+2, y))
	l.has("dependants")
}

func TestClickingADoctorFindingOpensItsRemedy(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("d")
	l.has("How to fix", "go work use ./libs/x") // the detail beside the list tells the selected check
	x, y := l.find("× Docker registries")
	l.hasNot("write web's registry")
	l.send(click(x+3, y)) // selecting a check shows its fix in the detail
	l.has("Fix", "write web's registry")
}

func TestClickingTheRunButtonOfTheDialogRunsAndCancelCancels(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("5", "x", "enter")
	x, y := l.find("Cancel (esc)")
	l.send(click(x+2, y))
	if len(l.started) != 0 {
		t.Fatal("cancel must not run")
	}
	l.hasNot("The command that will run:")
	l.keys("x", "enter")
	x, y = l.find("Release anyway (y)")
	l.send(click(x+2, y))
	if l.lastStarted() != "release --yes" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestClickingOutsideTheDialogCancelsIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("5", "x", "enter")
	l.send(click(2, 30))
	l.hasNot("The command that will run:")
	if len(l.started) != 0 {
		t.Fatal("nothing should run")
	}
}

func TestHoveringTheDialogButtonsHighlightsThem(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("5", "x", "enter")
	before := l.sh().View()
	x, y := l.find("Cancel (esc)")
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
	l.send(wheel(100, y-3, -1))
	l.has("paused")
	l.hasNot("output line 149")
	for i := 0; i < 60; i++ {
		l.send(wheel(100, y-3, 1))
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
	l.has("What are you building?")
	x, y := l.find("A package")
	l.send(click(x+2, y))
	l.has("› A package")
	l.doubleClick(x+2, y)
	l.hasNot("What are you building?")
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
	// Every tab is a button with its key, so the strip needs more room than the names alone.
	l := newLab(t, true, 99, 30)
	l.hasNot("›", "‹")
	for _, name := range []string{"Home", "Proj", "Doctor", "Deps", "Rel"} {
		l.has(name)
	}
	x, y := l.findAfter("Doctor", 0)
	l.send(click(x+1, y))
	if l.sh().tab != commands.TabDoctor {
		t.Fatalf("a click on Doctor opened %q", l.sh().tab)
	}
}

// modalCases open each kind of blocking overlay on the tab it is opened from;
// button is a button of that tab's content that sits clear of the modal.
var modalCases = []struct {
	name, tab, button string
	open              func(l *lab)
}{
	{"help", commands.TabProjects, "New project", func(l *lab) { l.keys("f1") }},
	{"palette", commands.TabProjects, "New project", func(l *lab) { l.keys("ctrl+p") }},
	{"confirmation", commands.TabRelease, "", func(l *lab) { l.keys("x", "enter") }},
	{"fix wizard", commands.TabDoctor, "fix all", func(l *lab) { l.keys("F") }},
}

// colourLab is a lab whose theme has colour, so a hover look would show on screen.
func colourLab(t *testing.T, w, h int, tab string) *lab {
	t.Helper()
	l := newLab(t, true, w, h)
	sh := l.sh()
	sh.th = theme.Default()
	l.m = sh
	return l.openTab(tab)
}

func TestWithoutAModalThePointerHighlightsTheTabsAndTheFooter(t *testing.T) {
	l := colourLab(t, 160, 50, commands.TabProjects)
	before := l.sh().View()
	x, y := l.find("Settings (6)")
	l.send(motion(x+2, y))
	if l.sh().View() == before {
		t.Fatal("the tab under the pointer shows its hover look")
	}
	before = l.sh().View()
	l.send(motion(3, l.sh().lay().Footer.Y))
	if l.sh().View() == before {
		t.Fatal("the footer button under the pointer shows its hover look")
	}
}

func TestWhileAModalIsOpenTheMouseActsOnlyOnIt(t *testing.T) {
	for _, c := range modalCases {
		t.Run(c.name, func(t *testing.T) {
			l := colourLab(t, 160, 50, c.tab)
			tx, ty := l.find("Settings (6)")
			if ty != 0 {
				t.Fatalf("the tab is on row %d", ty)
			}
			points := [][2]int{{tx + 2, ty}, {3, l.sh().lay().Footer.Y}, {0, 0}, {159, 0}}
			if c.button != "" {
				bx, by := l.find(c.button)
				points = append(points, [2]int{bx + 1, by})
			}
			c.open(l)
			if !l.sh().modalOpen() {
				t.Fatal("the modal is not open")
			}
			before, x0, y0 := l.sh().View(), l.sh().mouseX, l.sh().mouseY
			for _, p := range points {
				l.send(motion(p[0], p[1]))
				if got := l.sh().View(); got != before {
					t.Fatalf("a motion over (%d,%d) changed the screen:\n%s\nwas\n%s", p[0], p[1], pane.Strip(got), pane.Strip(before))
				}
			}
			if sh := l.sh(); sh.mouseX != x0 || sh.mouseY != y0 {
				t.Fatalf("the pointer behind the modal was tracked: %d,%d", sh.mouseX, sh.mouseY)
			}
			// A press outside the modal may close it (the palette and the help do),
			// so each press has a modal of its own.
			for _, p := range points {
				l := colourLab(t, 160, 50, c.tab)
				c.open(l)
				l.send(click(p[0], p[1])).send(letGo(p[0], p[1]))
				if sh := l.sh(); sh.tab != c.tab || sh.pressed != "" || sh.split.Dragging != "" || sh.selection.Active {
					t.Fatalf("a click at (%d,%d) reached the screen behind the modal: tab %q pressed %q", p[0], p[1], sh.tab, sh.pressed)
				}
			}
		})
	}
}

func TestTheWheelDoesNotScrollTheScreenBehindAModal(t *testing.T) {
	l := colourLab(t, 120, 40, commands.TabRelease)
	l.keys("x", "enter")
	before, lay := l.sh().View(), l.sh().lay()
	for _, dir := range []int{1, 1, 1, -1} {
		l.send(wheel(20, lay.Main.Y+3, dir))
	}
	if l.sh().View() != before {
		t.Fatal("the wheel moved the release list behind the confirmation")
	}
}

func TestAClickOnTheModalsOwnButtonStillWorks(t *testing.T) {
	l := newLab(t, true, 120, 40)
	l.openTab(commands.TabRelease)
	l.keys("x", "enter")
	l.has("[Cancel (esc)]")
	l.clickText("Cancel")
	if l.sh().modalOpen() {
		t.Fatal("Cancel closes the confirmation")
	}
	if len(l.started) != 0 {
		t.Fatalf("started %v", l.started)
	}
}

func TestAClickOnTheFixingDialogsStopButtonStillWorks(t *testing.T) {
	l := withInstallFix(newLab(t, true, 140, 34))
	l.applyHold = make(chan struct{})
	defer close(l.applyHold)
	l.keys("f", "enter", "enter")
	l.has("Fixing…")
	if !l.sh().modalOpen() || !l.sh().fixApplying() {
		t.Fatal("the fixing dialog is the modal")
	}
	l.clickText("Stop")
	cancelled := false
	for i := 0; i < 200 && !cancelled; i++ {
		l.fixMu.Lock()
		cancelled = l.applyCancel
		l.fixMu.Unlock()
		if !cancelled {
			time.Sleep(10 * time.Millisecond)
		}
	}
	if !cancelled {
		t.Fatal("the Stop button did not stop the step")
	}
}
