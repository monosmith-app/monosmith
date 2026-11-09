package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAWorkspaceOpensOnHomeAndChecksItselfInTheBackground(t *testing.T) {
	sh := NewShell(Config{Theme: theme.Default(), Sources: &sources.Sources{Doctor: func() doctorview.Result { return doctorview.Result{} }}})
	if sh.tab != commands.TabHome {
		t.Fatalf("opens on %q", sh.tab)
	}
	if !sh.doctor.Loading || sh.Init() == nil {
		t.Fatal("doctor runs from the start, before anyone opens the Doctor view")
	}
	if !strings.Contains(pane.Strip(sh.healthSummary()), "checking") {
		t.Fatalf("the header says it is checking: %q", sh.healthSummary())
	}
}

func TestHomeLeadsWithTheFirstThingToFixAndItsKey(t *testing.T) {
	l := newHomeLab(t, 230, 34)
	// The header badge is filled without the Doctor view having been opened.
	if first := l.lines()[0]; !strings.Contains(first, "×2") {
		t.Fatalf("the header carries the health: %q", first)
	}
	l.has("2 checks fail, 1 can be fixed now", "go.work lists every module · Docker registries of web agree", "[Fix (Enter)]", "[Open Doctor (3)]",
		"1 project is affected by your changes", "[Test what changed (w)]", "Updates not looked at yet", "[Open Deps (4)]", "acme · main · 4 projects")
	l.hasNot("Where to go", "Health", "nothing has run yet", "Next:")
	lines := l.lines()
	var title, fix, affected int
	for i, line := range lines {
		switch {
		case strings.Contains(line, "2 checks fail, 1 can"):
			title = i
		case strings.Contains(line, "[Fix (Enter)]"):
			fix = i
		case strings.Contains(line, "is affected by your changes"):
			affected = i
		}
	}
	if title == 0 || fix <= title || affected <= fix {
		t.Fatalf("the first item (row %d, buttons %d) comes before the others (row %d)", title, fix, affected)
	}
	footer := lines[len(lines)-2]
	for _, want := range []string{"next (Enter)", "build all 3 (B)", "changed (w)", "doctor (d)", "1-6 views"} {
		if !strings.Contains(footer, want) {
			t.Errorf("the Home footer lacks %q: %q", want, footer)
		}
	}
	if strings.Contains(footer, "follow (f)") || strings.Contains(footer, "fix all (F)") {
		t.Fatalf("keys of other views on Home: %q", footer)
	}
}

func TestEnterAndAClickDoTheSuggestedAction(t *testing.T) {
	l := newHomeLab(t, 140, 34)
	// With a fix ready the suggestion opens the fix wizard, not the Doctor
	// tab and not a non-interactive `doctor --fix`.
	l.keys("enter")
	if l.sh().fix == nil || l.sh().tab != commands.TabHome || len(l.started) != 0 {
		t.Fatalf("Enter did the next action: wizard %v, tab %q, started %v", l.sh().fix != nil, l.sh().tab, l.started)
	}
	l.has("Fix · review", "Apply (enter)")
	l = newHomeLab(t, 140, 34)
	l.clickText("[Fix (Enter)]")
	if l.sh().fix == nil {
		t.Fatal("a click on the fix button did not open the fix wizard")
	}
	// Without a fix doctor can make, it goes to the Doctor tab as before.
	l = newHomeLab(t, 140, 34)
	sh := l.sh()
	sh.doctor.Report.Findings[1].Fix, sh.doctor.Report.Fixes = nil, nil
	l.m = sh
	l.has("2 checks fail", "[Open Doctor (Enter)]")
	l.hasNot("[Fix (")
	l.keys("enter")
	if l.sh().tab != commands.TabDoctor || l.sh().fix != nil {
		t.Fatalf("with nothing to fix Enter opened %q", l.sh().tab)
	}
}

func TestAClickOnAButtonRunsItsCommandAndAClickElsewhereOnTheItemDoesNothing(t *testing.T) {
	l := newHomeLab(t, 140, 34)
	l.clickText("[Test what changed")
	if l.lastStarted() != "run lint test --affected" {
		t.Fatalf("the button started %q", l.lastStarted())
	}

	l = newHomeLab(t, 140, 34)
	l.clickText("[Open Doctor")
	if l.sh().tab != commands.TabDoctor {
		t.Fatalf("the second button of the first item opened %q", l.sh().tab)
	}

	l = newHomeLab(t, 140, 34)
	l.clickText("is affected by your changes")
	l.clickText("go.work lists every module")
	l.clickText("2 checks fail, 1 can")
	if len(l.started) != 0 || l.sh().tab != commands.TabHome || l.sh().fix != nil || l.sh().confirm != nil {
		t.Fatalf("a click on the text did something: tab %q, started %v", l.sh().tab, l.started)
	}
	x, y := l.find("[Fix (Enter)]")
	l.send(click(x+len("[Fix (Enter)]")+1, y)) // the gap between two buttons
	if l.sh().fix != nil || l.sh().tab != commands.TabHome {
		t.Fatal("a click between two buttons did something")
	}
}

