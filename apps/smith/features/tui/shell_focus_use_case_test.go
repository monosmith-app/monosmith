package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/projectpage"
)

func TestTabMovesTheKeyboardBetweenTheListAndTheDetailOfProjects(t *testing.T) {
	l := newLab(t, true, 140, 34).openTab(commands.TabProjects)
	l.has("Tab detail", "/ filter")
	l.keys("tab")
	if !l.sh().sideFocus {
		t.Fatal("Tab did not move the keyboard to the detail")
	}
	l.has("Tab section", "Enter run", "Esc back")
	l.hasNot("Tab detail", "/ filter")
	l.keys("tab")
	if !l.sh().sideFocus || l.sh().projSection != projectpage.Dependencies {
		t.Fatalf("a second Tab must show the next section: side %v, section %d", l.sh().sideFocus, l.sh().projSection)
	}
	l.keys("esc")
	if l.sh().sideFocus {
		t.Fatal("Esc did not give the keyboard back to the list")
	}
	l.has("Tab detail")
}

func TestOnAShortScreenTheRunButtonOfTheDetailStaysInView(t *testing.T) {
	for _, h := range []int{14, 18, 34} {
		l := newLab(t, true, 140, h).openTab(commands.TabProjects)
		l.keys("tab", "R") // the Run… menu
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
	// In the detail the side arrows change the section; Esc goes back to the list.
	l.keys("left")
	if !l.sh().sideFocus || l.sh().projSection != projectpage.Settings {
		t.Fatalf("Left in the detail must show the previous section: side %v, section %d", l.sh().sideFocus, l.sh().projSection)
	}
	l.keys("right")
	if l.sh().projSection != projectpage.Tasks {
		t.Fatalf("Right must show the next section, got %d", l.sh().projSection)
	}
	l.keys("esc")
	if l.sh().sideFocus {
		t.Fatal("Esc did not give the keyboard back to the list")
	}
}

func TestInTheRunMenuTheArrowsChooseAnActionAndEnterRunsIt(t *testing.T) {
	l := newLab(t, true, 140, 34).openTab(commands.TabProjects)
	selected := l.sh().selectedProject()
	l.keys("R", "down")
	if !l.sh().projMenu || !l.sh().sideFocus {
		t.Fatalf("R did not open the Run… menu with the keyboard in it: menu %v, side %v", l.sh().projMenu, l.sh().sideFocus)
	}
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

func TestEscClosesTheRunMenuAndTheSectionsComeBack(t *testing.T) {
	l := newLab(t, true, 140, 34).openTab(commands.TabProjects)
	l.keys("R")
	l.has("Run on " + l.sh().selectedProject())
	l.keys("esc")
	if l.sh().projMenu || !l.sh().sideFocus {
		t.Fatalf("Esc must close the menu and keep the keyboard in the detail: menu %v, side %v", l.sh().projMenu, l.sh().sideFocus)
	}
	l.has("Last result")
	l.hasNot("Run on ")
}

func TestInTheTasksSectionTheArrowsChooseATaskAndEnterRunsIt(t *testing.T) {
	l := newLab(t, true, 140, 34).openTab(commands.TabProjects)
	selected := l.sh().selectedProject()
	l.keys("tab", "down", "down")
	if l.sh().taskAt != 1 {
		t.Fatalf("the cursor is on task %d, want the last (1)", l.sh().taskAt)
	}
	if l.sh().selectedProject() != selected {
		t.Fatal("Down in the detail moved the project")
	}
	l.keys("enter")
	if got := l.lastStarted(); got != "run test -p "+selected {
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
	p, _ := l.sh().list.Selected()
	l.has(p.Root + " · " + p.Kind)
}

func TestTabGoesFromTheListToTheDetailWhereItShowsTheNextSection(t *testing.T) {
	l := projectsWithRuns(t, 140, 34, 5)
	l.keys("tab")
	if !l.sh().sideFocus || l.sh().projSection != projectpage.Tasks {
		t.Fatalf("from the list: side %v, section %d", l.sh().sideFocus, l.sh().projSection)
	}
	for _, want := range []projectpage.Section{projectpage.Dependencies, projectpage.Runs, projectpage.Settings, projectpage.Tasks} {
		l.keys("tab")
		if !l.sh().sideFocus || l.sh().projSection != want {
			t.Fatalf("Tab showed section %d, want %d", l.sh().projSection, want)
		}
	}
	l.keys("shift+tab")
	if l.sh().projSection != projectpage.Settings {
		t.Fatalf("Shift+Tab showed section %d, want Settings", l.sh().projSection)
	}
}

func TestOnANarrowScreenTabAndRightDoNothingInTheList(t *testing.T) {
	for _, l := range []*lab{projectsWithRuns(t, 80, 30, 5), newLab(t, true, 80, 30).openTab(commands.TabProjects)} {
		l.keys("tab", "right")
		if l.sh().sideFocus || l.sh().detailOpen {
			t.Fatal("with nothing beside the list, Tab and Right do nothing")
		}
	}
}

func TestAClickPicksTheSideThatHasTheKeyboard(t *testing.T) {
	l := newLab(t, true, 140, 34).openTab(commands.TabProjects)
	x, y := l.find("Last result")
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
	l.has("Tab list")
	l.keys("tab")
	if l.runPanelHasKeys() || !l.sh().runShown() {
		t.Fatal("Tab should keep the run open and give the keyboard to the list")
	}
	selected := l.sh().selectedProject()
	l.keys("down")
	if l.sh().selectedProject() == selected {
		t.Fatal("Down did not move the project while the run stays open")
	}
	l.keys("tab") // the run
	if !l.runPanelHasKeys() {
		t.Fatal("Tab did not come back to the run")
	}
}

// runPanelHasKeys is true while the keys go to the run that is open.
func (l *lab) runPanelHasKeys() bool { return l.sh().runShown() && l.sh().sideKeys() }

func TestOnDoctorTabAndRightMoveTheKeyboardToTheDetail(t *testing.T) {
	l := newLab(t, true, 140, 34).openTab(commands.TabDoctor)
	l.has("Tab detail")
	l.keys("tab")
	if !l.sh().sideFocus {
		t.Fatal("Tab did not move the keyboard to the detail of the check")
	}
	l.has("Tab list", "↑↓ choose")
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
	l.keys("R") // the Run… menu has a button for each task
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
	l.keys("R", "down", "down", "down", "down", "down", "down")
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

func TestTheTasksTableScrollsToFollowTheCursor(t *testing.T) {
	l := manyTasks(t, 140, 16, 20)
	l.keys("tab")
	l.hasNot("task19")
	for range 19 {
		l.keys("down")
		if !strings.Contains(l.view(), fmt.Sprintf("task%02d", l.sh().taskAt)) {
			t.Fatalf("task%02d, at the cursor, is out of view:\n%s", l.sh().taskAt, l.view())
		}
	}
	l.has("task19")
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
