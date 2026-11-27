package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/release"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/projectpage"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/releaseview"
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
	for _, line := range l.mainLines() {
		if i := strings.Index(line, projectpage.Pointer); i >= 0 && i < 4 {
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
	for _, line := range l.mainLines() {
		row := strings.TrimLeft(line, " ▶")
		if strings.HasPrefix(row, "–  "+next+" ") && !strings.Contains(line, projectpage.Pointer) {
			t.Fatalf("the mark did not follow the selection to %s: %q", next, line)
		}
	}
}

func TestAPendingReleaseShowsAsCurrentArrowNextInTheListAndTheDetail(t *testing.T) {
	// Beside the list the meta line of the detail of orders says it.
	l := newLab(t, true, 140, 34)
	pendingRelease(l)
	l.has("apps/orders · app · go · 0.4.0 → 0.5.0")
	// The list of a narrow screen has the room for a column of versions.
	n := newLab(t, true, 80, 30)
	pendingRelease(n)
	lines := strings.Join(n.lines(), "\n")
	if !strings.Contains(lines, "0.4.0 → 0.5.0") {
		t.Fatalf("the list says the pending release:\n%s", lines)
	}
	if strings.Contains(lines, "0.4.0 → 0.4.0") {
		t.Fatalf("web has no pending release:\n%s", lines)
	}
}

func TestTheListHasNoRowOfWorkspaceButtonsAndTheFooterKeepsTheirKeys(t *testing.T) {
	for _, w := range []int{130, 170} {
		l := newLab(t, true, w, 34)
		l.hasNot("New project (a)", "Verify (V)", "Build (B)", "Typecheck (C)")
		l.has("B build all")
	}
}

func TestTheLegendIsTheLastRowOfTheList(t *testing.T) {
	for _, size := range [][2]int{{140, 34}, {80, 30}} {
		l := newLab(t, true, size[0], size[1])
		lines := l.mainLines()
		if last := lines[len(lines)-1]; !strings.Contains(last, "✓ passed  × failed  – not run  ● changed") {
			t.Fatalf("%v: the last row of the list is %q", size, last)
		}
	}
}

func TestTheTopOfTheProjectDetailStaysWhileTheTasksScroll(t *testing.T) {
	l := manyTasks(t, 170, 18, 20)
	l.keys("tab")
	l.has("apps/orders · app", "Tasks    Dependencies    Runs    Settings")
	for range 19 {
		l.keys("down")
	}
	l.has("task19")
	l.has("apps/orders · app", "Tasks    Dependencies    Runs    Settings", "Run… (R)") // the top did not scroll away
}
