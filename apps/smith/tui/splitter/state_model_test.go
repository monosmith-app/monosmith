package splitter

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/muesli/termenv"
)

func init() { lipgloss.SetColorProfile(termenv.TrueColor) }

// detail is a column that sizes the panel after it: 40 wide on a 100-column screen.
func detail() Spot {
	return Spot{ID: "detail", Axis: Column, Hit: layout.Rect{X: 59, Y: 2, W: 1, H: 10}, Size: 40, Min: 28, Max: 64, Origin: 99, Sign: -1}
}

// log is a row that sizes the list before it: 5 rows, from screen row 5.
func logRow() Spot {
	return Spot{ID: "log", Axis: Row, Hit: layout.Rect{X: 0, Y: 8, W: 60, H: 1}, Size: 5, Min: 2, Max: 9, Origin: 3, Sign: 1}
}

func TestADragSetsTheSizeFromThePointerAndKeepsItInRange(t *testing.T) {
	s, ok := New(nil).Press([]Spot{detail()}, 59, 5)
	if !ok || s.Dragging != "detail" {
		t.Fatal("a press on the divider holds it")
	}
	s = s.Drag([]Spot{detail()}, 90, 5)
	if got, _ := s.Size("detail"); got != 28 {
		t.Fatalf("dragged right: the panel shrinks to its minimum, got %d", got)
	}
	s = s.Drag([]Spot{detail()}, 10, 5)
	if got, _ := s.Size("detail"); got != 64 {
		t.Fatalf("dragged far left: the panel stops at its maximum, got %d", got)
	}
	s = s.Drag([]Spot{detail()}, 49, 5)
	if got, _ := s.Size("detail"); got != 50 {
		t.Fatalf("the size follows the pointer: %d", got)
	}
}

func TestAPressOffTheDividerHoldsNothing(t *testing.T) {
	if s, ok := New(nil).Press([]Spot{detail()}, 30, 5); ok || s.Dragging != "" {
		t.Fatal("only the divider's own cells grab it")
	}
}

func TestReleaseSaysWhichDividerWasHeld(t *testing.T) {
	s, _ := New(nil).Press([]Spot{detail()}, 59, 5)
	s, id := s.Release()
	if id != "detail" || s.Dragging != "" {
		t.Fatalf("released %q", id)
	}
	if _, id := s.Release(); id != "" {
		t.Fatal("nothing was held")
	}
}

func TestTheKeysMoveTheSelectedDividerRightOrDown(t *testing.T) {
	spots := []Spot{detail(), logRow()}
	s, _ := New(nil).Select(spots, Row, 1)
	s, _, _ = s.Move(spots, 1, 2)
	if got, _ := s.Size("log"); got != 7 {
		t.Fatalf("down makes the list before it taller: %d", got)
	}
	s, _ = s.Select(spots, Column, 1)
	s, _, _ = s.Move(spots, 1, 4)
	if got, _ := s.Size("detail"); got != 36 {
		t.Fatalf("right makes the panel after it narrower: %d", got)
	}
	s, _, _ = s.Move(spots, -1, 100)
	if got, _ := s.Size("detail"); got != 64 {
		t.Fatalf("left is stopped at the maximum: %d", got)
	}
}

func TestWithNothingSelectedTheKeysMoveTheFirstDivider(t *testing.T) {
	s, p, ok := New(nil).Move([]Spot{detail()}, -1, 4)
	if !ok || p.ID != "detail" || s.Selected != "detail" {
		t.Fatalf("selected %q", s.Selected)
	}
	if _, _, ok := New(nil).Move(nil, 1, 4); ok {
		t.Fatal("no divider, nothing moves")
	}
}

func TestSelectWrapsAndKeepsToItsAxis(t *testing.T) {
	cols := []Spot{detail(), {ID: "other", Axis: Column, Hit: layout.Rect{X: 20, Y: 2, W: 1, H: 5}, Size: 10, Min: 5, Max: 20, Origin: 0, Sign: 1}, logRow()}
	s, _ := New(nil).Select(cols, Column, 1)
	s, _ = s.Select(cols, Column, 1)
	s, _ = s.Select(cols, Column, 1)
	if s.Selected != "detail" {
		t.Fatalf("wrapped to %q", s.Selected)
	}
	if _, ok := New(nil).Select([]Spot{detail()}, Row, 1); ok {
		t.Fatal("no row divider here")
	}
	if s, _ = s.Select(cols, Column, -1); s.Selected != "other" {
		t.Fatalf("backwards: %q", s.Selected)
	}
}

func TestCycleGoesBackwardsAndStartsAtTheLastDivider(t *testing.T) {
	spots := []Spot{detail(), logRow()}
	s, _ := New(nil).Cycle(spots, -1)
	if s.Selected != "log" {
		t.Fatalf("backwards from nothing: %q", s.Selected)
	}
	if s, _ = s.Cycle(spots, -1); s.Selected != "detail" {
		t.Fatalf("backwards: %q", s.Selected)
	}
	if s, _ = s.Cycle(spots, -1); s.Selected != "log" {
		t.Fatalf("wraps: %q", s.Selected)
	}
}

func TestCycleVisitsEveryDividerWithOneKey(t *testing.T) {
	spots := []Spot{detail(), logRow()}
	s, _ := New(nil).Cycle(spots, 1)
	s, _ = s.Cycle(spots, 1)
	if s.Selected != "log" {
		t.Fatalf("second divider: %q", s.Selected)
	}
	if s, _ = s.Cycle(spots, 1); s.Selected != "detail" {
		t.Fatalf("wraps: %q", s.Selected)
	}
	if _, ok := New(nil).Cycle(nil, 1); ok {
		t.Fatal("none to cycle")
	}
}

func TestStateIsAValueTheOldCopyKeepsItsSizes(t *testing.T) {
	a := New(map[string]int{"detail": 40})
	b, _, _ := a.Move([]Spot{detail()}, 1, 4)
	if got, _ := a.Size("detail"); got != 40 {
		t.Fatalf("the old state changed: %d", got)
	}
	if got, _ := b.Size("detail"); got != 36 {
		t.Fatalf("new state: %d", got)
	}
}

func TestLookSaysHeldBeforeSelected(t *testing.T) {
	s := State{Selected: "a", Dragging: "b"}
	if s.Look("b") != Dragging || s.Look("a") != Selected || s.Look("c") != Resting || s.Look("") != Resting {
		t.Fatal("looks")
	}
}

func TestPaintMarksOnlyTheDividersCells(t *testing.T) {
	th := theme.Default()
	lines := []string{strings.Repeat("a", 10), strings.Repeat("b", 10), strings.Repeat("c", 10)}
	p := Spot{ID: "x", Axis: Column, Hit: layout.Rect{X: 4, Y: 0, W: 1, H: 2}}
	if got := Paint(th, append([]string(nil), lines...), p, Resting); strings.Join(got, "") != strings.Join(lines, "") {
		t.Fatal("a resting divider is not touched")
	}
	got := Paint(th, append([]string(nil), lines...), p, Selected)
	if plain := lipglossStrip(got[0]); plain != "aaaa┃aaaaa" || lipglossStrip(got[2]) != lines[2] {
		t.Fatalf("column: %q", plain)
	}
	row := Spot{ID: "y", Axis: Row, Hit: layout.Rect{X: 2, Y: 1, W: 5, H: 1}}
	if plain := lipglossStrip(Paint(th, append([]string(nil), lines...), row, Dragging)[1]); plain != "bb━━━━━bbb" {
		t.Fatalf("row: %q", plain)
	}
}

// lipglossStrip removes the escape codes of a styled line.
func lipglossStrip(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
