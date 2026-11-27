package projectpage

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

func taskRows() []TaskRow {
	return []TaskRow{
		{Name: "build", Mark: Cached, Result: "passed", Time: "4.1s", Cache: "hit", Cached: true, Key: "b"},
		{Name: "test", Mark: Failed, Result: "failed", Time: "21.6s", Cache: "miss", Cached: true, Key: "t"},
		{Name: "package", Mark: NotRun, Result: "not run", Cached: true, Key: "p"},
		{Name: "start", Mark: NotRun, Result: "not run", Key: "s", OwnTerminal: true},
	}
}

func TestTheTasksTableHasAHeaderARuleAndARowForEachTask(t *testing.T) {
	lines, rowOf := TaskTable(plain, taskRows(), 76, 0, false, 0)
	head := pane.Strip(lines[0])
	for _, col := range []string{"Task", "Last result", "Time", "Cache", "Key"} {
		if !strings.Contains(head, col) {
			t.Fatalf("the header lacks %s: %q", col, head)
		}
	}
	for _, want := range []struct {
		line int
		text []string
	}{
		{2, []string{"build", "✓ passed", "4.1s", "hit", "b"}},
		{3, []string{"test", "× failed", "21.6s", "miss", "t"}},
		{4, []string{"package", "– not run", "–"}},
		{5, []string{"start", "own terminal", "never cached", "s"}},
	} {
		row := pane.Strip(lines[want.line])
		for _, s := range want.text {
			if !strings.Contains(row, s) {
				t.Fatalf("line %d lacks %q: %q", want.line, s, row)
			}
		}
		if rowOf[want.line] != want.line-2 {
			t.Fatalf("line %d answers a click with task %d", want.line, rowOf[want.line])
		}
	}
	if note := pane.Strip(strings.Join(lines[6:], " ")); !strings.Contains(note, "start runs in your terminal") {
		t.Fatalf("the table says which task needs its own terminal: %q", note)
	}
}

func TestTheRowAtTheCursorHasThePointerOnlyWhileTheDetailHasTheKeyboard(t *testing.T) {
	lines, _ := TaskTable(plain, taskRows(), 76, 1, true, 0)
	for i, l := range lines[2:6] {
		if strings.Contains(l, Pointer) != (i == 1) {
			t.Fatalf("row %d: pointer %v: %q", i, strings.Contains(l, Pointer), l)
		}
	}
	lines, _ = TaskTable(plain, taskRows(), 76, 1, false, 0)
	if strings.Contains(strings.Join(lines, ""), Pointer) {
		t.Fatal("no pointer while the list has the keyboard")
	}
}

func TestANarrowTableSaysResultAndDropsColumnsThatDoNotFit(t *testing.T) {
	for w := 30; w <= 90; w++ {
		lines, _ := TaskTable(plain, taskRows(), w, 0, true, 0)
		for _, l := range lines {
			if pane.Width(l) > w {
				t.Fatalf("w=%d: a line is %d cells: %q", w, pane.Width(l), pane.Strip(l))
			}
		}
		if head := pane.Strip(lines[0]); w < 64 && strings.Contains(head, "Last result") {
			t.Fatalf("w=%d: a narrow header says Result: %q", w, head)
		}
	}
}

func TestATableWithoutTasksSaysSo(t *testing.T) {
	lines, rowOf := TaskTable(plain, nil, 60, 0, true, 0)
	if len(lines) != 1 || len(rowOf) != 0 || !strings.Contains(lines[0], "no tasks") {
		t.Fatalf("%q", lines)
	}
}
