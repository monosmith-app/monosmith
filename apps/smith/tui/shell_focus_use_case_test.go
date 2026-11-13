package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
)

func TestTabMovesTheKeyboardBetweenTheListAndTheDetailOfProjects(t *testing.T) {
	l := newLab(t, true, 140, 34).openTab(commands.TabProjects)
	l.has("[detail (Tab)]")
	l.keys("tab")
	if !l.sh().sideFocus {
		t.Fatal("Tab did not move the keyboard to the detail")
	}
	l.has("[list (Tab)]", "↑↓ choose", "[go to run (Enter)]")
	l.keys("enter")
	l.has("[run (Enter)]")
	l.hasNot("go to run")
	l.keys("up")
	l.hasNot("[detail (Tab)]")
	l.keys("tab")
	if l.sh().sideFocus {
		t.Fatal("a second Tab did not give the keyboard back to the list")
	}
	l.has("[detail (Tab)]")
}

func TestOnAShortScreenTheRunButtonOfTheDetailStaysInView(t *testing.T) {
	for _, h := range []int{14, 18, 34} {
		l := newLab(t, true, 140, h).openTab(commands.TabProjects)
		l.keys("tab")
		r := l.sh().detailRect()
		c := l.sh().detailContent(r.W, r.H)
		shown := false
		for row, hit := range c.hits {
			if hit.id == detail.RunID && row < r.H {
				shown = true
			}
		}
		if !shown {
			t.Fatalf("%d rows: the run button is below the fold:\n%s", h, strings.Join(c.lines, "\n"))
		}
		// Moving down the actions never pushes it out.
		l.keys("down", "down", "down", "down", "down")
		c = l.sh().detailContent(r.W, r.H)
		shown = false
		for row, hit := range c.hits {
			if hit.id == detail.RunID && row < r.H {
				shown = true
			}
		}
		if !shown {
			t.Fatalf("%d rows, fifth action: the run button is below the fold", h)
		}
	}
}

func TestRightAndLeftMoveTheKeyboardAcrossTheListAndTheDetail(t *testing.T) {
	l := newLab(t, true, 140, 34).openTab(commands.TabProjects)
	l.keys("right")
	if !l.sh().sideFocus {
		t.Fatal("Right did not move the keyboard to the detail")
	}
	l.keys("left")
	if l.sh().sideFocus {
		t.Fatal("Left did not give the keyboard back to the list")
	}
	l.keys("right", "esc")
	if l.sh().sideFocus {
		t.Fatal("Esc did not give the keyboard back to the list")
	}
}

func TestInTheDetailTheArrowsChooseAnActionAndEnterRunsIt(t *testing.T) {
	l := newLab(t, true, 140, 34).openTab(commands.TabProjects)
	selected := l.sh().selectedProject()
	l.keys("tab", "down")
	if l.sh().selectedProject() != selected {
		t.Fatalf("Down in the detail moved the project to %q", l.sh().selectedProject())
	}
	if l.sh().detailCursor != 1 {
		t.Fatalf("the cursor is on action %d, want 1", l.sh().detailCursor)
	}
	l.has("› [Test app (t)]")
	l.keys("up", "up")
	if l.sh().detailCursor != 0 {
		t.Fatalf("Up went past the first action: %d", l.sh().detailCursor)
	}
	// Enter on an action of the split detail moves to the run button; Enter there runs it.
	l.keys("enter")
	if len(l.started) != 0 || !l.sh().detailOnRun {
		t.Fatalf("the first Enter must move to the run button, started %v", l.started)
	}
	l.keys("enter")
	if got := l.lastStarted(); got != "run build -p "+selected {
		t.Fatalf("Enter started %q", got)
	}
}

func TestInTheListTheArrowsStillMoveTheProjectAndTheDetailFollows(t *testing.T) {
	l := newLab(t, true, 140, 34).openTab(commands.TabProjects)
	first := l.sh().selectedProject()
	l.keys("down")
	if l.sh().selectedProject() == first {
		t.Fatal("Down did not move in the list")
	}
	l.has("Run on " + l.sh().selectedProject())
}

