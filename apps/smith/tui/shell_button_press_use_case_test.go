package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// buttonLab is a Projects tab in colour, whose first project has tasks the keymap
// has no key for and derived files, so every kind of action row is on screen.
func buttonLab(t *testing.T, w, h int) *lab {
	t.Helper()
	l := newLab(t, true, w, h)
	projects := append([]projectlist.Project(nil), sampleProjects...)
	projects[0].Tasks = append(projects[0].Tasks, projectlist.Task{Name: "package", Command: "npm pack"},
		projectlist.Task{Name: "slice-check", Command: "smith slices"}, projectlist.Task{Name: "start", Command: "npm start"})
	projects[0].Derived = []projectlist.DerivedFile{{Path: "apps/web/jest.config.js", Status: projectlist.StatusStale}}
	sh := l.sh()
	sh.th = theme.Default()
	sh.list = projectlist.New(projects).SelectProject("web")
	l.m = sh
	return l
}

// runButton chooses the action labelled label in the split detail and returns
// the cell of the run button under the rule, the one button that runs it.
func (l *lab) runButton(label string) (x, y int) {
	l.t.Helper()
	l.clickText(label)
	return l.findLast(label)
}

// rawRow is the screen row y with its colour codes, to tell the states of a button apart.
func (l *lab) rawRow(y int) string { return strings.Split(l.sh().View(), "\n")[y] }

func TestEveryActionOfTheProjectDetailIsAButtonWithItsKey(t *testing.T) {
	l := buttonLab(t, 160, 100)
	actions := l.sh().detailPanel().Actions
	if len(actions) < 12 {
		t.Fatalf("only %d actions", len(actions))
	}
	seen := map[string]string{}
	for _, a := range actions {
		if a.Shortcut == "" {
			t.Errorf("%s (%s) has no key", a.Label, a.ID)
		}
		if other, dup := seen[a.Shortcut]; dup {
			t.Errorf("%s and %s share the key %q", a.ID, other, a.Shortcut)
		}
		seen[a.Shortcut] = a.ID
		l.has(a.Label + " (" + a.Shortcut + ")")
	}
}

func TestTheKeyOfAButtonDoesWhatTheButtonSays(t *testing.T) {
	for _, a := range buttonLab(t, 160, 100).sh().detailPanel().Actions {
		task, ok := strings.CutPrefix(a.ID, "x:run:")
		if !ok {
			continue
		}
		l := buttonLab(t, 160, 100) // a run in progress would refuse the next key
		l.keys(a.Shortcut)
		if l.lastStarted() != "run "+task+" -p web" {
			t.Errorf("the key %q of %s started %q", a.Shortcut, a.Label, l.lastStarted())
		}
	}
}

func TestAHoveredButtonLooksDifferentAndLeavingRestoresIt(t *testing.T) {
	l := buttonLab(t, 160, 100)
	x, y := l.find("Test app")
	rest := l.rawRow(y)
	l.send(motion(x+2, y))
	hover := l.rawRow(y)
	if hover == rest {
		t.Fatal("the button under the pointer must show the hover state")
	}
	// The margin cell at the left edge is on no button and no row.
	l.send(motion(0, y))
	if got := l.rawRow(y); got != rest {
		t.Fatalf("the button must be back at rest:\n%q\n%q", got, rest)
	}
}

func TestAPressedButtonShowsTheActiveStateAndRunsOnRelease(t *testing.T) {
	l := buttonLab(t, 160, 100)
	x, y := l.runButton("Test app")
	l.send(motion(x+2, y))
	hover := l.rawRow(y)
	l.send(click(x+2, y))
	active := l.rawRow(y)
	if active == hover || active == "" {
		t.Fatal("a held button shows its active state, not the hover state")
	}
	if len(l.started) != 0 {
		t.Fatal("nothing runs while the mouse button is held")
	}
	l.send(letGo(x+2, y))
	if l.lastStarted() != "run test -p web" {
		t.Fatalf("the release over the button runs it, started %q", l.lastStarted())
	}
}