func TestTheButtonUnderThePointerShowsItsHoverState(t *testing.T) {
	l := newHomeLab(t, 140, 34)
	x, y := l.find("[Test what changed")
	rest := strings.Split(l.sh().View(), "\n")[y]
	l.send(motion(x+2, y))
	hover := strings.Split(l.sh().View(), "\n")[y]
	if hover == rest {
		t.Fatal("the button looks the same under the pointer")
	}
	l.send(motion(0, 0))
	if got := strings.Split(l.sh().View(), "\n")[y]; got != rest {
		t.Fatal("the button keeps its hover state after the pointer left")
	}
}

func TestHomeKeysDoWhatTheirHintsSay(t *testing.T) {
	l := newHomeLab(t, 140, 34)
	l.keys("w")
	if l.lastStarted() != "run lint test --affected" {
		t.Fatalf("w started %q", l.lastStarted())
	}
}

// healthyShell is a workspace with nothing to fix, nothing changed and no
// release to make.
func healthyShell(t *testing.T, affected bool) *lab {
	t.Helper()
	l := newHomeLab(t, 140, 34)
	sh := l.sh()
	projects := sampleProjects
	if !affected {
		projects = nil
		for _, p := range sampleProjects {
			p.Affected = false
			projects = append(projects, p)
		}
	}
	sh.sources.Projects = func() sources.ProjectsLoaded { return sources.ProjectsLoaded{Projects: projects, Branch: "main"} }
	sh.sources.Release = func() releaseview.Result { return releaseview.Result{} }
	sh.sources.Doctor = func() doctorview.Result {
		return doctorview.Result{Report: doctor.Report{Passed: 3, Findings: []doctor.Finding{{Finding: protocol.Finding{Check: "git is installed", OK: true}}}}}
	}
	l.m = sh
	l.keys("r")
	return l
}

func TestWhenAllIsWellHomeSaysSoAndOffersABuild(t *testing.T) {
	l := healthyShell(t, false)
	l.has("All good", "[Build all (Enter)]", "Updates not looked at yet")
	if first := l.lines()[0]; strings.Contains(first, "×") || !strings.Contains(first, "✓") {
		t.Fatalf("the header shows a clean bill: %q", first)
	}
	l.keys("enter")
	if l.lastStarted() != "run build" {
		t.Fatalf("Enter started %q", l.lastStarted())
	}
}

func TestAffectedProjectsComeBeforeABuild(t *testing.T) {
	l := healthyShell(t, true)
	l.has("1 project is affected by your changes", "[Test what changed (Enter)]")
	l.hasNot("All good")
}

func TestOutdatedDependenciesAreCountedOnceTheyAreKnown(t *testing.T) {
	l := healthyShell(t, false)
	l.keys("4", "1")
	l.has("3 dependencies have newer releases", "[Open Deps (Enter)]")
	sh := l.sh()
	sh.sources.Deps = func() depsview.Result { return depsview.Result{} }
	l.m = sh
	l.keys("4", "r", "1")
	l.has("All good")
	l.hasNot("newer releases", "Updates not looked at yet")
}

func TestHomeListsTheLastRunsAndAClickOpensOne(t *testing.T) {
	l := newLab(t, true, 140, 30)
	l.keys("b")
	l.keys("1")
	l.has("Recent runs", "✓ build -p orders")

	l = healthyShell(t, false)
	l.next = failingRun()
	l.keys("B")
	l.keys("1")
	l.has("orders:build failed", "[See why (Enter)]", "[Run it again]", "× build")

	l.clickText("× build")
	if sh := l.sh(); sh.tab != commands.TabProjects || !sh.runShown() {
		t.Fatalf("a click on the run opened %q (run panel %v)", sh.tab, sh.runShown())
	}
	l.has("orders:build")
}

func TestTheOpenRunAndRunAgainButtonsOfAFailedRun(t *testing.T) {
	l := healthyShell(t, false)
	l.next = failingRun()
	l.keys("B")
	l.keys("1")
	l.clickText("[Run it again")
	if got := l.lastStarted(); got != "run build" {
		t.Fatalf("run again started %q", got)
	}
	l = healthyShell(t, false)
	l.next = failingRun()
	l.keys("B")
	l.keys("1")
	l.keys("enter")
	if l.sh().tab != commands.TabProjects || !l.sh().runShown() {
		t.Fatalf("Enter on a failed run opened %q", l.sh().tab)
	}
}

