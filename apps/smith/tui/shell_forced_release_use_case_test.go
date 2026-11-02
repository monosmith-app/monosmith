package tui

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
)

func TestForcedReleaseActionRunsBumpAndForce(t *testing.T) {
	a := guide.ForcedReleaseAction([]string{"mvd", "core"}, nil)
	if len(a.Steps[0].Options) != 2 || a.Steps[0].Options[0].Value != "mvd" {
		t.Fatalf("%+v", a.Steps[0].Options)
	}
	got := a.Argv(map[string][]string{"package": {"mvd"}, "level": {"patch"}})
	if want := []string{"release", "--bump", "mvd=patch", "--force"}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if !isForcedRelease(got) || isForcedRelease([]string{"release", "--yes"}) {
		t.Fatal("isForcedRelease")
	}
}

func TestForcedReleaseDoesNotOfferAPackageThatWasNeverReleased(t *testing.T) {
	a := guide.ForcedReleaseAction([]string{"mvd", "fresh"}, []string{"fresh"})
	if len(a.Steps[0].Options) != 1 || a.Steps[0].Options[0].Value != "mvd" {
		t.Fatalf("a package smith refuses was offered: %+v", a.Steps[0].Options)
	}
	if note := a.Steps[0].Note; !strings.Contains(note, "fresh (never released)") || !strings.Contains(note, `version(fresh)[1.0.0]`) {
		t.Fatalf("the form does not say why, or how to set the first version: %q", note)
	}
}

func TestForcedReleaseWithOnlyNeverReleasedPackagesShowsTheCommitInANotice(t *testing.T) {
	l := newLab(t, true, 110, 32)
	l.keys("6") // the plan is read
	sh := l.sh()
	sh.release.Plan.Unchanged = []string{"fresh"}
	sh.release.Plan.NeverReleased = []string{"fresh"}
	l.m = sh
	l.keys("f")
	l.has("No package to release without a commit", "never released: fresh", `git commit --allow-empty -m "version(fresh)[1.0.0]"`)
}

func TestTheReleaseFooterOffersTheCommitsOnlyWhenSomethingReleases(t *testing.T) {
	l := newLab(t, true, 140, 40)
	l.keys("6")
	l.has("Enter commits")
	sh := l.sh()
	sh.release.Plan.Items = nil
	l.m = sh
	l.keys("down") // draws it again
	l.hasNot("Enter commits")
}

func TestForcedReleaseConfirmationSaysWhatHappens(t *testing.T) {
	title, why := commands.Describe([]string{"release", "--bump", "mvd=patch", "--force"})
	if title != guide.ForcedReleaseTitle || !strings.Contains(why, "tag points at HEAD") || !strings.Contains(why, "cannot be undone") {
		t.Fatalf("%q %q", title, why)
	}
	if lines := forcedReleaseDetails([]string{"release", "--bump", "mvd=patch", "--force"}); !strings.Contains(strings.Join(lines, "\n"), "mvd  patch bump") {
		t.Fatalf("%v", lines)
	}
}