func TestTabGoesFromTheListToTheRecentRunsToTheDetailAndBack(t *testing.T) {
	l := projectsWithRuns(t, 140, 34, 5)
	step := func(recent, side bool, name string) {
		t.Helper()
		l.keys("tab")
		if l.sh().recentFocus != recent || l.sh().sideFocus != side {
			t.Fatalf("%s: recent %v, side %v", name, l.sh().recentFocus, l.sh().sideFocus)
		}
	}
	step(true, false, "from the list")
	step(false, true, "from the recent runs")
	step(false, false, "from the detail")
}

func TestOnANarrowScreenTabSwitchesBetweenTheListAndTheRecentRunsOnly(t *testing.T) {
	l := projectsWithRuns(t, 80, 30, 5)
	l.keys("tab")
	if !l.sh().recentFocus {
		t.Fatal("Tab did not reach the recent runs")
	}
	l.keys("tab")
	if l.sh().recentFocus || l.sh().sideFocus {
		t.Fatalf("back on the list: recent %v, side %v", l.sh().recentFocus, l.sh().sideFocus)
	}
	l = newLab(t, true, 80, 30).openTab(commands.TabProjects)
	l.keys("tab", "right")
	if l.sh().sideFocus || l.sh().recentFocus {
		t.Fatal("with nothing beside the list and no run, Tab and Right do nothing")
	}
}

func TestAClickPicksTheSideThatHasTheKeyboard(t *testing.T) {
	l := newLab(t, true, 140, 34).openTab(commands.TabProjects)
	x, y := l.find("Run on")
	l.send(click(x, y))
	if !l.sh().sideFocus {
		t.Fatal("a click in the detail did not give it the keyboard")
	}
	x, y = l.find("web")
	l.send(click(x, y))
	if l.sh().sideFocus {
		t.Fatal("a click in the list did not take the keyboard back")
	}
}

func TestAnOpenRunHasTheKeyboardUntilTabGivesItToTheList(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.next = failingRun()
	l.keys("b")
	if !l.runPanelHasKeys() {
		t.Fatal("a run that starts does not take the keyboard")
	}
	l.has("[list (Tab)]")
	l.keys("tab")
	if l.runPanelHasKeys() || !l.sh().runShown() {
		t.Fatal("Tab should keep the run open and give the keyboard to the list")
	}
	selected := l.sh().selectedProject()
	l.keys("down")
	if l.sh().selectedProject() == selected {
		t.Fatal("Down did not move the project while the run stays open")
	}
	l.keys("tab") // the recent runs
	l.keys("tab") // the run
	if !l.runPanelHasKeys() {
		t.Fatal("Tab did not come back to the run")
	}
}

// runPanelHasKeys is true while the keys go to the run that is open.
func (l *lab) runPanelHasKeys() bool { return l.sh().runShown() && l.sh().sideKeys() }

func TestOnDoctorTabAndRightMoveTheKeyboardToTheDetail(t *testing.T) {
	l := newLab(t, true, 140, 34).openTab(commands.TabDoctor)
	l.has("[detail (Tab)]")
	l.keys("tab")
	if !l.sh().sideFocus {
		t.Fatal("Tab did not move the keyboard to the detail of the check")
	}
	l.has("[list (Tab)]", "↑↓ choose")
	cursor := l.sh().doctor.Cursor
	l.keys("down")
	if l.sh().doctor.Cursor != cursor {
		t.Fatal("Down in the detail moved the check")
	}
	l.keys("left")
	if l.sh().sideFocus {
		t.Fatal("Left did not give the keyboard back to the list")
	}
	l.keys("down")
	if l.sh().doctor.Cursor == cursor {
		t.Fatal("Down in the list did not move the check")
	}
	// Doctor groups the checks by project: Down until the cursor is on a check, not a header.
	for i := 0; i < 5 && l.sh().doctor.OnHeader(); i++ {
		l.keys("down")
	}
	l.keys("right")
	if !l.sh().sideFocus {
		t.Fatal("Right on a check did not move the keyboard to its detail")
	}
	l.keys("esc")
	if l.sh().sideFocus {
		t.Fatal("Esc did not give the keyboard back to the list")
	}
}

func TestOnDoctorRightOnAGroupHeaderFoldsItAsBefore(t *testing.T) {
	l := newLab(t, true, 140, 34).openTab(commands.TabDoctor)
	// The checks are grouped by project: Down until the cursor is on a header.
	for i := 0; i < 5 && !l.sh().doctor.OnHeader(); i++ {
		l.keys("down")
	}
	if !l.sh().doctor.OnHeader() {
		t.Fatalf("the cursor is not on a group header (row %d)", l.sh().doctor.Cursor)
	}
	l.keys("left", "right")
	if l.sh().sideFocus {
		t.Fatal("Right on a header moved the keyboard instead of unfolding the group")
	}
}

