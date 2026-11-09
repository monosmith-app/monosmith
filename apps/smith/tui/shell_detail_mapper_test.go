package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestATaskHintSaysWhatRunsFirstWithoutNxSyntax(t *testing.T) {
	hint := taskHint(projectlist.Task{Name: "package", DependsOn: []string{"^build", "build"}, Cache: true})
	if strings.Contains(hint, "^") || !strings.Contains(hint, "runs after the build of the dependencies and its own build") {
		t.Fatalf("%s", hint)
	}
}

func TestTheStackRowSaysWhatTheStackGives(t *testing.T) {
	if got := stackText("node-jest"); got != "node-jest (ESLint and Jest tasks)" {
		t.Fatalf("%s", got)
	}
	if got := stackText("custom"); got != "custom" {
		t.Fatalf("%s", got)
	}
	if taskHint(projectlist.Task{Name: "deps-check"}) == "a task of this project" {
		t.Fatal("deps-check is unexplained")
	}
}

// detailLabels are the labels of the rows the detail draws, from the text of
// its lines: the first word of a line that has a label and a value.
func detailLabels(s Shell) map[string]bool {
	lines, _ := s.detailPanel().View(s.th, 60, "")
	got := map[string]bool{}
	for _, l := range lines {
		if f := strings.Fields(pane.Strip(l)); len(f) > 1 {
			got[f[0]] = true
		}
	}
	return got
}

func TestBesideTheListTheProjectDetailLeavesOutKindLanguageAndStack(t *testing.T) {
	for _, w := range []int{120, 140} {
		l := newLab(t, true, w, 34)
		l.keys("down") // web, which has a stack
		got := detailLabels(l.sh())
		for _, gone := range []string{"kind", "language", "stack"} {
			if got[gone] {
				t.Errorf("%d columns: the detail says %q, the list has it as a column", w, gone)
			}
		}
		for _, kept := range []string{"root", "affected"} {
			if !got[kept] {
				t.Errorf("%d columns: the detail lost %q, which the list does not show", w, kept)
			}
		}
		l.hasNot("language")
	}
}

func TestWhenTheListIsNotVisibleTheProjectDetailKeepsKindLanguageAndStack(t *testing.T) {
	// Narrow: the detail is opened over the list with Enter.
	l := newLab(t, true, 80, 30)
	l.keys("down")
	if got := detailLabels(l.sh()); !got["kind"] || !got["language"] || !got["stack"] || !got["root"] {
		t.Fatalf("80 columns: %v", got)
	}
	l.keys("enter")
	l.has("kind", "language", "stack", "root")
	// Wide, with the detail zoomed over the list.
	w := newLab(t, true, 120, 34)
	w.keys("down", "z")
	if !w.sh().detailFull() {
		t.Fatal("z did not zoom the detail")
	}
	w.has("kind", "language", "stack", "root")
}

func TestBesideTheListTheRunPanelLeavesOutWhatTheTaskRowShows(t *testing.T) {
	for _, w := range []int{120, 140} {
		l := newLab(t, true, w, 36)
		l.next = failingRun()
		l.keys("b")
		// The row of the failed task shows its state, time and reason, so the lines
		// under the rule say only the command; the log stays below them.
		if got := l.sh().taskInfo(w / 2); len(got) != 1 {
			t.Errorf("%d columns: the lines under the rule repeat the task row: %v", w, got)
		}
		l.has("$ go build ./...", "exited with code 2")
	}
}

func TestTheRunPanelKeepsTheStateTimeAndReasonOfTheSelectedTaskAndItsCommand(t *testing.T) {
	l := newLab(t, true, 80, 36)
	l.next = failingRun()
	l.keys("b")
	// The failed task is selected: its row is cut in a panel this narrow, so the
	// lines under the rule of the log say what the row had no room for.
	l.has("orders:build", "$ go build ./...", "exited with code 2")
}

func TestAPanelKeepsAnUnmarkedRowAndLeavesOutAMarkedOneBesideTheList(t *testing.T) {
	r := detail.Panel{ListVisible: true, Sections: []detail.Section{{Rows: []detail.Row{
		{Label: "reason", Value: "from cache", ShownByList: false}, {Label: "state", Value: "passed", ShownByList: true}}}}}
	lines, _ := r.View(theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII}), 40, "")
	if text := pane.Strip(strings.Join(lines, "\n")); !strings.Contains(text, "reason") || strings.Contains(text, "state") {
		t.Fatalf("%s", text)
	}
}

func TestTheDoctorDetailDoesNotRepeatTheStateTheHighlightedRowShows(t *testing.T) {
	for _, w := range []int{80, 120} {
		l := newLab(t, true, w, 34)
		l.keys("3")
		if w < layout.WideFrom {
			l.keys("enter")
		}
		lines, _ := l.sh().detailPanel().View(l.sh().th, 60, "")
		text := pane.Strip(strings.Join(lines, "\n"))
		for _, state := range []string{" failed", " warning", " passed"} {
			if strings.Contains(text, state) {
				t.Errorf("%d columns: the detail says %q:\n%s", w, state, text)
			}
		}
		for _, kept := range []string{"What is wrong", "How to fix", "Check again (r)"} {
			if !strings.Contains(text, kept) {
				t.Errorf("%d columns: the detail lost %q:\n%s", w, kept, text)
			}
		}
	}
}

