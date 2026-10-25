package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
)

func footerOf(l *lab) string {
	lines := l.lines()
	return lines[len(lines)-2]
}

func TestThePagesThatLeftTheInterfaceAreCommandsOfThePaletteShownInTheRunTab(t *testing.T) {
	for name, want := range map[string]string{
		"Plugins":           "plugin list",
		"Recipes":           "recipe",
		"Settings":          "config list",
		"Commands and docs": "help",
	} {
		l := newLab(t, true, 140, 30)
		l.keys("ctrl+p").typeText(name).keys("enter")
		if l.lastStarted() != want {
			t.Fatalf("%s started %q, want %q", name, l.lastStarted(), want)
		}
		l.has("$ smith " + want)
	}
}

func TestThePluginsCommandOfThePaletteSaysWhatItDoes(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("ctrl+p").typeText("Plugins")
	l.has("Plugins", "smith plugin list")
}

func TestChangingASettingFromTheCliAsksFirstAndThenRunsInTheRunTab(t *testing.T) {
	l := newLab(t, true, 140, 30)
	sh := l.sh()
	cmd := sh.execute([]string{"config", "set", "ci.devcontainer", "false"}, false)
	l.m = sh
	l.drive(cmd)
	l.has("The command that will run:", "$ smith config set ci.devcontainer false")
	if len(l.started) != 0 {
		t.Fatal("config set ran before the answer")
	}
	l.keys("y")
	if len(l.started) != 1 || l.started[0][0] != "config" {
		t.Fatalf("the confirmed command did not run in the Run tab: %v", l.started)
	}
}

func TestTheRunFooterOffersOnlyTheKeysThatWork(t *testing.T) {
	l := newLab(t, true, 160, 30)
	l.keys("3")
	footer := footerOf(l)
	for _, key := range []string{"follow", "task", "again", "failed", "copy log", "runs"} {
		if strings.Contains(footer, key) {
			t.Fatalf("nothing has run, and the footer offers %q: %q", key, footer)
		}
	}
	l.next = failingRun()
	l.keys("b")
	footer = footerOf(l)
	for _, key := range []string{"r again", "R failed", "Y copy log", "f follow", "task"} {
		if !strings.Contains(footer, key) {
			t.Fatalf("a failed run offers %q: %q", key, footer)
		}
	}
	if strings.Contains(footer, "runs") {
		t.Fatalf("there is one run, so there is no other to show: %q", footer)
	}
	l.keys("r")
	l.next = &scripted{lines: []string{"kit:build ✔ 1ms"}}
	l.keys("r")
	if footer = footerOf(l); strings.Contains(footer, "R failed") || !strings.Contains(footer, "runs") {
		t.Fatalf("a run that passed has no failed tasks, and three runs can be browsed: %q", footer)
	}
}

func TestTheDoctorExpandsACheckInPlaceOnANarrowScreen(t *testing.T) {
	l := newLab(t, true, 90, 30)
	l.keys("4")
	l.has("2 failed", "go.work lists every module")
	l.hasNot("libs/x is missing")
	if !strings.Contains(footerOf(l), "Enter expand") {
		t.Fatalf("Enter expands in the single pane: %q", footerOf(l))
	}
	l.keys("enter")
	if n := strings.Count(l.view(), "libs/x is missing"); n != 1 {
		t.Fatalf("the problem shows %d times:\n%s", n, l.view())
	}
	if n := strings.Count(l.view(), "go work use ./libs/x"); n != 1 {
		t.Fatalf("the remedy shows %d times:\n%s", n, l.view())
	}
}

func TestBesideTheDetailTheDoctorTellsEachCheckOnceAndEnterIsNotOffered(t *testing.T) {
	l := newLab(t, true, 140, 30)
	l.keys("4")
	if strings.Contains(footerOf(l), "Enter") {
		t.Fatalf("the detail is already beside the list: %q", footerOf(l))
	}
	l.keys("enter")
	if n := strings.Count(l.view(), "libs/x is missing"); n != 1 {
		t.Fatalf("the problem shows %d times:\n%s", n, l.view())
	}
	if n := strings.Count(l.view(), "go work use ./libs/x"); n != 1 {
		t.Fatalf("the remedy shows %d times:\n%s", n, l.view())
	}
}

func TestTheDepsFooterOffersOnlyWhatWorksHere(t *testing.T) {
	l := newLab(t, true, 140, 30)
	l.keys("5")
	if footer := footerOf(l); !strings.Contains(footer, "u update") || !strings.Contains(footer, "v align") {
		t.Fatalf("with something to update: %q", footer)
	}
	sh := l.sh()
	sh.sources.Deps = func() depsview.Result { return depsview.Result{} }
	l.m = sh
	l.keys("r")
	if footer := footerOf(l); strings.Contains(footer, "u update") || strings.Contains(footer, "v align") || !strings.Contains(footer, "r reload") {
		t.Fatalf("with nothing to update: %q", footer)
	}
	l.keys("u")
	l.has("is unavailable: no dependency to update")
}

func TestTheReleaseFooterOffersReleaseOnlyWhenThereIsSomethingToRelease(t *testing.T) {
	l := newLab(t, true, 140, 30)
	l.keys("6")
	if footer := footerOf(l); !strings.Contains(footer, "x release") || !strings.Contains(footer, "n notes") {
		t.Fatalf("with a plan: %q", footer)
	}
	l.has("2 packages to release (1 minor · 1 patch)")
}

func TestLongNamesAreShortenedInTheMiddleInDepsAndRelease(t *testing.T) {
	long := "@acme-platform/design-system-components-react"
	l := newLab(t, true, 100, 30)
	sh := l.sh()
	sh.sources.Deps = func() depsview.Result {
		return depsview.Result{Rows: []depsview.Row{{Name: long, Current: "1.0.0", Latest: "1.0.1", Level: "patch", Projects: []string{"web"}}}}
	}
	l.m = sh
	l.keys("5")
	l.has("@acme-pl")
	l.has("-react")
	l.hasNot(long)
	if !strings.Contains(l.view(), "…") {
		t.Fatalf("the name is cut with an ellipsis in the middle:\n%s", l.view())
	}
}
