package tui

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
)

func TestForcedReleaseActionRunsBumpAndForce(t *testing.T) {
	a := guide.ForcedReleaseAction([]string{"mvd", "core"})
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

func TestForcedReleaseConfirmationSaysWhatHappens(t *testing.T) {
	title, why := commands.Describe([]string{"release", "--bump", "mvd=patch", "--force"})
	if title != guide.ForcedReleaseTitle || !strings.Contains(why, "tag points at HEAD") || !strings.Contains(why, "cannot be undone") {
		t.Fatalf("%q %q", title, why)
	}
	if lines := forcedReleaseDetails([]string{"release", "--bump", "mvd=patch", "--force"}); !strings.Contains(strings.Join(lines, "\n"), "mvd  patch bump") {
		t.Fatalf("%v", lines)
	}
}