func TestDraggingAwayFromAPressedButtonCancelsIt(t *testing.T) {
	l := buttonLab(t, 160, 100)
	x, y := l.runButton("Test app")
	l.send(click(x+2, y)).send(drag(x+2, y+4)).send(letGo(x+2, y+4))
	if len(l.started) != 0 {
		t.Fatalf("a release elsewhere cancels the press, started %q", l.lastStarted())
	}
	if l.sh().pressed != "" {
		t.Fatal("the press is over")
	}
	// The press is also over when the release is lost: the next motion without a button ends it.
	l.send(click(x+2, y)).send(motion(x+2, y+4))
	if l.sh().pressed != "" {
		t.Fatal("a motion with no button held ends the press")
	}
}

func TestAPressOnOneButtonReleasedOnAnotherRunsNothing(t *testing.T) {
	l := buttonLab(t, 160, 100)
	tx, ty := l.find("Test app")
	x, y := l.runButton("Build app")
	l.send(click(x+2, y)).send(drag(tx+2, ty)).send(letGo(tx+2, ty))
	if len(l.started) != 0 {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestTheKeyboardCursorOfTheOpenDetailKeepsItsMark(t *testing.T) {
	l := buttonLab(t, 90, 60)
	l.keys("enter") // the detail fills the pane
	if !l.sh().detailOpen {
		t.Fatal("the detail is open")
	}
	mark := strings.TrimRight(l.sh().th.Icons.Pointer, " ")
	markedRow := func() string {
		var rows []string
		for _, line := range l.lines() {
			if strings.Contains(line, mark+" ") && strings.Contains(line, "(") {
				rows = append(rows, line)
			}
		}
		if len(rows) != 1 {
			t.Fatalf("exactly one action has the cursor mark, found %d:\n%s", len(rows), l.view())
		}
		return rows[0]
	}
	if !strings.Contains(markedRow(), "Build app (b)") {
		t.Fatalf("the cursor starts on the first action:\n%s", l.view())
	}
	l.keys("down")
	if !strings.Contains(markedRow(), "Test app (t)") {
		t.Fatalf("the mark moves with the cursor:\n%s", l.view())
	}
}

func TestEveryActionOfTheRunAndDoctorDetailsShowsAKey(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.next = failingRun()
	l.keys("b")
	// The buttons of the run panel, Back and Stop first.
	if len(l.sh().runButtons()) < 2 {
		t.Fatalf("the run panel has buttons %v", l.sh().runButtons())
	}
	for _, a := range l.sh().runButtons() {
		if a.key == "" {
			t.Errorf("Run: %s has no key", a.label)
		}
	}
	l.keys("3")
	for _, a := range l.sh().detailPanel().Actions {
		if a.Shortcut == "" {
			t.Errorf("Doctor: %s has no key", a.Label)
		}
	}
}

func TestNoTabNamesTheFrameAtAnyWidth(t *testing.T) {
	for _, w := range []int{80, 120} {
		for _, tab := range []string{commands.TabProjects, commands.TabDoctor, commands.TabDeps, commands.TabRelease, commands.TabSettings} {
			l := newLab(t, true, w, 34)
			l.openTab(tab)
			top := l.lines()[l.sh().lay().Main.Y]
			if strings.Trim(top, "╭╮─┬ ") != "" {
				t.Errorf("%s at %d columns: the top edge of the pane has a title: %q", tab, w, top)
			}
		}
	}
}

func TestTheProjectsListHasNoOnLineAndItsRowIsGivenToTheList(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.hasNot("On: ")
	l.has("4 projects")
	in := l.sh().mainInner()
	if got := strings.TrimSpace(string([]rune(l.lines()[in.Y])[in.X:])); !strings.HasPrefix(got, "4 projects") {
		t.Fatalf("the count line is the first line of the pane, got %q", got)
	}
	if strings.Count(l.view(), "4 projects") != 1 {
		t.Fatal("the count is said once")
	}
}
