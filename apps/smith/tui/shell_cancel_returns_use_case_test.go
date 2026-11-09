package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
)

// depsLab is a lab on the Deps tab with the first package marked.
func depsLab(t *testing.T) *lab {
	t.Helper()
	l := newHomeLab(t, 140, 40)
	l.openTab(commands.TabDeps)
	l.keys("down", " ")
	l.has("1 marked")
	return l
}

// Esc on the confirmation of an install, which a form on the Guide led to,
// shows the Deps tab again with its marks, and not the Guide menu.
func TestEscOnTheInstallConfirmationReturnsToTheDepsTab(t *testing.T) {
	l := depsLab(t)
	l.keys("i")
	l.has("Install a package")
	l.typeText("zod").keys("enter", "enter", "enter")
	l.has("The command that will run:")
	l.keys("esc")
	if got := l.sh().tab; got != commands.TabDeps {
		t.Fatalf("Esc must return to the Deps tab, not %q", got)
	}
	l.has("1 marked", "Package")
	l.hasNot("Run tasks")
}

// Esc on the first question of the form returns to the tab it was opened from.
func TestEscOutOfTheInstallFormReturnsToTheDepsTab(t *testing.T) {
	l := depsLab(t)
	l.keys("i", "esc")
	if got := l.sh().tab; got != commands.TabDeps {
		t.Fatalf("Esc must return to the Deps tab, not %q", got)
	}
	l.has("1 marked")
}

// The same holds when the command is run, then another form is cancelled: the
// place to return to is the tab of the form that was opened last.
func TestTheFormOriginIsTheTabItWasOpenedFrom(t *testing.T) {
	l := depsLab(t)
	l.keys("i", "esc")
	l.openTab(commands.TabProjects)
	l.keys("a")
	l.has("What are you building?")
	l.keys("esc")
	if got := l.sh().tab; got != commands.TabProjects {
		t.Fatalf("Esc must return to Projects, not %q", got)
	}
}

// A plan that smith refuses is shown in a notice, and closing it also returns
// to the tab the form was opened from.
func TestClosingTheNoticeOfAFormReturnsToItsTab(t *testing.T) {
	l := depsLab(t)
	l.answers(map[string]sources.Captured{"install -p . zod --dry-run": {Text: "smith: no network"}})
	l.keys("i")
	l.typeText("zod").keys("enter", "enter", "enter")
	if !strings.Contains(l.view(), "The command that will run:") && !strings.Contains(l.view(), "smith said") {
		t.Fatalf("a dialog must be open:\n%s", l.view())
	}
	l.keys("esc")
	if got := l.sh().tab; got != commands.TabDeps {
		t.Fatalf("closing the dialog must return to the Deps tab, not %q", got)
	}
}

// Esc on the confirmations that a tab opens itself leaves the tab as it was.
func TestEscOnTheConfirmationsOfTheDepsTabStaysOnTheTab(t *testing.T) {
	for _, key := range []string{"u", "o", "x"} {
		t.Run(key, func(t *testing.T) {
			l := depsLab(t)
			l.keys(key, "esc")
			if got := l.sh().tab; got != commands.TabDeps {
				t.Fatalf("%s: Esc must stay on the Deps tab, not %q", key, got)
			}
			l.has("1 marked")
		})
	}
}
