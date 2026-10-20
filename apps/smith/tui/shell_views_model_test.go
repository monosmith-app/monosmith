package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

func TestTheProjectsViewShowsStackNeedsAndWhatIsAffected(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.has("4 projects", "1 affected", "Stack", "node-vitest", "Language")
}

func TestTheProjectsViewRunsATaskAndGraphsWhatItNeeds(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("ctrl+b")
	if got := l.lastStarted(); !strings.Contains(got, "build") {
		t.Fatalf("ctrl+b started %q", got)
	}
	l.keys("alt+1")
	l.clickText("Graph: needs and used by")
	if got := l.lastStarted(); got != "graph --focus orders" {
		t.Fatalf("the graph action started %q", got)
	}
	l.has("graph --focus orders") // the command line is printed before it runs
}

func TestThePluginsViewListsBuiltInAndInstalledPluginsWithTheirUse(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+7")
	l.has("Plugins", "4 built-in · 1 installed · 4 on", "go", "node", "python", "vitest", "acme-lint", "language", "orders, kit", "Used by")
	l.has("from     built-in")
}

func TestThePluginsViewTurnsAPluginOnOnlyAfterConfirmingTheCommand(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+7", "down", "down") // python, which is off
	l.has("e runs smith plugin enable python")
	l.keys("e")
	l.has("smith plugin enable python") // the dialog shows the command line
	if len(l.started) != 0 {
		t.Fatalf("started %v before the confirmation", l.started)
	}
	l.keys("y")
	if got := l.lastStarted(); got != "plugin enable python" {
		t.Fatalf("started %q", got)
	}
}

func TestThePluginsViewDisablesAnEnabledPluginAndLeavesAddonsAlone(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+7", "e", "y")
	if got := l.lastStarted(); got != "plugin disable go" {
		t.Fatalf("started %q", got)
	}
	l = newLab(t, true, 140, 34)
	l.keys("alt+7", "down", "down", "down", "e") // vitest is not switchable
	if len(l.started) != 0 || l.sh().confirm != nil {
		t.Fatal("an add-on plugin was switched")
	}
}

func TestTheDepsViewListsNewerReleasesAndThePolicy(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+8")
	l.has("3 with newer releases: 1 patch · 1 minor · 1 major", "axios", "1.6.8", "zod", "react", "web, api",
		"One version", "every production dependency has one version", "axios (node) should be ^1.6.8 everywhere")
}

func TestTheDepsViewUpdatesTheMarkedSelectionWithDepsUp(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+8", "space", "down", "down", "space") // axios and react
	l.has("[x] axios", "[ ] zod", "[x] react", "2 marked")
	l.keys("u")
	l.has("smith deps up axios react")
	if len(l.started) != 0 {
		t.Fatalf("started %v before the confirmation", l.started)
	}
	l.keys("y")
	if got := l.lastStarted(); got != "deps up axios react" {
		t.Fatalf("started %q", got)
	}
}

func TestTheDepsViewUpdatesTheRowUnderTheCursorWhenNothingIsMarked(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+8", "down", "enter", "y")
	if got := l.lastStarted(); got != "deps up zod" {
		t.Fatalf("started %q", got)
	}
}

func TestTheDepsViewMarksAllAndAlignsVersions(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+8", "a")
	l.has("[x] axios", "[x] zod", "[x] react", "3 marked")
	l.keys("a")
	l.has("0 marked")
	l.keys("v", "y")
	if got := l.lastStarted(); got != "deps up --align" {
		t.Fatalf("started %q", got)
	}
}

func TestTheDepsViewSaysWhenThereIsNothingToUpdateAndWhenThePolicyIsOff(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.m = func() Shell {
		sh := l.sh()
		sh.sources.Deps = func() depsLoaded { return depsLoaded{} }
		return sh
	}()
	l.keys("alt+8")
	l.has("Every dependency is on its latest release.", "not required")
	l.keys("u", "v")
	if len(l.started) != 0 || l.sh().confirm != nil {
		t.Fatal("a command was offered with nothing to update")
	}
}

