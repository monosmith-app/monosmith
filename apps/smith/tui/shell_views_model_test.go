package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

func TestTheProjectsViewShowsStackNeedsAndWhatIsAffected(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.has("4 projects", "1 affected", "Stack", "node-vitest", "Language")
}

func TestTheProjectsViewRunsATaskAndGraphsWhatItNeeds(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("b")
	if got := l.lastStarted(); !strings.Contains(got, "build") {
		t.Fatalf("b started %q", got)
	}
	l.keys("2")
	l.clickText("Graph: needs and used by")
	if got := l.lastStarted(); got != "graph --focus orders" {
		t.Fatalf("the graph action started %q", got)
	}
	l.has("graph --focus orders") // the command line is printed before it runs
}

func TestTheDepsViewListsNewerReleasesAndThePolicy(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("5")
	l.has("3 with newer releases: 1 patch · 1 minor · 1 major", "axios", "1.6.8", "zod", "react", "web, api",
		"One version", "every production dependency has one version", "axios (node) should be ^1.6.8 everywhere")
}

func TestTheDepsViewUpdatesTheMarkedSelectionWithDepsUp(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("5", "space", "down", "down", "space") // axios and react
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
	l.keys("5", "down", "enter", "y")
	if got := l.lastStarted(); got != "deps up zod" {
		t.Fatalf("started %q", got)
	}
}

func TestTheDepsViewMarksAllAndAlignsVersions(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("5", "a")
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
		sh.sources.Deps = func() depsview.Result { return depsview.Result{} }
		return sh
	}()
	l.keys("5")
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
		sh.sources.Deps = func() depsview.Result { return depsview.Result{Err: errString("the registry is unreachable")} }
		return sh
	}()
	l.keys("5")
	l.has("the registry is unreachable")
}

func TestTheDoctorTabGroupsByProjectAndAppliesOneSafeFixAfterConfirming(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("4", "g")
	l.has("workspace", "web  1 failed", "2 checks passed", "x fix all")
	l.hasNot("f fix") // the cursor is on go.work, which has no fix: the key is not offered
	l.keys("f")
	if l.sh().confirm != nil || len(l.started) != 0 {
		t.Fatal("a finding without a safe fix offered a command")
	}
	l.has("no safe fix")
	l.keys("down")
	l.has("f fix")
	l.keys("f") // web's Docker finding
	l.has("Apply one doctor fix", "$ smith doctor --fix --only 'Docker registries of web agree' --yes")
	if len(l.started) != 0 {
		t.Fatal("nothing runs before the answer")
	}
	l.keys("y")
	want := "doctor --fix --only Docker registries of web agree --yes"
	if got := l.lastStarted(); got != want {
		t.Fatalf("started %q, want %q", got, want)
	}
}

func TestTheDoctorTabListsFlatUntilGroupingIsOn(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("4")
	l.hasNot("orders  all passed", "web  1 failed")
	l.has("Docker registries of web agree")
	l.keys("g")
	l.has("web  1 failed")
}

func TestTheReleaseTabDrillsDownToTheCommitsAndPrintsTheNotesCommand(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("6")
	l.has("2 packages to release", "web", "1.0.0", "1.1.0", "kit", "0.2.1", "n notes")
	l.hasNot("a1b2c3d")
	l.keys("enter")
	l.has("a1b2c3d feat(ui): dark mode", "9f8e7d6 fix!: cookie parsing", "BREAKING: drops the old cookie format")
	l.keys("down", "enter")
	l.has("dependency web releases (no commit of its own releases it)")
	l.keys("n")
	if got := l.lastStarted(); got != "release --dry-run --notes" {
		t.Fatalf("the notes started %q", got)
	}
	l.has("$ smith release --dry-run --notes")
}

type errString string

func (e errString) Error() string { return string(e) }

func TestRReloadsTheViewsThatLoadFromTheWorkspace(t *testing.T) {
	l := newLab(t, true, 140, 34)
	calls := 0
	l.m = func() Shell {
		sh := l.sh()
		sh.sources.Deps = func() depsview.Result { calls++; return sampleDeps }
		return sh
	}()
	l.keys("5")
	l.keys("r")
	if calls != 2 {
		t.Fatalf("the dependencies were looked up %d times, want 2", calls)
	}
	l.keys("4", "r")
	l.has("go.work lists every module")
}

