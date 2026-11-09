package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dashboard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
)

const (
	moreBelow = "↓ more (PgDn, Ctrl+D)"
	moreAbove = "↑ more (PgUp, Ctrl+U)"
)

// homeTargets are the buttons and runs of Home the keyboard can reach, in reading order.
func homeTargets(l *lab) []dashboard.Target {
	sh := l.sh()
	return sh.homePage(sh.mainInner().W, dashboard.Away).Targets
}

// finishedRun puts a run that ended in the history of the lab.
func finishedRun(l *lab, argv ...string) {
	sh := l.sh()
	r := runview.NewRun(argv, nil, l.clock)
	r.End(0, false, l.clock)
	sh.runs = append(sh.runs, r)
	l.m = sh
}

func TestAShortHomeScrollsAndSaysWhatIsCut(t *testing.T) {
	l := newHomeLab(t, 120, 14)
	l.has(moreBelow, "2 checks fail")
	l.hasNot("Compose a stack", moreAbove)
	for i := 0; i < 5 && !strings.Contains(l.view(), "Compose a stack"); i++ {
		l.keys("pgdown")
		l.has(moreAbove)
	}
	l.has("Compose a stack", "Add a project")
	l.hasNot(moreBelow)
	l.keys("pgup", "pgup", "pgup")
	l.has("2 checks fail", moreBelow)
	l.hasNot(moreAbove)
}

func TestCtrlDAndCtrlUAndTheWheelScrollHome(t *testing.T) {
	l := newHomeLab(t, 120, 14)
	l.keys("ctrl+d")
	l.has(moreAbove)
	l.hasNot("2 checks fail")
	l.keys("ctrl+u")
	l.has("2 checks fail")
	l.hasNot(moreAbove)
	x, y := l.find("2 checks fail")
	l.send(wheel(x, y, 1))
	if l.sh().homeTop == 0 {
		t.Fatal("the wheel did not scroll Home")
	}
	l.send(wheel(x, y, -1))
	if l.sh().homeTop != 0 {
		t.Fatalf("the wheel up left Home at line %d", l.sh().homeTop)
	}
	// The wheel never moves the keyboard.
	if l.sh().homeSel != 0 {
		t.Fatalf("the wheel moved the keyboard to %d", l.sh().homeSel)
	}
}

func TestTheArrowsMoveThroughTheButtonsOfHomeInReadingOrder(t *testing.T) {
	l := newHomeLab(t, 160, 34)
	l.has("[Fix (Enter)]", "[Open Doctor (3)]")
	l.keys("down")
	l.has("[Fix (F)]", "[Open Doctor (Enter)]")
	l.keys("right")
	if l.sh().homeSel != 2 {
		t.Fatalf("Right moved to %d", l.sh().homeSel)
	}
	l.keys("left", "up")
	if l.sh().homeSel != 0 {
		t.Fatalf("Left and Up went back to %d", l.sh().homeSel)
	}
	l.keys("up", "left")
	if l.sh().homeSel != 0 {
		t.Fatalf("the first target is the end: %d", l.sh().homeSel)
	}
	l.keys("end")
	last := len(homeTargets(l)) - 1
	if l.sh().homeSel != last {
		t.Fatalf("End went to %d, want %d", l.sh().homeSel, last)
	}
	l.keys("down", "right")
	if l.sh().homeSel != last {
		t.Fatalf("past the last target the keyboard stays: %d", l.sh().homeSel)
	}
	l.has("[Compose a stack (Enter)]")
	l.keys("j", "k", "home")
	if l.sh().homeSel != 0 {
		t.Fatalf("Home went to %d", l.sh().homeSel)
	}
}

func TestEnterPressesTheButtonTheKeyboardIsOn(t *testing.T) {
	l := newHomeLab(t, 160, 34)
	l.keys("down", "enter") // Open Doctor
	if l.sh().tab != commands.TabDoctor {
		t.Fatalf("Enter on Open Doctor showed %q", l.sh().tab)
	}
	l = newHomeLab(t, 160, 34)
	l.keys("enter") // Fix: the first button, as before
	if l.sh().fix == nil {
		t.Fatal("Enter on the first button did not open the fix")
	}
	l = newHomeLab(t, 160, 34)
	l.keys("end", "enter")                // Compose a stack
	l.has("Could not read the catalogue") // the lab has no smith to answer: the command ran
}

func TestEnterOnARunOfHomeOpensItInTheRunPanel(t *testing.T) {
	l := newHomeLab(t, 160, 34)
	finishedRun(l, "run", "test")
	l.has("Recent runs")
	for i := 0; i < 20 && homeTargets(l)[min(l.sh().homeSel, len(homeTargets(l))-1)].Command != ""; i++ {
		l.keys("down")
	}
	if homeTargets(l)[l.sh().homeSel].Command != "" {
		t.Fatal("no run among the targets")
	}
	l.keys("enter")
	if l.sh().tab != commands.TabProjects || !l.sh().runShown() {
		t.Fatalf("tab %q, run shown %v", l.sh().tab, l.sh().runShown())
	}
}

func TestTheKeyboardTargetStaysInViewWhileHomeScrolls(t *testing.T) {
	l := newHomeLab(t, 120, 14)
	l.keys("end")
	l.has("[Compose a stack (Enter)]", moreAbove)
	l.hasNot(moreBelow)
	l.keys("home")
	l.has("[Fix (Enter)]")
	l.hasNot(moreAbove)
	// Each step keeps the target on screen, under no marker.
	for i := range homeTargets(l) {
		l.keys("down")
		if i < len(homeTargets(l))-1 && !strings.Contains(l.view(), "(Enter)]") {
			t.Fatalf("after %d steps the target is out of view:\n%s", i+1, l.view())
		}
	}
}

func TestAClickOnAScrolledHomeHitsTheButtonUnderIt(t *testing.T) {
	l := newHomeLab(t, 120, 14)
	l.keys("pgdown")
	l.has("[Open Release")
	l.clickText("[Open Release")
	if l.sh().tab != commands.TabRelease {
		t.Fatalf("the click showed %q", l.sh().tab)
	}
}

func TestTheFooterOffersScrollingOnlyWhenHomeScrolls(t *testing.T) {
	l := newHomeLab(t, 160, 34)
	l.has("move")
	l.hasNot("PgUp/PgDn")
	l = newHomeLab(t, 120, 14)
	l.has("move", "PgUp/PgDn scroll")
}

func TestHomeDrawsAtEveryHeightFromTheFloorUp(t *testing.T) {
	for _, w := range []int{72, 100, 160} {
		for h := minHeight; h <= 30; h++ {
			l := newHomeLab(t, w, h)
			lines := l.lines()
			if len(lines) != h {
				t.Fatalf("%dx%d: %d rows", w, h, len(lines))
			}
			// Page down until the last button shows: nothing stays out of reach.
			for i := 0; i < 20 && !strings.Contains(l.view(), "Compose a stack"); i++ {
				l.keys("pgdown")
			}
			if !strings.Contains(l.view(), "Compose a stack") {
				t.Fatalf("%dx%d: the buttons at the end of Home cannot be reached:\n%s", w, h, l.view())
			}
		}
	}
}
