package pane

import "testing"

func TestCellEndsInAnEllipsisAndLeavesAGap(t *testing.T) {
	got := Cell("language, toolchain", 10)
	if Width(got) != 10 || Strip(got)[len(Strip(got))-1] != ' ' {
		t.Fatalf("cell must be 10 wide and end blank: %q", got)
	}
	if short := Cell("go", 6); short != "go    " {
		t.Fatalf("got %q", short)
	}
}
