package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseview"
)

// pendingRelease gives the lab the plan of the Release tab: orders goes from 0.4.0 to 0.5.0.
func pendingRelease(l *lab) {
	sh := l.sh()
	projects := append([]projectlist.Project(nil), sampleProjects...)
	for i := range projects {
		if projects[i].Name == "orders" || projects[i].Name == "web" {
			projects[i].Version = "0.4.0"
		}
	}
	sh.list = projectlist.New(projects)
	sh.release = releaseview.FromResult(releaseview.Result{Plan: release.Plan{Items: []release.Item{{Project: "orders", Current: "0.4.0", Next: "0.5.0", Level: "minor"}}}})
	l.m = sh
}

func TestTheSelectedProjectHasTheCursorMarkOfEveryPickListAndOnlyIt(t *testing.T) {
	l := newLab(t, true, 140, 34).openTab("projects")
	sel := l.sh().list.SelectedName()
	marked := 0
	for _, line := range l.lines() {
		if i := strings.Index(line, "› "); i >= 0 && i < 4 {
			marked++
			if !strings.Contains(line, sel) {
				t.Fatalf("the mark is on another row: %q (selected %s)", line, sel)
			}
		}
	}
	if marked != 1 {
		t.Fatalf("%d rows carry the cursor mark, want 1", marked)
	}
	l.keys("down")
	next := l.sh().list.SelectedName()
	for _, line := range l.lines() {
		row := strings.TrimLeft(line, " ›")
		if (strings.HasPrefix(row, "○ "+next+" ") || strings.HasPrefix(row, "● "+next+" ")) && !strings.Contains(line, "› ") {
			t.Fatalf("the mark did not follow the selection to %s: %q", next, line)
		}
	}
}

func TestAPendingReleaseShowsAsCurrentArrowNextInTheListAndTheDetail(t *testing.T) {
	l := newLab(t, true, 140, 34)
	pendingRelease(l)
	l.has("Version", "0.4.0 → 0.5.0") // the list column, and the detail of the selected orders
	lines := strings.Join(l.lines(), "\n")
	if strings.Count(lines, "0.4.0 → 0.5.0") < 2 {
		t.Fatalf("the list and the detail both say it:\n%s", lines)
	}
	if strings.Contains(lines, "0.4.0 → 0.4.0") {
		t.Fatalf("web has no pending release:\n%s", lines)
	}
}

func TestTheWorkspaceActionsAreOneLineNextToNewProjectAndWrapWhenNarrow(t *testing.T) {
	l := newLab(t, true, 140, 34)
	var line string
	for _, s := range l.lines() {
		if strings.Contains(s, "New project") {
			line = s
		}
	}
	for _, want := range []string{"New project (a)", "Verify (V)", "Build (B)", "Test (T)", "Lint (L)", "Typecheck (C)"} {
		if !strings.Contains(line, want) {
			t.Fatalf("%q is not on the line of New project:\n%s", want, line)
		}
	}
	l = newLab(t, true, 100, 34)
	l.has("Verify (V)", "Typecheck (C)") // still all there: a button that does not fit starts a new line
}

func TestTheSummaryAndTheLegendShareALine(t *testing.T) {
	l := newLab(t, true, 140, 34)
	for _, s := range l.lines() {
		if strings.Contains(s, "affected by the changes") {
			if !strings.Contains(s, "= dependencies") {
				t.Fatalf("the legend is not on the line of the summary: %q", s)
			}
			return
		}
	}
	t.Fatal("no summary line")
}

func TestTheFixedTopOfTheProjectDetailStaysWhileTheActionsScroll(t *testing.T) {
	l := newLab(t, true, 140, 22)
	l.keys("down") // web: dependencies, a stack and many actions
	l.has("dependencies", "root")
	for range 12 {
		l.keys("tab", "down")
	}
	l.has("dependencies", "root") // the facts did not scroll away with the actions
	l.has("Graph")
}