func TestOtherViewsHaveNoSidePanelToTabTo(t *testing.T) {
	for _, tab := range []string{commands.TabDeps, commands.TabRelease} {
		l := newLab(t, true, 140, 34).openTab(tab)
		l.keys("tab", "right", "left")
		if l.sh().sideFocus {
			t.Fatalf("%s: the keyboard moved to a side panel", tab)
		}
	}
}

// manyTasks is the Projects tab whose first project has n tasks, so its detail is far taller than a short pane.
func manyTasks(t *testing.T, w, h, n int) *lab {
	t.Helper()
	l := newLab(t, true, w, h)
	projects := append([]projectlist.Project(nil), sampleProjects...)
	projects[1].Tasks = nil
	for i := 0; i < n; i++ {
		projects[1].Tasks = append(projects[1].Tasks, projectlist.Task{Name: fmt.Sprintf("task%02d", i), Command: "echo " + fmt.Sprint(i)})
	}
	sh := l.sh()
	sh.list = projectlist.New(projects).SelectProject("orders")
	l.m = sh
	return l.openTab(commands.TabProjects)
}

func TestTheDetailScrollsToFollowTheCursorDownAndUp(t *testing.T) {
	l := manyTasks(t, 140, 20, 12)
	l.keys("tab")
	l.hasNot("task11")
	seen := false
	for i := 0; i < 14; i++ {
		l.keys("down")
		seen = seen || strings.Contains(l.view(), "task11")
		a := l.sh().detailActions()[l.sh().detailCursor]
		if !strings.Contains(l.view(), "› ["+a.Label) {
			t.Fatalf("step %d: the cursor is out of view (cursor %d, top %d):\n%s", i+1, l.sh().detailCursor, l.sh().detailTop, l.view())
		}
		// The cursor mark is on screen, on the action the cursor chose.
		if !strings.Contains(l.view(), "› ["+a.Label) && !strings.Contains(l.view(), "›  ["+a.Label) && !strings.Contains(l.view(), a.Label) {
			t.Fatalf("step %d: %q is not on screen:\n%s", i+1, a.Label, l.view())
		}
	}
	if !seen {
		t.Fatal("task11, below the fold at the start, never came into view")
	}
	if l.sh().detailTop == 0 {
		t.Fatal("the detail never scrolled")
	}
	for i := 0; i < 20; i++ {
		l.keys("up")
	}
	l.has("Run on orders")
	l.hasNot("↑ more")
}

func TestTheDetailShowsTheDescriptionOfTheActionTheCursorReaches(t *testing.T) {
	l := manyTasks(t, 140, 20, 12)
	l.keys("tab", "down", "down", "down", "down", "down", "down")
	a := l.sh().detailActions()[l.sh().detailCursor]
	r := l.sh().detailRect()
	panel := l.sh().detailPanel()
	panel.Cursor = a.ID
	_, rows := panel.View(l.sh().th, r.W, "")
	for row, id := range rows {
		if id != a.ID {
			continue
		}
		if row < l.sh().detailTop || row >= l.sh().detailTop+r.H {
			t.Fatalf("row %d of %q is outside the window at %d of %d rows", row, a.Label, l.sh().detailTop, r.H)
		}
	}
}

func TestDoctorDetailScrollsToFollowTheCursorToo(t *testing.T) {
	l := newLab(t, true, 140, 12).openTab(commands.TabDoctor)
	l.keys("tab")
	actions := len(l.sh().detailActions())
	for i := 0; i < actions+2; i++ {
		l.keys("down")
	}
	a := l.sh().detailActions()[l.sh().detailCursor]
	r := l.sh().detailRect()
	panel := l.sh().detailPanel()
	panel.Cursor = a.ID
	_, rows := panel.View(l.sh().th, r.W, "")
	for row, id := range rows {
		if id == a.ID && (row < l.sh().detailTop || row >= l.sh().detailTop+r.H) {
			t.Fatalf("the action %q (row %d) is out of the window at %d:\n%s", a.Label, row, l.sh().detailTop, l.view())
		}
	}
}
