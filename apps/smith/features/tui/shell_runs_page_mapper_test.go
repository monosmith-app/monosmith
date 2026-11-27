package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

// clickTask clicks the row of a task in the tree of the Runs view.
func (l *lab) clickTask(id string) *lab {
	l.t.Helper()
	sh := l.sh()
	in := sh.mainInner()
	w, h := sh.runRect()
	for row, hh := range sh.runsContent(w, h).hits {
		if hh.kind == "task" && len(hh.spans) > 0 && hh.spans[0].id == id {
			return l.send(click(in.X+3, in.Y+row)).send(letGo(in.X+3, in.Y+row))
		}
	}
	l.t.Fatalf("no row of %q in the tree:\n%s", id, l.view())
	return l
}

// framedRows checks that every row of the main frame is w cells and ends in
// its right border, so nothing of the view runs past the frame.
func framedRows(t *testing.T, l *lab, w int) {
	t.Helper()
	lines := l.lines()
	m := l.sh().lay().Main
	for i, line := range lines {
		if got := pane.Width(line); got > w {
			t.Fatalf("%d: line %d is %d cells: %q", w, i, got, line)
		}
	}
	for y := m.Y; y < m.Y+m.H && y < len(lines); y++ {
		row := []rune(lines[y])
		if end := m.X + m.W - 1; end >= len(row) || !strings.ContainsRune("│╮╯┤", row[end]) {
			t.Fatalf("%d: row %d does not end in the frame: %q", w, y, lines[y])
		}
	}
}

func TestTheRunsViewStaysInsideItsFrameAtEverySize(t *testing.T) {
	for _, size := range [][2]int{{80, 30}, {120, 40}, {150, 44}} {
		l := newLab(t, true, size[0], size[1])
		l.next = failingRun()
		l.keys("b")
		if l.sh().tab != commands.TabRuns {
			t.Fatalf("%v: the run opened in %q", size, l.sh().tab)
		}
		framedRows(t, l, size[0])
		if size[0] < 100 {
			l.has("‹ Runs / smith run build", "Tasks", "tab switches")
			l.keys("tab")
			framedRows(t, l, size[0])
			t.Logf("%dx%d log\n%s", size[0], size[1], l.view())
			l.keys("tab")
		} else {
			l.has("Run · smith run build", "orders", "└─", "main.go:12: undefined: Foo")
		}
		l.has("Failed 4 tasks", "× 1 failed", "✓ 1 passed", "≈ 1 cached", "4/4")
		t.Logf("%dx%d\n%s", size[0], size[1], l.view())
	}
}

func TestTheWideRunsViewHasTheTreeTheDividerAndTheLogSideBySide(t *testing.T) {
	l := newLab(t, true, 150, 44)
	l.next = failingRun()
	l.keys("b")
	_, y := l.find("main.go:12: undefined: Foo")
	row := l.lines()[y]
	in := l.sh().mainInner()
	g := l.sh().runsGeom(l.sh().runRect())
	if r := []rune(row); r[in.X+g.treeW] != '│' {
		t.Fatalf("the divider is not at column %d: %q", in.X+g.treeW, row)
	}
	// The rule meets the divider in a tee and the frame's bottom border too.
	rule := []rune(l.lines()[in.Y+g.ruleRow])
	bottom := []rune(l.lines()[in.Y+in.H])
	if rule[in.X+g.treeW] != '┬' || bottom[in.X+g.treeW] != '┴' {
		t.Fatalf("no tees: rule %q, bottom %q", string(rule), string(bottom))
	}
}

func TestANarrowRunsViewSwitchesBetweenTheTasksAndTheLogWithTab(t *testing.T) {
	l := newLab(t, true, 80, 30)
	l.next = failingRun()
	l.keys("b")
	l.has("All output", "orders")
	l.hasNot("main.go:12: undefined: Foo")
	l.keys("tab")
	if !l.sh().runOnLog {
		t.Fatal("Tab did not show the log")
	}
	l.has("main.go:12: undefined: Foo", "tasks")
	l.hasNot("└─")
	l.keys("tab")
	l.has("└─")
	// Enter on a task shows its log too.
	l.keys("enter")
	if !l.sh().runOnLog {
		t.Fatal("Enter on a task did not show its log")
	}
}

func TestAClickOnTheSwitchShowsTheLogOrTheTasks(t *testing.T) {
	l := newLab(t, true, 80, 30)
	l.next = failingRun()
	l.keys("b")
	l.clickText("Log: orders:build")
	if !l.sh().runOnLog {
		t.Fatalf("a click on Log did not show it:\n%s", l.view())
	}
	l.has("main.go:12: undefined: Foo")
	l.clickText("Tasks")
	if l.sh().runOnLog {
		t.Fatal("a click on Tasks did not show them")
	}
}

func TestErrorsOnlyKeepsTheFailedTasksInTheTree(t *testing.T) {
	l := newLab(t, true, 150, 44)
	l.next = failingRun()
	l.keys("b")
	l.keys("e")
	if got := l.sh().runIDs(); len(got) != 2 || got[1] != "orders:build" {
		t.Fatalf("errors only lists %v", got)
	}
	tree := strings.Join(l.inside(treeRect(l)), "\n")
	if strings.Contains(tree, "kit") || !strings.Contains(tree, "orders") {
		t.Fatalf("the tree keeps a task that did not fail:\n%s", tree)
	}
	l.has("e all tasks")
	l.keys("e")
	if got := l.sh().runIDs(); len(got) != 5 {
		t.Fatalf("all tasks again lists %v", got)
	}
}

