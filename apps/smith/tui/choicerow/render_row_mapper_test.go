package choicerow

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestBoxIsBlankTickedOrPartialAndPartialWins(t *testing.T) {
	for _, c := range []struct {
		checked, tri bool
		want         string
	}{{false, false, "[ ]"}, {true, false, "[x]"}, {false, true, "[-]"}, {true, true, "[-]"}} {
		if got := Box(c.checked, c.tri); got != c.want {
			t.Fatalf("Box(%v, %v) = %q, want %q", c.checked, c.tri, got, c.want)
		}
	}
}

func TestRadioMarksTheChosenOneAndOnlyThatOne(t *testing.T) {
	th := theme.Default()
	if got := pane.Strip(Radio(th, true)); got != "("+th.Icons.Dot+")" {
		t.Fatalf("on: %q", got)
	}
	if got := Radio(th, false); got != "( )" {
		t.Fatalf("off: %q", got)
	}
}

func TestRenderKeepsTheBoxInTheSameColumnWhetherOrNotSelected(t *testing.T) {
	th := theme.Default()
	off := Render(th, "name", false, false, false)
	on := Render(th, "name", true, true, false)
	if off != "  [ ] name" {
		t.Fatalf("off: %q", off)
	}
	if on != th.Icons.Pointer+" [x] name" {
		t.Fatalf("on: %q", on)
	}
	if pane.Width(off) != pane.Width(on) {
		t.Fatalf("width moved: %q vs %q", off, on)
	}
}
