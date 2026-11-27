package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/projectpage"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/runview"
)

func TestARunOpenKeepsTheScreenWhenTheTerminalGetsNarrowAndWideAgain(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.next = failingRun()
	l.keys("b")
	l.has("Back (Esc)", "Filter:") // the run beside the list
	l.send(tea.WindowSizeMsg{Width: 80, Height: 34})
	l.has("Back (Esc)")
	l.hasNot("Filter:")
	l.send(tea.WindowSizeMsg{Width: 140, Height: 34})
	l.has("Back (Esc)", "Filter:")
	if l.sh().detailOpen {
		t.Fatal("beside the list the run is not the detail that fills the pane")
	}
}

// projectsWithRuns is the Projects tab with n finished runs kept, the newest named build-<letter>.
func projectsWithRuns(t *testing.T, w, h, n int) *lab {
	t.Helper()
	l := newLab(t, true, w, h)
	for i := 0; i < n; i++ {
		finishedRun(l, "run", "build", string(rune('a'+i)))
	}
	return l.openTab(commands.TabProjects)
}

// taskRun puts a finished run of one task of a project in the history of the
// lab: the line is what smith printed when the task ended ("✓ 106ms").
func taskRun(l *lab, project, task, end string) {
	sh := l.sh()
	id := project + ":" + task
	r := runview.NewRun([]string{"run", task, "-p", project}, []string{id}, l.clock)
	r.Feed(id+" ▶ "+task, l.clock)
	r.Feed(id+" "+end, l.clock.Add(2*time.Second))
	r.End(0, false, l.clock.Add(2*time.Second))
	sh.runs = append(sh.runs, r)
	l.m = sh
}

// runsSection is the Projects tab with the Runs section of orders open and
// its keyboard in the detail: three runs of orders, the newest a failed test.
func runsSection(t *testing.T, w, h int) *lab {
	t.Helper()
	l := newLab(t, true, w, h)
	taskRun(l, "orders", "build", "✓ 106ms")
	taskRun(l, "kit", "build", "✓ 90ms")
	taskRun(l, "orders", "build", "✓ from cache")
	taskRun(l, "orders", "test", "× exited with code 1")
	l.openTab(commands.TabProjects)
	sh := l.sh()
	sh.list = sh.list.SelectProject("orders")
	l.m = sh
	return l.clickText("Runs")
}

// screenRow is the screen row that holds text, or -1.
func screenRow(l *lab, text string) int {
	for y, line := range l.lines() {
		if strings.Contains(line, text) {
			return y
		}
	}
	return -1
}

func TestTheRunsSectionListsTheRunsOfTheProjectNewestFirst(t *testing.T) {
	l := runsSection(t, 150, 44)
	if l.sh().projSection != projectpage.Runs {
		t.Fatalf("the click on the tab opened section %d", l.sh().projSection)
	}
	test, cached, build := screenRow(l, "× test"), screenRow(l, "✓ build"), -1
	for y, line := range l.lines() {
		if strings.Contains(line, "✓ build") && y > cached {
			build = y
		}
	}
	if test < 0 || cached < 0 || build < 0 || test >= cached || cached >= build {
		t.Fatalf("the runs of orders are not newest first (test %d, cached %d, build %d):\n%s", test, cached, build, l.view())
	}
	if strings.Contains(l.view(), "kit:build") {
		t.Fatal("a run of another project is listed")
	}
}

func TestTheRunsSectionSaysWhenTheProjectHasNotRun(t *testing.T) {
	l := projectsWithRuns(t, 150, 44, 3)
	l.clickText("Runs")
	l.has("No run of this project in this session yet.")
}

func TestTabFromTheListMovesTheKeyboardToTheDetailAndEscBringsItBack(t *testing.T) {
	l := projectsWithRuns(t, 170, 39, 7)
	l.keys("tab")
	if !l.sh().sideFocus {
		t.Fatal("Tab did not give the keyboard to the detail")
	}
	l.keys("esc")
	if l.sh().sideFocus {
		t.Fatal("Esc left the keyboard on the detail")
	}
}

func TestInTheRunsSectionTheArrowsChooseARunAndEnterOpensIt(t *testing.T) {
	l := runsSection(t, 150, 44)
	// The click on the tab gave the keyboard to the detail.
	if l.sh().projSection != projectpage.Runs || !l.sh().sideFocus {
		t.Fatalf("the keyboard is on section %d (side %v)", l.sh().projSection, l.sh().sideFocus)
	}
	l.keys("down", "down", "down")
	if l.sh().projRunAt != 2 {
		t.Fatalf("the cursor is on run %d, want the oldest (2)", l.sh().projRunAt)
	}
	l.keys("enter")
	if !l.sh().runShown() || l.sh().runAt != 0 {
		t.Fatalf("Enter did not open the oldest run of orders (run shown %v, at %d)", l.sh().runShown(), l.sh().runAt)
	}
}

func TestAClickOnARunOfTheRunsSectionChoosesItAndADoubleClickOpensIt(t *testing.T) {
	l := runsSection(t, 150, 44)
	x, y := l.find("✓ build")
	l.send(click(x, y))
	if l.sh().projRunAt != 1 || l.sh().runShown() {
		t.Fatalf("the click chose run %d (shown %v), want 1 and no run open", l.sh().projRunAt, l.sh().runShown())
	}
	l.advance(time.Second)
	l.doubleClick(x, y)
	if !l.sh().runShown() || l.sh().runAt != 2 {
		t.Fatalf("the double click did not open the cached build (shown %v, at %d)", l.sh().runShown(), l.sh().runAt)
	}
}

func TestAShortProjectsPaneStillFitsTheScreen(t *testing.T) {
	for h := minHeight; h <= 24; h++ {
		l := projectsWithRuns(t, 140, h, 7)
		if got := len(l.lines()); got != h {
			t.Fatalf("h=%d: %d rows", h, got)
		}
		l.keys("tab", "tab")
		if got := len(l.lines()); got != h {
			t.Fatalf("h=%d, in the detail: %d rows", h, got)
		}
		if h >= 18 && !strings.Contains(l.view(), "orders") {
			t.Fatalf("h=%d: the list lost its projects:\n%s", h, l.view())
		}
	}
}