func TestTheDepsViewShowsAFailedLookup(t *testing.T) {
	l := newLab(t, true, 100, 30)
	l.m = func() Shell {
		sh := l.sh()
		sh.sources.Deps = func() depsLoaded { return depsLoaded{err: errString("the registry is unreachable")} }
		return sh
	}()
	l.keys("alt+8")
	l.has("the registry is unreachable")
}

type errString string

func (e errString) Error() string { return string(e) }

func TestRReloadsTheViewsThatLoadFromTheWorkspace(t *testing.T) {
	l := newLab(t, true, 140, 34)
	calls := 0
	l.m = func() Shell {
		sh := l.sh()
		sh.sources.Deps = func() depsLoaded { calls++; return sampleDeps }
		return sh
	}()
	l.keys("alt+8")
	l.keys("r")
	if calls != 2 {
		t.Fatalf("the dependencies were looked up %d times, want 2", calls)
	}
	l.keys("alt+3", "r")
	l.has("go.work lists every module")
}

func TestQuestionMarkOpensTheHelpOfEveryTab(t *testing.T) {
	cases := []struct {
		key  string
		want string
	}{
		{"alt+1", "Show or hide this help"},
		{"alt+7", "Show or hide this help"},
		{"alt+8", "Show or hide this help"},
		{"alt+2", "Reload the view"},
		{"alt+3", "Reload the view"},
		{"alt+4", "Show or hide this help"},
		{"alt+5", "Show or hide this help"},
	}
	for _, c := range cases {
		l := newLab(t, true, 100, 40)
		l.keys(c.key, "?")
		l.keys("end") // the keys every view shares are last
		l.has(c.want, "Esc", "Mouse")
		l.keys("?")
		l.hasNot("Keyboard shortcuts")
	}
}

func TestQuestionMarkIsTextWhileASettingIsBeingEdited(t *testing.T) {
	l := newLab(t, true, 100, 30)
	l.keys("alt+5", "enter")
	sh := l.sh()
	if !sh.editing() {
		t.Skip("the first setting is not text")
	}
	l.keys("?")
	if l.sh().help != nil {
		t.Fatal("the help opened while typing")
	}
}

func TestEveryTabFitsEightyColumns(t *testing.T) {
	l := newLab(t, true, 80, 26)
	l.has("Projects", "Run", "Doctor", "Release", "Plugins", "Deps", "Settings", "Guide")
	for _, key := range []string{"alt+1", "alt+2", "alt+3", "alt+4", "alt+7", "alt+8", "alt+5", "alt+6"} {
		l.keys(key)
		for i, line := range strings.Split(l.sh().View(), "\n") {
			if w := pane.Width(line); w != 80 {
				t.Fatalf("%s: line %d is %d cells wide: %q", key, i, w, line)
			}
		}
		if got := len(strings.Split(l.sh().View(), "\n")); got != 26 {
			t.Fatalf("%s: %d lines", key, got)
		}
	}
}

func TestTheTabsStayReachableBelowEightyColumns(t *testing.T) {
	l := newLab(t, true, 79, 24)
	l.has("Workspace", "Plugins", "Deps", "Guide")
	l.clickText("Deps")
	l.has("3 with newer releases")
	l.clickText("Plugins")
	l.has("built-in")
}

func TestClickingAPluginRowSelectsItAndAMarkBoxMarksADependency(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+7")
	l.clickText("acme-lint")
	l.has("from     /home/me/.smith/plugins")
	l.keys("alt+8")
	l.clickText("zod")
	l.clickText("zod") // a second click on the row marks it
	l.has("[x] zod")
}

func TestTheWheelMovesThroughThePluginsAndDeps(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+7")
	x, y := l.find("vitest")
	l.send(wheel(x+5, y, 1))
	l.has("e runs smith plugin disable node")
}

func TestTheHelpOfPluginsAndDepsListsTheirOwnKeysFirst(t *testing.T) {
	for key, want := range map[string][]string{
		"alt+1": {"Projects view", "Run the selected task"},
		"alt+7": {"Plugins view", "smith plugin enable|disable"},
		"alt+8": {"Deps view", "smith deps up", "--align"},
	} {
		l := newLab(t, true, 100, 40)
		l.keys(key, "?")
		l.has(want...)
	}
}
