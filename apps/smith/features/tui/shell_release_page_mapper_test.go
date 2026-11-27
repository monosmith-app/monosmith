package tui

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/release"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/releaseview"
)

func TestTheReleaseViewShowsThePlanTheChecksAndTheActions(t *testing.T) {
	l := newLab(t, true, 150, 44)
	l.keys("6")
	l.has("✓ 2 packages to release · 1 minor · 1 patch", "Package", "Next", "Bump", "Why", "2 releasing commit(s)",
		"Checks before release", "Health: 2 findings are open", "Registry sign-in", "Nothing is published until you confirm the plan.",
		"[Release now (x)]", "[Dry run (p)]", "[Release notes (n)]", "$ smith release · $ smith release --dry-run")
	// The detail beside the plan tells the selected package.
	l.has("web 1.0.0 → 1.1.0", "Why it releases", "Changes", "Tag", "web@1.1.0", "$ smith release --dry-run --notes")
}

func TestReleaseNowAsksFirstAndTheDryRunStartsAtOnce(t *testing.T) {
	l := newLab(t, true, 150, 44)
	l.keys("6")
	l.clickText("Release now (x)")
	l.has("Release: hold back or bump")
	if len(l.started) != 0 {
		t.Fatalf("a click published without a question: %v", l.started)
	}
	l = newLab(t, true, 150, 44)
	l.keys("6")
	l.clickText("Dry run (p)")
	if got := l.lastStarted(); got != "release --dry-run" {
		t.Fatalf("the dry run started %q", got)
	}
	l.has("Preview: smith release --dry-run", "[Release now (x)]")
	l.hasNot("Checks before release")
}

func TestNothingToReleaseHasNoChecksAndNoButtons(t *testing.T) {
	l := newLab(t, true, 150, 44)
	l.keys("6")
	l.releaseRead(releaseview.Result{Plan: release.Plan{Unchanged: []string{"sdk"}}})
	l.hasNot("Checks before release", "Release now (x)")
	if l.sh().hasDetail() {
		t.Fatal("an empty plan has no package to tell in a detail")
	}
}

func TestOnANarrowScreenEnterOpensThePackageOfTheReleaseAndEscGoesBack(t *testing.T) {
	l := newLab(t, true, 80, 30)
	l.keys("6")
	l.hasNot("a1b2c3d")
	l.keys("enter")
	l.has("‹ Release / web 1.0.0 → 1.1.0", "a1b2c3d feat(ui): dark mode", "Esc back")
	if len(l.started) != 0 {
		t.Fatalf("Enter opened the package, it ran %v", l.started)
	}
	l.keys("esc")
	l.hasNot("‹ Release /", "a1b2c3d")
	l.has("Checks before release")
	l.keys("down", "enter")
	l.has("‹ Release / kit 0.2.0 → 0.2.1", "no commit of its own releases it")
}

func TestOnANarrowScreenEnterOpensTheDependencyAndEscGoesBack(t *testing.T) {
	l := newLab(t, true, 80, 30)
	l.keys("5", "down", "enter")
	l.has("‹ Dependencies / zod", "Used by", "$ smith deps up zod", "Esc back")
	if len(l.started) != 0 || l.sh().confirm != nil {
		t.Fatal("Enter opens the package; it neither runs nor asks")
	}
	l.keys("esc")
	l.hasNot("‹ Dependencies /")
	l.has("Patch and minor releases · 2")
}

func TestOnAWideScreenEnterLeavesTheListAsItIs(t *testing.T) {
	for _, tab := range []string{"5", "6"} {
		l := newLab(t, true, 150, 44)
		l.keys(tab, "enter")
		if l.sh().detailFull() || l.sh().confirm != nil || len(l.started) != 0 {
			t.Fatalf("view %s: Enter beside the detail changed something", tab)
		}
	}
}
