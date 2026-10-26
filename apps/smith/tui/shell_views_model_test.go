package tui

import (
	"fmt"
	"strconv"
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

func TestTheDoctorTabOffersTheFixOnlyOnTheChecksThatHaveOne(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("4", "g")
	l.has("workspace", "web  1 failed", "2 checks passed", "F fix all")
	l.hasNot("f fix", "Fix:") // the cursor is on go.work, which has no fix: no action, no key
	l.has("Check again")
	l.keys("f")
	if l.sh().fix != nil || len(l.started) != 0 {
		t.Fatal("a finding without a fix opened a wizard")
	}
	l.has("no fix")
	l.keys("down")
	l.has("f fix", "Fix: write web's registry", "writes apps/web/Dockerfile")
	l.keys("f") // web's Docker finding
	l.has("Fix · review", "write web's registry", "apps/web/Dockerfile")
	if len(l.started) != 0 || len(l.appliedFixes()) != 0 {
		t.Fatal("nothing runs before the confirmation")
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

// longDeps is a lab with n outdated dependencies named pkg00, pkg01, ...
func longDeps(t *testing.T, n int) *lab {
	t.Helper()
	l := newLab(t, true, 140, 34)
	sh := l.sh()
	var rows []depsview.Row
	for i := 0; i < n; i++ {
		rows = append(rows, depsview.Row{Name: fmt.Sprintf("pkg%02d", i), Current: "1.0.0", Latest: "1.0.1", Level: "patch", Projects: []string{"web"}})
	}
	sh.sources.Deps = func() depsview.Result { return depsview.Result{Rows: rows} }
	l.m = sh
	l.keys("5")
	l.keys("r")
	return l
}

func (l *lab) shiftClick(x, y int) *lab {
	m := click(x, y)
	m.Shift = true
	return l.send(m)
}

func TestClickingADependencyRowOnlySelectsItAndItsBoxMarksIt(t *testing.T) {
	l := longDeps(t, 8)
	l.clickText("pkg03")
	if d := l.sh().deps; d.Cursor != 3 || len(d.Marked) != 0 {
		t.Fatalf("a click on the name selects the row and marks nothing: cursor %d, marked %v", d.Cursor, d.Marked)
	}
	l.clickText("pkg03")
	if len(l.sh().deps.Marked) != 0 {
		t.Fatal("a second click on the name marked the row")
	}
	x, y := l.find("[ ] pkg03")
	l.send(click(x+1, y))
	l.has("[x] pkg03")
}

func TestTheWheelScrollsTheDepsListAndNeverMovesItsCursor(t *testing.T) {
	l := longDeps(t, 60)
	l.has("pkg00")
	x, y := l.find("pkg05")
	for range 4 {
		l.send(wheel(x+5, y, 1))
	}
	if d := l.sh().deps; d.Cursor != 0 {
		t.Fatalf("the wheel moved the cursor to %d", d.Cursor)
	}
	l.hasNot("pkg00")
	l.hasNot("pkg01 ")
}

func TestClickingARowAfterScrollingDoesNotCenterTheList(t *testing.T) {
	l := longDeps(t, 60)
	x, y := l.find("pkg05")
	for range 4 {
		l.send(wheel(x+5, y, 1))
	}
	first := ""
	for _, line := range l.lines() {
		if i := strings.Index(line, "pkg"); i >= 0 {
			first = line[i : i+5]
			break
		}
	}
	x, _ = l.find("[ ] " + first)
	_, below := l.find("[ ] " + fmt.Sprintf("pkg%02d", mustAtoi(first[3:])+6))
	l.send(click(x+8, below))
	l.has(first)
	if d := l.sh().deps; d.Cursor != mustAtoi(first[3:])+6 {
		t.Fatalf("cursor %d", d.Cursor)
	}
}

func mustAtoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return n
}

func TestShiftClickMarksTheRowsBetweenTwoClicksAndShiftArrowsExtendThem(t *testing.T) {
	l := longDeps(t, 12)
	x, y := l.find("[ ] pkg03")
	l.send(click(x+1, y))
	_, y6 := l.find("[ ] pkg06")
	l.shiftClick(x+1, y6)
	d := l.sh().deps
	for i := 3; i <= 6; i++ {
		if !d.Marked[fmt.Sprintf("pkg%02d", i)] {
			t.Fatalf("pkg%02d is not marked: %v", i, d.Marked)
		}
	}
	if len(d.Marked) != 4 {
		t.Fatalf("marked %v, want pkg03..pkg06", d.Marked)
	}
	l.keys("home", "shift+down", "shift+down")
	d = l.sh().deps
	if len(d.Marked) != 3 || !d.Marked["pkg00"] || !d.Marked["pkg02"] {
		t.Fatalf("shift+down from the top marked %v", d.Marked)
	}
	l.keys("shift+up")
	if d = l.sh().deps; len(d.Marked) != 2 || d.Marked["pkg02"] {
		t.Fatalf("reversing shrinks the marks: %v", d.Marked)
	}
}

func TestTheBoxInTheHeadingMarksEveryRowThenNoneAndShowsSome(t *testing.T) {
	l := longDeps(t, 6)
	x, y := l.find("[ ] Package")
	l.send(click(x+1, y))
	if n := len(l.sh().deps.Marked); n != 6 {
		t.Fatalf("the heading box marked %d of 6", n)
	}
	l.has("[x] Package")
	l.send(click(x+1, y))
	if n := len(l.sh().deps.Marked); n != 0 {
		t.Fatalf("a second click left %d marked", n)
	}
	bx, by := l.find("[ ] pkg02")
	l.send(click(bx+1, by))
	l.has("[-] Package")
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

func TestTheHeaderNamesTheWorkspaceAndTheStatusBarTheVersion(t *testing.T) {
	l := newLab(t, true, 100, 40)
	if last := l.lines()[39]; !strings.Contains(last, "smith 1.2.3") {
		t.Fatalf("status bar %q", last)
	}
	if first := l.lines()[0]; !strings.Contains(first, "acme") || strings.HasPrefix(first, "╭") {
		t.Fatalf("the first row is the header: %q", first)
	}
}

func TestAddingAPackageToTheSelectedProjectAsksWhichAndRunsSmithInstall(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("+")
	l.has("Install packages", "Project: ‹ orders ›", "Packages, separated by spaces", "development dependency")
	l.typeText("zod").keys("enter")
	if got := l.lastStarted(); got != "install -w orders zod" {
		t.Fatalf("started %q", got)
	}
}

func TestTheInstallActionOfTheProjectDetailOpensTheSamePrompt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.has("Install a package…", "Remove a package…")
	l.clickText("Install a package")
	l.has("Install packages")
	l.keys("esc")
	l.hasNot("Install packages")
	if len(l.started) != 0 {
		t.Fatalf("cancelling started %v", l.started)
	}
	l.clickText("Remove a package")
	l.has("Remove packages")
	l.typeText("left-pad").keys("enter")
	if got := l.lastStarted(); got != "uninstall -w orders left-pad" {
		t.Fatalf("started %q", got)
	}
}

func TestAScrollbarShowsOnlyWhenTheListHidesRowsAndFollowsTheWheel(t *testing.T) {
	short := longDeps(t, 8)
	short.hasNot("┃")
	long := longDeps(t, 60)
	long.has("┃")
	thumbRow := func(l *lab) int {
		for y, line := range l.lines() {
			if strings.Contains(line, "┃") {
				return y
			}
		}
		return -1
	}
	before := thumbRow(long)
	x, y := long.find("pkg05")
	for range 6 {
		long.send(wheel(x+5, y, 1))
	}
	if after := thumbRow(long); after <= before {
		t.Fatalf("the thumb stayed at row %d after scrolling down (now %d)", before, after)
	}
}