// treeRect is the cells of the tree of a wide Runs view.
func treeRect(l *lab) layout.Rect {
	in := l.sh().mainInner()
	g := l.sh().runsGeom(l.sh().runRect())
	return layout.Rect{X: in.X, Y: in.Y + g.bodyTop, W: g.treeW, H: g.bodyRows}
}

func TestKeyThreeOpensTheRunsViewAndTheOtherViewsMovedUpOne(t *testing.T) {
	l := newHomeLab(t, 150, 44)
	l.keys("3")
	if l.sh().tab != commands.TabRuns {
		t.Fatalf("3 opened %q", l.sh().tab)
	}
	l.has("Nothing ran yet", "Runs", "Ctrl+K")
	for key, tab := range map[string]string{"4": commands.TabDoctor, "5": commands.TabDeps, "6": commands.TabRelease, "7": commands.TabSettings, "2": commands.TabProjects} {
		l.keys(key)
		if l.sh().tab != tab {
			t.Fatalf("%s opened %q, want %q", key, l.sh().tab, tab)
		}
	}
	x, y := l.railSpot(commands.TabRuns)
	l.send(click(x, y)).send(letGo(x, y))
	if l.sh().tab != commands.TabRuns {
		t.Fatalf("the rail row of Runs opened %q", l.sh().tab)
	}
}

func TestARunStartedFromProjectsOpensInRunsAndEscGoesBack(t *testing.T) {
	l := newLab(t, true, 150, 44)
	l.next = failingRun()
	l.keys("b")
	if l.sh().tab != commands.TabRuns || !l.sh().runShown() {
		t.Fatalf("the run opened in %q", l.sh().tab)
	}
	// The first Esc shows all output, the next goes back to Projects.
	l.keys("esc")
	if l.sh().runSel != "" || l.sh().tab != commands.TabRuns {
		t.Fatal("the first Esc did not show all output")
	}
	l.keys("esc")
	if l.sh().tab != commands.TabProjects {
		t.Fatalf("Esc went to %q", l.sh().tab)
	}
	l.contentHasNot("All output")
}

func TestAClickOnATaskRowShowsItsLog(t *testing.T) {
	l := newLab(t, true, 150, 44)
	l.next = failingRun()
	l.keys("b")
	l.clickTask("kit:build")
	if l.sh().runSel != "kit:build" {
		t.Fatalf("the click chose %q", l.sh().runSel)
	}
	l.has("ok example.com/kit 0.3s")
	l.clickTask("")
	if l.sh().runSel != "" {
		t.Fatal("a click on All output did not show it")
	}
}

func TestTheRunsOfTheSessionAreListedAndAClickOpensOne(t *testing.T) {
	l := newLab(t, true, 150, 44)
	l.next = failingRun()
	l.keys("b")
	l.keys("esc", "esc", "t")
	l.has("Runs this session", "smith run build", "smith run test")
	_, y := l.find("× smith run build")
	l.send(click(40, y)).send(letGo(40, y))
	if l.sh().runAt != 0 {
		t.Fatalf("the click opened run %d", l.sh().runAt)
	}
}

func TestTheButtonsOfTheRunsViewKeepTheirRowsWhateverTaskIsChosen(t *testing.T) {
	l := newLab(t, true, 140, 36)
	l.next = failingRun()
	l.keys("b")
	before := l.sh().runsGeom(l.sh().runRect()).headRows
	// web:build has no output: Copy log stays, faint, so no row moves.
	l.clickTask("web:build")
	if after := l.sh().runsGeom(l.sh().runRect()).headRows; after != before {
		t.Fatalf("the header went from %d to %d rows", before, after)
	}
	l.has("Copy log (Y)")
}

func TestTheLogOfTheRunsViewHasTheOneCellPaddingOfEveryPane(t *testing.T) {
	l := newLab(t, true, 140, 36)
	l.next = failingRun()
	l.keys("b")
	sh := l.sh()
	g := sh.runsGeom(sh.runRect())
	logged := 0
	for _, line := range sh.logColumn(g, g.logW)[g.logTop-g.bodyTop:] {
		if plain := pane.Strip(line); strings.TrimSpace(plain) != "" {
			logged++
			if !strings.HasPrefix(plain, " ") {
				t.Fatalf("a log line starts at the divider: %q", plain)
			}
		}
	}
	if logged == 0 {
		t.Fatal("the log is empty")
	}
}

func TestTheChosenTaskHasTheCursorMarkOfEveryPickListAndOnlyIt(t *testing.T) {
	l := newLab(t, true, 140, 36)
	l.next = failingRun()
	l.keys("b")
	l.clickTask("web:build")
	marked := 0
	for _, line := range l.inside(treeRect(l)) {
		if strings.Contains(line, "› ") {
			marked++
			if !strings.Contains(line, "build") || strings.Contains(line, "All output") {
				t.Fatalf("the mark is on another row: %q", line)
			}
		}
	}
	if marked != 1 || l.sh().runSel != "web:build" {
		t.Fatalf("%d rows carry the cursor mark, want 1 (chosen %q)", marked, l.sh().runSel)
	}
}
