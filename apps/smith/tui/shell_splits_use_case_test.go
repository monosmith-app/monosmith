package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/splitter"
)

func drag(x, y int) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionMotion, Button: tea.MouseButtonLeft}
}

func letGo(x, y int) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionRelease, Button: tea.MouseButtonNone}
}

// keepingSplits makes the lab record every save of the panel sizes.
func keepingSplits(l *lab) *[]map[string]int {
	var saved []map[string]int
	sh := l.sh()
	sh.saveSplits = func(m map[string]int) { saved = append(saved, m) }
	l.m = sh
	return &saved
}

func TestDraggingTheDividerResizesTheDetailAndKeepsTheSize(t *testing.T) {
	l := doctorLab(t, 140, 30)
	saved := keepingSplits(l)
	lay := l.sh().lay()
	before, x, y := lay.Detail.W, lay.Detail.X-1, lay.Main.Y+4
	l.send(click(x, y)).send(drag(x-6, y)).send(drag(x-10, y)).send(letGo(x-10, y))
	if got := l.sh().lay().Detail.W; got != before+10 {
		t.Fatalf("dragging 10 columns left widens the detail by 10: %d -> %d", before, got)
	}
	if len(*saved) != 1 || (*saved)[0]["doctor.detail"] != before+10 {
		t.Fatalf("the size is kept once, on release: %v", *saved)
	}
	if l.sh().split.Dragging != "" {
		t.Fatal("the divider is let go")
	}
}

func TestAPressOnTheDividerSelectsNoRowBeneathIt(t *testing.T) {
	l := doctorLab(t, 140, 30)
	cursor := l.sh().doctor.Cursor
	lay := l.sh().lay()
	l.send(click(lay.Detail.X-1, lay.Main.Y+3)).send(letGo(lay.Detail.X-1, lay.Main.Y+3))
	if l.sh().doctor.Cursor != cursor {
		t.Fatal("the press belongs to the divider")
	}
}

func TestShiftArrowsSelectTheDividerAndBracketsMoveItUntilEscLetsGo(t *testing.T) {
	l := doctorLab(t, 140, 30)
	before := l.sh().lay().Detail.W
	l.hasNot("┃")
	l.send(tea.KeyMsg{Type: tea.KeyShiftRight})
	l.has("┃", "Divider selected")
	l.keys("]")
	if got := l.sh().lay().Detail.W; got != before-splitter.Step {
		t.Fatalf("] moves the selected divider right: %d -> %d", before, got)
	}
	l.keys("esc")
	l.hasNot("┃")
}

func TestTabPicksTheNextDividerWhileOneIsSelected(t *testing.T) {
	l := doctorLab(t, 140, 30)
	l.send(tea.KeyMsg{Type: tea.KeyShiftRight})
	if l.sh().split.Selected != "doctor.detail" {
		t.Fatalf("selected %q", l.sh().split.Selected)
	}
	l.keys("tab")
	if l.sh().split.Selected != "doctor.detail" {
		t.Fatal("with one divider tab stays on it and does not leave the screen")
	}
}

func TestTheBracketsMoveTheOnlyDividerWithNoSelection(t *testing.T) {
	l := doctorLab(t, 140, 30)
	before := l.sh().lay().Detail.W
	l.keys("[")
	if got := l.sh().lay().Detail.W; got != before+splitter.Step {
		t.Fatalf("[ widens the detail: %d -> %d", before, got)
	}
}

func TestNoDividerNoCapture(t *testing.T) {
	l := newLab(t, true, 140, 30)
	l.keys("1")
	l.send(tea.KeyMsg{Type: tea.KeyShiftDown})
	if l.sh().split.Selected != "" {
		t.Fatal("Home has no divider")
	}
}

func TestAnOpenDialogHidesTheDividers(t *testing.T) {
	l := doctorLab(t, 140, 30)
	l.keys("?")
	lay := l.sh().lay()
	l.send(click(lay.Detail.X-1, lay.Main.Y+3))
	if l.sh().split.Dragging != "" {
		t.Fatal("a divider under a dialog cannot be held")
	}
}

func TestDraggingTheRunDividerResizesTheTaskList(t *testing.T) {
	l := newLab(t, true, 140, 36)
	l.next = failingRun()
	l.keys("b")
	in := l.sh().mainInner()
	before := l.sh().runGeom(in.W, in.H).listRows
	sepY := in.Y + l.sh().runGeom(in.W, in.H).sepRow
	l.send(click(in.X+3, sepY)).send(drag(in.X+3, sepY-2)).send(letGo(in.X+3, sepY-2))
	if got := l.sh().runGeom(in.W, in.H).listRows; got != before-2 {
		t.Fatalf("dragging the divider up 2 rows takes 2 from the task list: %d -> %d", before, got)
	}
}

func TestTheSettingsDividersMoveWithTheKeys(t *testing.T) {
	l, _ := newSettingsLab(t, 140, 40)
	l.keys("7")
	in := l.sh().mainInner()
	before := l.sh().settings.Geometry(in.W, in.H)
	l.send(tea.KeyMsg{Type: tea.KeyShiftRight})
	l.keys("]")
	after := l.sh().settings.Geometry(in.W, in.H)
	if after.ListW != before.ListW+splitter.Step {
		t.Fatalf("the list of sections is %d wide, want %d", after.ListW, before.ListW+splitter.Step)
	}
	if after.SepRow < 0 {
		t.Skip("the section has no detail to resize")
	}
	l.send(tea.KeyMsg{Type: tea.KeyShiftDown})
	l.keys("]")
	moved := l.sh().settings.Geometry(in.W, in.H)
	if moved.DetailRows != after.DetailRows-splitter.Step && moved.DetailRows != moved.DetailMin {
		t.Fatalf("moving the row divider down shrinks the detail: %d -> %d", after.DetailRows, moved.DetailRows)
	}
}

func TestBackslashSelectsTheNextDividerAndTheBarThePrevious(t *testing.T) {
	l := newLab(t, true, 140, 36)
	l.next = failingRun()
	l.keys("b")
	l.keys(`\`)
	if l.sh().split.Selected != "run.detail" {
		t.Fatalf("first: %q", l.sh().split.Selected)
	}
	l.keys(`\`)
	if l.sh().split.Selected != "run.list" {
		t.Fatalf("second: %q", l.sh().split.Selected)
	}
	l.keys("|")
	if l.sh().split.Selected != "run.detail" {
		t.Fatalf("the bar goes back: %q", l.sh().split.Selected)
	}
}
