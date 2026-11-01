package tui

import (
	"strings"
	"testing"
)

func TestTheFooterOfHomeNamesWhatBuildAndBuildAllRunOn(t *testing.T) {
	l := newLab(t, true, 140, 34)
	footer := l.lines()[len(l.lines())-2]
	for _, want := range []string{"b build orders", "B build all 3", "t test orders"} {
		if !strings.Contains(footer, want) {
			t.Fatalf("the footer lacks %q: %q", want, footer)
		}
	}
}

func TestShiftBBuildsEveryProjectAndBBuildsTheSelectedOne(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("b")
	if got := l.lastStarted(); got != "run build -p orders" {
		t.Fatalf("b started %q", got)
	}
	l = newLab(t, true, 140, 34)
	l.keys("B")
	if got := l.lastStarted(); got != "run build" {
		t.Fatalf("B started %q", got)
	}
}

func TestOnHomeTheRunKeysCoverTheWholeWorkspaceAndSayWhichScope(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.openTab("home")
	footer := l.lines()[len(l.lines())-2]
	if !strings.Contains(footer, "b build all 3") || strings.Contains(footer, "build orders") {
		t.Fatalf("the Home footer is %q", footer)
	}
	l.keys("b")
	if got := l.lastStarted(); got != "run build" {
		t.Fatalf("b on Home started %q", got)
	}
	l.has("Build · all 4 projects")
}

func TestOnProjectsTheBuildKeyNamesAndRunsTheSelectedProject(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.openTab("projects")
	l.keys("b")
	if got := l.lastStarted(); got != "run build -p orders" {
		t.Fatalf("b on Projects started %q", got)
	}
	l.has("Build · orders")
}

func TestTheRunTabOffersBuildEverythingAsAButtonAndPointsAtWhatIsOnScreen(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.openTab("run")
	l.has("Nothing has run yet", "Build everything (4)", "Test everything (4)")
	l.hasNot("in the list")
	l.clickText("Build everything")
	if got := l.lastStarted(); got != "run build" {
		t.Fatalf("the button started %q", got)
	}
}

func TestThePaletteInstallOpensTheGuidedFormOnTheProjectOnScreenAndOffersNoRawCommand(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("ctrl+p").typeText("install")
	l.has("Install a package", "Install everything")
	l.hasNot("uninstall >", "[--save-dev]")
	l.keys("enter")
	l.has("Install packages", "Project: ‹ orders ›", "development dependency")
	l.hasNot("Command palette")
	// Nothing to confirm until a package is typed: the form never offers a command that fails.
	l.keys("enter")
	l.has("type at least one package")
	if len(l.started) != 0 {
		t.Fatalf("started %v", l.started)
	}
	l.typeText("zod").keys("enter")
	if got := l.lastStarted(); got != "install -w orders zod" {
		t.Fatalf("started %q", got)
	}
}

func TestThePaletteRemoveAPackageOpensTheSameForm(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("ctrl+p").typeText("remove a").keys("enter")
	l.has("Remove packages", "Project: ‹ orders ›")
}

func TestThePaletteGroupsItsCommandsAndGivesEveryKeyOnce(t *testing.T) {
	l := newLab(t, true, 140, 42)
	l.keys("ctrl+p")
	l.has("Build", "Projects")
	l.has("Build orders", "Build all 3", "smith run build -p orders")
	l.hasNot("uninstall >", "install >")
	seen := map[string]int{}
	for _, e := range l.sh().pal.Entries() {
		if e.Shortcut != "" {
			seen[e.Shortcut]++
		}
	}
	for key, n := range seen {
		if n > 1 {
			t.Errorf("the key %q is listed %d times", key, n)
		}
	}
}