func TestQuestionMarkOpensTheHelpOfEveryTab(t *testing.T) {
	cases := []struct {
		key  string
		want string
	}{
		{"1", "Do the suggested next action"},
		{"2", "Open the detail of the project"},
		{"3", "Run the last command again"},
		{"4", "Read the view again"},
		{"5", "Update the marked ones"},
		{"6", "Print the release notes"},
	}
	for _, c := range cases {
		l := newLab(t, true, 100, 40)
		l.keys(c.key, "?")
		l.has(c.want) // the keys of the open view come first
		l.keys("end") // the keys every view shares are last
		l.has("? / F1", "Esc", "Mouse")
		l.keys("?")
		l.hasNot("Keyboard shortcuts")
	}
}

func TestEveryTabFitsEightyColumns(t *testing.T) {
	l := newLab(t, true, 80, 26)
	// Nine tabs do not fit 79 cells: the long names are shortened and the strip scrolls.
	l.has("Proj", "Run", "Doctor")
	for _, key := range []string{"1", "2", "3", "4", "5", "6"} {
		l.keys(key)
		for i, line := range strings.Split(l.sh().View(), "\n") {
			if w := pane.Width(line); w != 80 {
				t.Fatalf("%s: line %d is %d cells wide: %q", key, i, w, line)
			}
		}
		if got := len(strings.Split(l.sh().View(), "\n")); got != 26 {
			t.Fatalf("%s: %d lines", key, got)
		}
		// Whatever the strip scrolled to, the open tab is in it.
		open := false
		for _, sp := range l.sh().renderStrip("").Spans {
			open = open || sp.ID == l.sh().tab
		}
		if !open {
			t.Fatalf("%s: the open tab %q is out of view:\n%s", key, l.sh().tab, stripText(l))
		}
	}
}

func TestTheTabsStayReachableBelowEightyColumns(t *testing.T) {
	l := newLab(t, true, 79, 24)
	l.keys("5")
	l.has("Deps", "3 with newer releases")
	l.clickText("Home")
	l.has("Next:")
}

func TestClickingADependencyRowMarksItOnTheSecondClick(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("5")
	l.clickText("zod")
	l.clickText("zod") // a second click on the row marks it
	l.has("[x] zod")
}

func TestTheWheelMovesTheCursorOfDeps(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("5")
	x, y := l.find("react")
	l.send(wheel(x+5, y, 1))
	l.keys("space")
	l.has("[x] zod")
}

func TestTheHelpOfDepsListsItsOwnKeysFirst(t *testing.T) {
	for key, want := range map[string][]string{
		"2": {"Projects view", "Open the detail of the project"},
		"5": {"Deps view", "Update the marked ones", "Bring drifting versions to one"},
	} {
		l := newLab(t, true, 100, 40)
		l.keys(key, "?")
		l.has(want...)
	}
}

func TestAboutShowsTheLogoWhatSmithIsForAndTheDocsLink(t *testing.T) {
	l := newLab(t, true, 100, 40)
	l.keys("ctrl+p").typeText("About").keys("enter")
	l.has("About MonoSmith", "▄████", "scaffolds projects", "https://monosmith-app.github.io/monosmith/README.html", "[ OK ]")
	l.keys("esc")
	l.hasNot("About MonoSmith")
}

func TestAboutClosesOnTheOKButton(t *testing.T) {
	l := newLab(t, true, 100, 40)
	l.keys("ctrl+p").typeText("About").keys("enter")
	l.clickText("[ OK ]")
	l.hasNot("About MonoSmith")
}

func TestTheHeaderNamesTheWorkspaceAndTheStatusBarTheVersion(t *testing.T) {
	l := newLab(t, true, 100, 40)
	if last := l.lines()[39]; !strings.Contains(last, "smith 1.2.3") {
		t.Fatalf("status bar %q", last)
	}
	if first := l.lines()[0]; !strings.Contains(first, "acme") || strings.HasPrefix(first, "╭") {
		t.Fatalf("the first row is the header: %q", first)
	}
}
