package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
)

// answers makes the lab's captured commands answer by their command line.
func (l *lab) answers(by map[string]sources.Captured) {
	l.capture = func(argv []string) sources.Captured {
		if res, ok := by[strings.Join(argv, " ")]; ok {
			return res
		}
		return defaultCapture(argv)
	}
}

// defaultCapture is what a command the test did not script prints: success,
// and no choices for `smith choices`.
func defaultCapture(argv []string) sources.Captured {
	if len(argv) > 0 && argv[0] == "choices" {
		return sources.Captured{OK: true, Text: "[]"}
	}
	return sources.Captured{OK: true}
}

func TestUndoShowsWhatItRevertsFirstAndRunsOnlyAfterYes(t *testing.T) {
	l := newHomeLab(t, 120, 40)
	l.answers(map[string]sources.Captured{"undo --dry-run": {OK: true, Text: "Would revert 3b2c1a0 feat: add web\n  apps/web/package.json"}})
	l.keys("ctrl+p").typeText("Undo the last").keys("enter")
	l.has("Undo the last change?", "Would revert 3b2c1a0", "apps/web/package.json", "smith undo --yes")
	if len(l.started) != 0 {
		t.Fatalf("nothing may run before the answer: %v", l.started)
	}
	l.keys("esc")
	if len(l.started) != 0 {
		t.Fatalf("Esc must leave the change in place: %v", l.started)
	}
	l.keys("ctrl+p").typeText("Undo the last").keys("enter", "enter")
	if got := l.lastStarted(); got != "undo --yes" {
		t.Fatalf("started %q", got)
	}
}

func TestUndoThatCannotRunTellsWhySmithRefused(t *testing.T) {
	l := newHomeLab(t, 120, 40)
	l.answers(map[string]sources.Captured{"undo --dry-run": {Text: "✖ smith has no operation to undo here\n  → set commit_operations = true under [workspace] in smith.toml"}})
	l.keys("ctrl+p").typeText("Undo the last").keys("enter")
	l.has("Nothing to undo", "smith has no operation to undo here", "commit_operations = true", "OK")
	l.keys("enter")
	l.hasNot("Nothing to undo")
	if len(l.started) != 0 {
		t.Fatalf("a refused undo runs nothing: %v", l.started)
	}
}

func TestAddAsksForTheChoicesThenShowsThePlanBeforeWriting(t *testing.T) {
	l := newLab(t, true, 120, 40)
	l.answers(map[string]sources.Captured{
		"choices --json": {OK: true, Text: `[{"axis":"framework","name":"react","description":"a React app","language":"go"},{"axis":"test","name":"vitest","description":"tests","language":"go","default":true}]`},
		"add app api --lang go --with react --fresh --dry-run": {OK: true, Text: "Plan: add app api\n  Project api\n    create  apps/api/  (follows web: same test runner)"},
	})
	l.keys("a")
	l.keys("enter", "enter") // an app, in the workspace's main language (go)
	l.typeText("api").keys("enter")
	l.has("Build it from choices", "framework: react", "test: vitest")
	l.keys("space", "enter") // react
	l.has("Should it do what the other projects here do?", "smith add --fresh")
	l.keys("down", "enter")
	l.has("Add api?", "Nothing is written yet", "follows web", "smith add app api --lang go --with react --fresh")
	for _, argv := range l.started {
		t.Fatalf("the plan must come before any write: %v", argv)
	}
	l.keys("enter")
	if got := l.lastStarted(); got != "add app api --lang go --with react --fresh" {
		t.Fatalf("started %q", got)
	}
}

func TestAddWithAWrongCombinationShowsTheCLIErrorInsteadOfCrashing(t *testing.T) {
	l := newLab(t, true, 120, 40)
	l.answers(map[string]sources.Captured{
		"add app api --lang go --fresh --dry-run": {Text: "✖ svelte needs the node language\n  → pick a node project or another framework"},
	})
	l.keys("a", "enter", "enter")
	l.typeText("api").keys("enter") // no choices for go, then the sibling question
	l.keys("down", "enter")
	l.has("Cannot add api like that", "svelte needs the node language", "pick a node project")
	l.keys("enter")
	l.hasNot("Cannot add api")
	if len(l.started) != 0 {
		t.Fatalf("a refused plan writes nothing: %v", l.started)
	}
}

func TestAFinishedAddSaysItCanBeUndone(t *testing.T) {
	l := newLab(t, true, 140, 40)
	l.keys("a", "enter", "enter")
	l.typeText("api").keys("enter", "enter") // follow the siblings
	l.keys("enter")                          // the plan, then Run
	l.has("can be undone", "Undo the last change")
}

func TestSearchingThePluginMarketplaceInstallsTheOnePicked(t *testing.T) {
	l := newLab(t, true, 120, 40)
	l.answers(map[string]sources.Captured{
		"plugin search deploy --json": {OK: true, Text: `[{"name":"deployer","newest":"0.2.0","description":"Deploys to the moon","roles":["deploy"]},{"name":"widget","newest":"1.0.0","description":"Scaffolds widgets"}]`},
	})
	sh := l.sh()
	cmd := sh.openPluginSearch("deploy")
	l.m = sh
	l.drive(cmd)
	l.has("Which plugin?", "deployer", "v0.2.0", "Deploys to the moon", "widget")
	l.keys("enter")
	l.has("Install a plugin", "smith plugin install deployer")
	l.keys("enter")
	if got := l.lastStarted(); got != "plugin install deployer" {
		t.Fatalf("started %q", got)
	}
}

func TestSearchingWithNoMatchSaysHowToWidenIt(t *testing.T) {
	l := newLab(t, true, 120, 40)
	l.answers(map[string]sources.Captured{"plugin search zzz --json": {OK: true, Text: "[]"}})
	sh := l.sh()
	cmd := sh.openPluginSearch("zzz")
	l.m = sh
	l.drive(cmd)
	l.has("No plugin matches", "leave the field empty")
}

func TestTheReportIsWrittenToATemporaryFileAndOpened(t *testing.T) {
	l := newHomeLab(t, 140, 40)
	l.keys("ctrl+p").typeText("Open the report").keys("enter")
	got := l.lastStarted()
	if !strings.HasPrefix(got, "report -o ") || !strings.HasSuffix(got, ".html") {
		t.Fatalf("started %q", got)
	}
	if len(l.opened) != 1 || !strings.HasSuffix(got, l.opened[0]) {
		t.Fatalf("the report must be opened: %v", l.opened)
	}
	l.has("Report opened in your browser", "smith-report-")
}

func TestReportTargetReadsTheOutputFile(t *testing.T) {
	for argv, want := range map[string]string{"report -o a.html": "a.html", "report --output b.html": "b.html", "report": "", "doctor -o x": ""} {
		got, ok := reportTarget(strings.Fields(argv))
		if got != want || ok != (want != "") {
			t.Errorf("%q: %q %v", argv, got, ok)
		}
	}
}
