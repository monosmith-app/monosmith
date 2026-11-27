package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

func TestTheDepsViewGroupsTheSafeAndTheMajorReleasesWithTheirCounts(t *testing.T) {
	l := newLab(t, true, 150, 44)
	l.keys("5")
	l.has("↑ 3 with newer releases · 1 patch · 1 minor · 1 major", "Patch and minor releases · 2 · safe to take together", "Major releases · 1 · may break your code")
	view := l.view()
	if strings.Index(view, "axios") > strings.Index(view, "Major releases") || strings.Index(view, "react ") < strings.Index(view, "Major releases") {
		t.Fatalf("the safe rows come first, the major one under its own heading:\n%s", view)
	}
	if n := strings.Count(view, "Used by"); n < 2 {
		t.Fatalf("each group has its column heading: %d\n%s", n, view)
	}
}

func TestTheDepsDetailBesideTheListTellsTheSelectedPackage(t *testing.T) {
	l := newLab(t, true, 150, 44)
	l.keys("5")
	l.has("─ axios ─", "1.6.0 → 1.6.8", "web  apps/web", "Registry", "One version", "$ smith deps up axios")
	l.keys("down", "down")
	l.has("─ react ─", "17.0.2 → 18.3.0", "can break your code", "$ smith deps up react --major")
}

func TestTheDepsBoxSaysWhatAnUpdateTakesAndShowsItsCommandLine(t *testing.T) {
	l := newLab(t, true, 150, 44)
	l.keys("5")
	l.has("Nothing marked: u updates axios 1.6.8", "[Update axios (u)]", "[Update all safe (U)]", "[Repair dependencies (H)]")
	l.keys(" ")
	l.has("1 marked · axios 1.6.8 in 2 projects", "[Update marked (u)]", "$ smith deps up axios")
	l.keys("down", " ")
	l.has("2 marked: axios, zod", "$ smith deps up axios zod")
}

func TestTheUpdateButtonAsksFirstAndRunsOnlyWhatWasConfirmed(t *testing.T) {
	l := newLab(t, true, 150, 44)
	l.keys("5")
	l.clickText("Update axios (u)")
	l.has("$ smith deps up axios --yes", "Cancel (Esc)")
	if len(l.started) != 0 {
		t.Fatalf("started %v before the confirmation", l.started)
	}
	l.keys("esc")
	if len(l.started) != 0 {
		t.Fatalf("a cancelled update ran: %v", l.started)
	}
	l.clickText("Update axios (u)")
	l.keys("y")
	if got := l.lastStarted(); got != "deps up axios --yes" {
		t.Fatalf("started %q", got)
	}
}

func TestTheUpdateAllSafeButtonOpensTheFormOfDepsUp(t *testing.T) {
	l := newLab(t, true, 150, 44)
	l.keys("5")
	l.clickText("Update all safe (U)")
	if len(l.started) != 0 {
		t.Fatalf("a click ran %v without a question", l.started)
	}
	// The lab has no form for deps up: the click says so, and changes nothing.
	if sh := l.sh(); sh.tab != commands.TabGuide && sh.confirm == nil && !strings.Contains(sh.hint, "deps up") {
		t.Fatalf("the click answered nothing: tab %q hint %q", sh.tab, sh.hint)
	}
}

func TestOnANarrowScreenAClickOnTheSelectedPackageOpensIt(t *testing.T) {
	l := newLab(t, true, 80, 30)
	l.keys("5")
	x, y := l.find("zod ")
	l.send(click(x, y)).send(letGo(x, y))
	if l.sh().detailFull() || l.sh().deps.Cursor != 1 {
		t.Fatalf("the first click selects zod: cursor %d", l.sh().deps.Cursor)
	}
	l.send(click(x, y)).send(letGo(x, y))
	l.has("‹ Dependencies / zod", "Used by")
	l.keys("esc")
	l.has("Patch and minor releases · 2")
}

func TestTheDepsAndReleaseViewsFitEverySize(t *testing.T) {
	for _, size := range [][2]int{{80, 30}, {120, 40}, {150, 44}} {
		w, h := size[0], size[1]
		for _, steps := range [][]string{{"5"}, {"5", "down", "enter"}, {"6"}, {"6", "enter"}} {
			l := newLab(t, true, w, h)
			l.keys(steps...)
			lines := l.lines()
			if len(lines) > h {
				t.Errorf("%dx%d %v: %d lines", w, h, steps, len(lines))
			}
			for _, line := range lines {
				if pane.Width(line) > w {
					t.Errorf("%dx%d %v: line wider than the screen: %q", w, h, steps, line)
				}
			}
		}
	}
}

func TestTheDryRunButtonShowsItsCommandAndRunsIt(t *testing.T) {
	l := newLab(t, true, 150, 44)
	l.openTab(commands.TabDeps)
	l.has("Dry run (D)", "$ smith deps up axios --dry-run")
	l.keys("D")
	if got := l.lastStarted(); got != "deps up axios --dry-run" {
		t.Fatalf("D runs the dry run at once, with no question: %q", got)
	}
	if l.sh().confirm != nil {
		t.Fatal("a dry run writes nothing, so it asks nothing")
	}
}

func TestTheDryRunFitsEveryScreen(t *testing.T) {
	for _, size := range [][2]int{{80, 30}, {120, 40}, {150, 44}} {
		l := newLab(t, true, size[0], size[1])
		l.openTab(commands.TabDeps)
		lines := l.lines()
		if len(lines) != size[1] {
			t.Fatalf("%v: %d rows", size, len(lines))
		}
		for y, line := range lines {
			if pane.Width(line) != size[0] {
				t.Fatalf("%v: row %d is %d cells: %q", size, y, pane.Width(line), line)
			}
		}
		l.has("Dry run (D)")
	}
}
