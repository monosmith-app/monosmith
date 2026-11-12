package choicerow

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/muesli/termenv"
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

func TestALineKeepsTheBoxInTheSameColumnWhetherOrNotSelected(t *testing.T) {
	th := theme.Default()
	off := pane.Strip(Line(th, Ticks("name", false, false), 0))
	selected := Ticks("name", true, false)
	selected.Selected = true
	on := pane.Strip(Line(th, selected, 0))
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

func TestEveryLookOfARowDiffersAndAHoveredRowIsHoveredOnEveryLine(t *testing.T) {
	defer lipgloss.SetColorProfile(lipgloss.ColorProfile())
	lipgloss.SetColorProfile(termenv.TrueColor)
	th := theme.Default()
	r := Row{Lead: Box(false, false) + " ", Label: "a long label that wraps", Hint: "and its hint"}
	rest := Lines(th, r, 20)
	r.Hovered = true
	hover := Lines(th, r, 20)
	if len(hover) < 2 || len(hover) != len(rest) {
		t.Fatalf("the row wraps the same way hovered or not: %d vs %d lines", len(hover), len(rest))
	}
	for i := range hover {
		if hover[i] == rest[i] || !strings.HasPrefix(hover[i], th.RowHover("")[:4]) {
			t.Fatalf("line %d of a hovered row lacks the row hover look: %q", i, hover[i])
		}
		if pane.Width(hover[i]) != 20 {
			t.Fatalf("a hovered line fills the width: %d", pane.Width(hover[i]))
		}
	}
	r.Hovered, r.Selected = false, true
	if sel := Lines(th, r, 20); sel[0] == rest[0] || sel[0] == hover[0] {
		t.Fatal("the selected row has its own look")
	}
	r.Selected, r.Disabled = false, true
	if off := Lines(th, r, 20); off[0] == rest[0] {
		t.Fatal("a disabled row is muted")
	}
}

func TestAPressAnywhereOnAWrappedRowFindsIt(t *testing.T) {
	spans := []Span{{0, 1}, {2, 4}, {4, 5}}
	for line, want := range map[int]int{0: 0, 1: -1, 2: 1, 3: 1, 4: 2, 5: -1, -1: -1} {
		if got := At(spans, line); got != want {
			t.Errorf("At(%d) = %d, want %d", line, got, want)
		}
	}
}