func TestTheProjectDetailKeepsAFactWhenTheListIsTooNarrowToDrawItsColumn(t *testing.T) {
	var dropped, drawn bool
	// A project with a long name squeezes the columns after it: that is how the
	// table runs out of room now that it has no Tasks column to give way.
	for w, long := 100, false; w <= 150; w, long = w+5*btoi(long), !long {
		l := newLab(t, true, w, 34)
		if long {
			projects := append([]projectlist.Project(nil), sampleProjects...)
			projects[1].Name = "a-project-with-a-very-long-name-indeed"
			sh := l.sh()
			sh.list = projectlist.New(projects)
			l.m = sh
		}
		l.keys("down")
		got := detailLabels(l.sh())
		for _, c := range []struct{ label, header string }{{"kind", "Kind"}, {"language", "Language"}, {"stack", "Stack"}} {
			inList := strings.Contains(l.view(), c.header+" ")
			if got[c.label] == inList {
				t.Errorf("%d columns: list has %s column=%v, detail has %s row=%v", w, c.header, inList, c.label, got[c.label])
			}
			dropped, drawn = dropped || !inList, drawn || inList
		}
	}
	if !dropped || !drawn {
		t.Fatalf("the widths did not cover both cases (dropped %v, drawn %v)", dropped, drawn)
	}
}

func TestTheRunPanelSaysTheReasonUnderTheRuleOnlyWhenTheTaskRowIsCutBeforeIt(t *testing.T) {
	l := newLab(t, true, 120, 36)
	l.next = failingRun()
	l.keys("b")
	sh := l.sh()
	if got := sh.taskInfo(30); len(got) < 2 {
		t.Fatalf("a 30 cell panel should cut the reason and say it under the rule: %v", got)
	}
	if got := sh.taskInfo(100); len(got) != 1 {
		t.Fatalf("a 100 cell panel shows the reason in the row and says only the command: %v", got)
	}
}

func TestTheDoctorDetailFoldsBuildMetadataAndSaysEachThingOnce(t *testing.T) {
	l := newLab(t, true, 140, 34)
	sh := l.sh()
	const cause = "/go/bin/golangci-lint (golangci-lint has version 2.1.6"
	sh.tab = commands.TabDoctor
	sh.doctor.Report = &doctor.Report{Warnings: 1, Findings: []doctor.Finding{{Finding: protocol.Finding{
		Check: "golangci-lint matches CI", Warning: true,
		Detail: cause + ` built with go1.24.2 from (unknown, modified: ?, mod sum: "h1:abcDEF0123456789abcdef=") on windows/amd64), older than CI`,
		Remedy: "Install the version CI uses",
	}}}}
	l.m = sh
	panel := func(expanded bool) string {
		p := l.sh().detailPanel()
		p.Expanded = expanded
		lines, _ := p.View(l.sh().th, 80, "")
		return pane.Strip(strings.Join(lines, "\n"))
	}
	folded := panel(false)
	for _, hidden := range []string{"h1:", "modified: ?", "built with"} {
		if strings.Contains(folded, hidden) {
			t.Errorf("the build metadata %q shows before it is asked for:", hidden)
		}
	}
	if !strings.Contains(folded, "What is wrong") || !strings.Contains(folded, "Build details") || !strings.Contains(folded, "older than CI") {
		t.Error("the cause or the fold is missing")
	}
	if n := strings.Count(folded, "Install the version CI uses"); n != 1 {
		t.Errorf("the remedy shows %d times", n)
	}
	if open := panel(true); !strings.Contains(open, "modified: ?") {
		t.Error("e does not show the build metadata")
	}
}

// btoi is 1 for true and 0 for false.
func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func TestTaskButtonsSayWhatTheTaskDoesToTheKindOfProject(t *testing.T) {
	for _, tc := range []struct{ task, kind, want string }{
		{"build", "app", "Build app"}, {"build", "lib", "Build library"}, {"build", "package", "Build package"},
		{"test", "app", "Test app"}, {"test", "lib", "Test library"}, {"lint", "lib", "Lint library"},
		{"typecheck", "package", "Typecheck package"},
		{"package", "app", "Package app"}, {"package", "lib", "Package library"},
		// "package" is also a kind: the task of a package is told as what it does.
		{"package", "package", "Pack for publishing"},
		// A task with no verb, or a project of an unknown kind, keeps the raw name.
		{"stage-web", "app", "Run stage-web"}, {"package-dmg", "lib", "Run package-dmg"}, {"build", "tool", "Run build"},
	} {
		if got := taskLabel(tc.task, tc.kind); got != tc.want {
			t.Errorf("%s on a %s is %q, want %q", tc.task, tc.kind, got, tc.want)
		}
	}
}

func TestTheDetailNamesTheKindInItsTaskButtonsAndKeepsTheRawTaskInTheHint(t *testing.T) {
	l := buttonLab(t, 160, 100) // web, an app with package, slice-check and start tasks
	l.has("Build app (b)", "Test app (t)", "Package app (p)", "Run slice-check (s)", "Run start (S)")
	l.hasNot("Run build", "Run test", "Run package")
	l.has("smith run package", "smith run build")
	l.hasNot("smith run slice-check") // a task with no verb says nothing more than "Run <task>"
	actions := l.sh().detailPanel().Actions
	if actions[2].ID != "x:run:package" || actions[2].Shortcut != "p" {
		t.Fatalf("the Package app button changed its id or key: %+v", actions[2])
	}
	// The same task on a library.
	l.keys("down") // kit, a library
	l.has("Build library (b)", "Test library (t)")
	l.hasNot("Package app")
}
