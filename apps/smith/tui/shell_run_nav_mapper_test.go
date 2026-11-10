package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
)

func TestARunOpenKeepsTheScreenWhenTheTerminalGetsNarrowAndWideAgain(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.next = failingRun()
	l.keys("b")
	l.has("Back (Esc)", "Recent runs")
	l.send(tea.WindowSizeMsg{Width: 80, Height: 34})
	l.has("Back (Esc)")
	l.hasNot("Recent runs")
	l.send(tea.WindowSizeMsg{Width: 140, Height: 34})
	l.has("Back (Esc)", "Recent runs")
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

// screenRow is the screen row that holds text, or -1.
func screenRow(l *lab, text string) int {
	for y, line := range l.lines() {
		if strings.Contains(line, text) {
			return y
		}
	}
	return -1
}

func TestTheRecentRunsSitOnTheBottomRowsOfThePaneUnderARule(t *testing.T) {
	for _, h := range []int{20, 36} {
		l := projectsWithRuns(t, 140, h, 2)
		lines := l.lines()
		footer := h - 2 // the footer and the status bar are the last two rows
		if !strings.Contains(lines[footer-1], "smith run build a") || !strings.Contains(lines[footer-2], "smith run build b") {
			t.Fatalf("h=%d: the runs are not on the last rows of the pane:\n%s", h, l.view())
		}
		title := screenRow(l, "Recent runs")
		if title != footer-3 || !strings.HasPrefix(strings.TrimSpace(lines[title-1]), "───") {
			t.Fatalf("h=%d: the title is on row %d and the rule above it is %q:\n%s", h, title, lines[title-1], l.view())
		}
	}
}

func TestTheRecentRunsShowTheLastThreeAndSayHowManyMoreThereAre(t *testing.T) {
	l := projectsWithRuns(t, 140, 36, 7)
	l.has("smith run build g", "smith run build f", "smith run build e", "4 more runs")
	l.hasNot("smith run build d")
	l = projectsWithRuns(t, 140, 36, 4)
	l.has("1 more run,")
	l.hasNot("1 more runs", "smith run build a")
	l = projectsWithRuns(t, 140, 36, 3)
	l.has("smith run build a")
	l.hasNot("more run")
}

func TestTabOpensTheRecentRunsUpwardAndEscOrTabClosesThem(t *testing.T) {
	l := projectsWithRuns(t, 140, 36, 7)
	l.keys("tab")
	if !l.sh().recentFocus {
		t.Fatal("Tab did not give the keyboard to the recent runs")
	}
	l.has("smith run build a", "smith run build g", "Esc or Tab closes")
	l.hasNot("more runs")
	// The block grew upward: its last row is still on the bottom of the pane.
	if !strings.Contains(l.lines()[36-3], "smith run build a") {
		t.Fatalf("the block does not end on the bottom row:\n%s", l.view())
	}
	l.keys("esc")
	l.hasNot("smith run build a")
	l.has("4 more runs")
	if l.sh().recentFocus || l.sh().sideFocus {
		t.Fatalf("Esc left the keyboard on the recent runs (%v) or the detail (%v)", l.sh().recentFocus, l.sh().sideFocus)
	}
	l.keys("tab", "tab")
	l.hasNot("smith run build a")
	if l.sh().recentFocus || !l.sh().sideFocus {
		t.Fatalf("a second Tab moves on to the detail: recent %v, side %v", l.sh().recentFocus, l.sh().sideFocus)
	}
}

func TestTheOpenRecentRunsMoveWithTheArrowsAndEnterOpensTheChosenRun(t *testing.T) {
	l := projectsWithRuns(t, 140, 36, 7)
	l.keys("tab", "down", "down", "down", "down", "down", "down")
	if l.sh().recentAt != 6 {
		t.Fatalf("the cursor is on row %d, want the oldest (6)", l.sh().recentAt)
	}
	l.has("› smith run build a")
	l.keys("down") // the oldest run is the end of the list
	if l.sh().recentAt != 6 {
		t.Fatalf("Down went past the end: %d", l.sh().recentAt)
	}
	l.keys("enter")
	if !l.sh().runShown() || l.sh().runAt != 0 {
		t.Fatalf("Enter did not open the oldest run (run shown %v, at %d)", l.sh().runShown(), l.sh().runAt)
	}
}

func TestAClickOnTheClosedRecentRunsOpensThemAndADoubleClickOpensTheRun(t *testing.T) {
	l := projectsWithRuns(t, 140, 36, 7)
	x, y := l.find("smith run build f")
	l.send(click(x, y))
	if !l.sh().recentFocus || l.sh().recentAt != 1 {
		t.Fatalf("the click left recent focus %v at row %d", l.sh().recentFocus, l.sh().recentAt)
	}
	l.has("smith run build a", "Esc or Tab closes")
	if l.sh().runShown() {
		t.Fatal("a single click opened the run")
	}
	x, y = l.find("smith run build c")
	l.advance(time.Second)
	l.doubleClick(x, y)
	if !l.sh().runShown() || l.sh().runAt != 2 {
		t.Fatalf("the double click did not open the run (shown %v, at %d)", l.sh().runShown(), l.sh().runAt)
	}
}

func TestAShortProjectsPaneStillShowsTheProjectsAboveTheRecentRuns(t *testing.T) {
	for h := minHeight; h <= 24; h++ {
		l := projectsWithRuns(t, 140, h, 7)
		if got := len(l.lines()); got != h {
			t.Fatalf("h=%d: %d rows", h, got)
		}
		l.keys("tab")
		if got := len(l.lines()); got != h {
			t.Fatalf("h=%d, open: %d rows", h, got)
		}
		if h >= 18 && !strings.Contains(l.view(), "Project") {
			t.Fatalf("h=%d: the open runs took the whole list:\n%s", h, l.view())
		}
	}
}
