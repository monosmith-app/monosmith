package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

// pickTwo picks the first two projects of the list with Space and gives their names.
func pickTwo(l *lab) []string {
	l.openTab(commands.TabProjects)
	first := l.sh().list.SelectedName()
	l.keys(" ")
	second := l.sh().list.SelectedName()
	l.keys(" ")
	return []string{first, second}
}

func TestSpaceSelectsProjectsAndTheRunKeysRunThemTogether(t *testing.T) {
	l := newLab(t, true, 150, 44)
	names := pickTwo(l)
	if got := l.sh().pickedNames(); !reflect.DeepEqual(got, names) {
		t.Fatalf("Space picks the row and moves down: %v, want %v", got, names)
	}
	l.has("2 selected · ", "[x]", "[ ]", "2 selected: "+names[0]+", "+names[1])
	l.has("$ smith run build -p " + names[0] + " -p " + names[1])
	if foot := l.lines()[l.sh().lay().Footer.Y]; !strings.Contains(foot, "build 2 selected") || !strings.Contains(foot, "Esc clear selection") {
		t.Fatalf("the footer names the selection: %q", foot)
	}
	l.keys("t")
	if got := l.lastStarted(); got != "run test -p "+names[0]+" -p "+names[1] {
		t.Fatalf("t runs the test of both: %q", got)
	}
	if l.sh().tab != commands.TabRuns {
		t.Fatalf("the run opens in the Runs view, not %q", l.sh().tab)
	}
}

func TestSpaceOnAPickedProjectLetsItGoAndEscClearsTheSelection(t *testing.T) {
	l := newLab(t, true, 150, 44)
	names := pickTwo(l)
	l.keys("up", " ")
	if got := l.sh().pickedNames(); !reflect.DeepEqual(got, names[:1]) {
		t.Fatalf("Space on a picked row lets it go: %v", got)
	}
	l.keys("/")
	l.typeText("kit")
	l.keys("enter", "esc")
	if len(l.sh().pickedNames()) != 0 || l.sh().list.Filter() != "kit" {
		t.Fatalf("the first Esc clears the selection, not the filter: %v, %q", l.sh().pickedNames(), l.sh().list.Filter())
	}
	l.hasNot("[x]", "[ ]", "selected")
	l.keys("esc")
	if l.sh().list.Filter() != "" {
		t.Fatal("the next Esc clears the filter")
	}
	l.keys("b")
	if got := l.lastStarted(); strings.Count(got, "-p") != 1 {
		t.Fatalf("with nothing selected b builds the project under the cursor: %q", got)
	}
}

func TestAClickOnARunButtonOfTheSelectionRunsIt(t *testing.T) {
	l := newLab(t, true, 150, 44)
	names := pickTwo(l)
	l.clickText("Test (t)")
	if got := l.lastStarted(); got != "run test -p "+names[0]+" -p "+names[1] {
		t.Fatalf("the Test button runs what t runs: %q", got)
	}
}

func TestTheSelectionCountsOnlyOnProjects(t *testing.T) {
	l := newLab(t, true, 150, 44)
	pickTwo(l)
	l.openTab(commands.TabHome)
	l.keys("b")
	if got := l.lastStarted(); strings.Count(got, "-p") > 1 {
		t.Fatalf("Home builds the selected project, not the picked ones: %q", got)
	}
}

func TestThePaletteNamesTheSelection(t *testing.T) {
	l := newLab(t, true, 150, 44)
	pickTwo(l)
	for _, e := range commands.PaletteEntries(l.sh().state()) {
		if e.ID == "run.build" && (e.Title != "Build 2 selected" || !strings.HasPrefix(e.Hint, "smith run build -p ")) {
			t.Fatalf("build: %q, %q", e.Title, e.Hint)
		}
	}
}

func TestASelectionAndPinsFitEveryScreen(t *testing.T) {
	for _, size := range [][2]int{{80, 30}, {120, 40}, {150, 44}} {
		l := newLab(t, true, size[0], size[1])
		sh := l.sh()
		sh.pins = []string{"web", "kit", "orders", "sdk"}
		l.m = sh
		pickTwo(l)
		lines := l.lines()
		if len(lines) != size[1] {
			t.Fatalf("%v: %d rows", size, len(lines))
		}
		for y, line := range lines {
			if got := pane.Width(line); got != size[0] {
				t.Fatalf("%v: row %d is %d cells: %q", size, y, got, line)
			}
		}
		if size[0] == 80 {
			l.has("2 selected · ")
		}
	}
}