func TestTheReleaseItemAppearsWhenThePlanArrivesAndHomeNeverWaitsForIt(t *testing.T) {
	l := newHomeLab(t, 140, 34)
	sh := l.sh()
	sh.release, sh.releaseLoading = releaseview.Model{}, true
	l.m = sh
	// The plan is still being read: Home shows everything else at once.
	l.has("2 checks fail", "[Fix (Enter)]")
	l.hasNot("Ready to release")
	l.keys("enter")
	if l.sh().fix == nil {
		t.Fatal("keys wait for the release plan")
	}
	l = newHomeLab(t, 140, 34)
	sh = l.sh()
	sh.release, sh.releaseLoading = releaseview.Model{}, true
	l.m = sh
	l.send(releaseview.Result{Plan: release.Plan{Items: []release.Item{
		{Project: "web", Current: "1.0.0", Next: "1.1.0", Commits: []string{"a", "b", "c"}},
		{Project: "kit", Current: "0.2.0", Next: "0.2.1", Commits: []string{"a"}},
	}}})
	l.has("Ready to release: 2 packages, web 1.0.0 → 1.1.0 (3 commits)", "[Open Release (5)]")
	l.clickText("[Open Release")
	if l.sh().tab != commands.TabRelease {
		t.Fatalf("Open Release opened %q", l.sh().tab)
	}
}

func TestAFailedPlanShowsNothingOnHome(t *testing.T) {
	l := newHomeLab(t, 140, 34)
	l.send(releaseview.Result{Err: errText("no git")})
	l.hasNot("Ready to release", "no git")
	l.has("[Fix (Enter)]")
}

func TestTheReleasePlanIsReadOnceForHomeAndTheReleaseTab(t *testing.T) {
	l := newHomeLab(t, 140, 34)
	calls := 0
	sh := l.sh()
	inner := sh.sources.Release
	sh.sources.Release = func() releaseview.Result { calls++; return inner() }
	sh.release, sh.releaseLoading = releaseview.Model{}, true // the read of the start is still going
	l.m = sh
	l.openTab(commands.TabRelease)
	if calls != 0 {
		t.Fatalf("the Release tab started a second read of the plan (%d)", calls)
	}
	l.send(releaseview.Result{Plan: release.Plan{Items: []release.Item{{Project: "web", Current: "1.0.0", Next: "1.1.0"}}}})
	l.has("web")
	if calls != 0 || l.sh().releaseLoading {
		t.Fatalf("calls %d, loading %v", calls, l.sh().releaseLoading)
	}
}

func TestTheStartReadsTheReleasePlanInTheBackground(t *testing.T) {
	read := false
	sh := NewShell(Config{Theme: theme.Default(), Sources: &sources.Sources{
		Doctor:  func() doctorview.Result { return doctorview.Result{} },
		Release: func() releaseview.Result { read = true; return releaseview.Result{} },
	}})
	if !sh.releaseLoading {
		t.Fatal("the plan is not marked as being read")
	}
	if sh.Init() == nil || read {
		t.Fatal("Init must return the read as a command, not run it")
	}
}

func TestHomeFitsTheLeastWidthAndReloadChecksAgain(t *testing.T) {
	l := newHomeLab(t, 72, 24)
	l.has("[Fix (Enter)]")
	for i, line := range l.lines() {
		if w := len([]rune(line)); w > 72 {
			t.Fatalf("line %d is %d cells: %q", i, w, line)
		}
	}
	calls := 0
	sh := l.sh()
	inner := sh.sources.Doctor
	sh.sources.Doctor = func() doctorview.Result { calls++; return inner() }
	l.m = sh
	l.keys("r")
	if calls != 1 {
		t.Fatalf("r on Home checked %d times", calls)
	}
}

func TestHomeNeverSaysAllGoodWhileTheDoctorWarns(t *testing.T) {
	l := healthyShell(t, false)
	sh := l.sh()
	sh.sources.Doctor = func() doctorview.Result {
		return doctorview.Result{Report: doctor.Report{Passed: 3, Warnings: 1, Findings: []doctor.Finding{
			{Finding: protocol.Finding{Check: "golangci-lint matches CI", Warning: true, Problem: "golangci-lint is older than CI"}}}}}
	}
	l.m = sh
	l.keys("r")
	l.has("1 warning: golangci-lint is older than CI", "[Open Doctor (Enter)]")
	l.hasNot("All good")
}
