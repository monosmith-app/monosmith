package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

func TestTheFixShowsInTheDetailOfHealthReviewApplyThenCheckAgain(t *testing.T) {
	l := withInstallFix(newLab(t, true, 150, 44))
	l.keys("f")
	if !l.sh().inlineFix() || l.sh().modalOpen() {
		t.Fatal("the fix is drawn in the detail, not as a dialog")
	}
	l.has("× Failed", "node_modules is missing", "1 Review ── 2 Apply ── 3 Check again", "Which package manager?", "or in a terminal: smith doctor --fix --only")
	l.keys("2")
	l.has("Fix · review", "$ pnpm install", "Apply (Enter)", "enter apply")
	if len(l.appliedFixes()) != 0 {
		t.Fatal("applied before the person agreed")
	}
	l.keys("enter")
	if got := l.appliedFixes(); len(got) != 1 || got[0].Steps[0].Argv[0] != "pnpm" {
		t.Fatalf("applied %+v", got)
	}
	// The third step: doctor checked again and the check passes now.
	if l.sh().fixed == nil || !l.sh().fixedShown() {
		t.Fatal("the detail does not show the step Check again")
	}
	l.has("✓ Passed", "The check passes now.", "Applied: install the dependencies", "Check again (r)", "r check again")
	before := l.sh().doctor.Report
	l.clickText("Check again (r)")
	if l.sh().doctor.Report == before {
		t.Fatal("Check again did not run the checks again")
	}
	l.has("The check passes now.")
	l.keys("esc")
	if l.sh().fixed != nil {
		t.Fatal("Esc leaves the step Check again")
	}
	l.hasNot("The check passes now.")
}

func TestTheFixThatDidNotClearItsCheckSaysSoInTheDetail(t *testing.T) {
	l := withInstallFix(newLab(t, true, 150, 44))
	l.fixStaysBroken = true
	l.keys("f", "1", "enter")
	if !l.sh().inlineFix() || l.sh().fixed != nil {
		t.Fatal("the failure shows in the detail, in place of the step Check again")
	}
	l.has("The fix ran, but this check still fails", "Retry (r)", "r retry")
}

func TestOnANarrowScreenEnterOpensTheFixScreenAndEscGoesBack(t *testing.T) {
	l := withInstallFix(newLab(t, true, 80, 30))
	l.keys("enter")
	l.has("‹ Health / dependencies are installed", "[Fix (f)]")
	l.keys("f")
	l.has("‹ Health / dependencies are installed", "1 Review ── 2 Apply ── 3 Check again", "1) npm")
	if !l.sh().inlineFix() {
		t.Fatal("the fix shows in the frame of the check")
	}
	l.keys("esc")
	if l.sh().fix != nil {
		t.Fatal("Esc closes the fix")
	}
	l.has("‹ Health / dependencies are installed")
	l.keys("esc")
	l.hasNot("‹ Health /")
	l.has("Failed  2")
}

func TestAClickBesideTheFixLeavesItAndAClickOnItsButtonsWorks(t *testing.T) {
	l := withInstallFix(newLab(t, true, 150, 44))
	l.keys("f")
	l.clickText("2) pnpm")
	l.has("Fix · review")
	x, y := l.find("Failed  2")
	l.send(click(x, y))
	if l.sh().fix != nil {
		t.Fatal("a click on the list left the fix open")
	}
	if len(l.appliedFixes()) != 0 {
		t.Fatal("leaving applied something")
	}
}

func TestTheHealthFixAndThePaletteStayInsideTheScreenAtEverySize(t *testing.T) {
	for _, size := range [][2]int{{80, 30}, {120, 40}, {150, 44}} {
		for _, step := range []string{"list", "fix", "review", "checked", "palette"} {
			name := fmt.Sprintf("%dx%d %s", size[0], size[1], step)
			l := withInstallFix(newLab(t, true, size[0], size[1]))
			if size[0] < 100 {
				l.keys("enter")
			}
			switch step {
			case "fix":
				l.keys("f")
			case "review":
				l.keys("f", "1")
			case "checked":
				l.keys("f", "1", "enter")
			case "palette":
				l.keys("ctrl+k")
			}
			lines := strings.Split(l.sh().View(), "\n")
			if len(lines) != size[1] {
				t.Fatalf("%s: %d rows", name, len(lines))
			}
			for i, row := range lines {
				if w := pane.Width(row); w != size[0] {
					t.Fatalf("%s: row %d is %d wide: %q", name, i, w, pane.Strip(row))
				}
			}
			if step == "review" {
				l.has("Apply (Enter)")
			}
			if step == "palette" {
				l.has("↑/↓ move", "esc close")
			}
		}
	}
}

func TestAClickOnTheCommandLineOfAPaletteEntryRunsIt(t *testing.T) {
	l := newLab(t, true, 80, 30)
	l.keys("ctrl+k").typeText("Build all")
	_, y := l.find("smith run build")
	x, _ := l.find("smith run build")
	l.send(click(x, y))
	if l.sh().palOpen {
		t.Fatal("a click on the command line of an entry chooses it")
	}
	if got := l.lastStarted(); got != "run build" {
		t.Fatalf("started %q", got)
	}
}
