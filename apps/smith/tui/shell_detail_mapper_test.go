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
	if strings.Contains(hint, "^") || !strings.Contains(hint, "runs after the build of the projects it needs and its own build") {
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

func TestBesideTheListTheRunDetailLeavesOutWhatTheTaskRowShows(t *testing.T) {
	for _, w := range []int{120, 140} {
		l := newLab(t, true, w, 36)
		l.next = failingRun()
		l.keys("b")
		got := detailLabels(l.sh())
		for _, gone := range []string{"state", "time", "reason"} {
			if got[gone] {
				t.Errorf("%d columns: the detail says %q, the task row has it", w, gone)
			}
		}
		for _, kept := range []string{"command", "log"} {
			if !got[kept] {
				t.Errorf("%d columns: the detail lost %q", w, kept)
			}
		}
	}
}

func TestWhenTheListIsNotVisibleTheRunDetailKeepsStateTimeAndReason(t *testing.T) {
	l := newLab(t, true, 80, 36)
	l.next = failingRun()
	l.keys("b")
	if got := detailLabels(l.sh()); !got["state"] || !got["reason"] || !got["command"] {
		t.Fatalf("80 columns: %v", got)
	}
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
		l.keys("4")
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
	for w := 100; w <= 150; w += 5 {
		l := newLab(t, true, w, 34)
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

func TestTheRunDetailKeepsTheReasonWhenTheTaskRowIsCutBeforeIt(t *testing.T) {
	l := newLab(t, true, 120, 36)
	l.next = failingRun()
	l.keys("b")
	sh := l.sh()
	tk := sh.run.Task(sh.runSel)
	if _, _, shown := sh.taskRowShows(tk, 30, sh.now()); shown {
		t.Fatal("a 30 cell list should cut the reason")
	}
	if _, _, shown := sh.taskRowShows(tk, 100, sh.now()); !shown {
		t.Fatal("a 100 cell list should show the reason")
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
