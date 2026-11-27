package splitter

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/layout"
)

func TestSpotsAreBuiltFromPlainValues(t *testing.T) {
	l := layout.DivideWith(140, 30, true, 0)
	d := DetailSpot("doctor", l, 140)
	if d.ID != "doctor.detail" || d.Hit.X != l.Detail.X-1 || d.Size != l.Detail.W || d.Origin != 139 || d.Sign != -1 {
		t.Fatalf("detail spot %+v", d)
	}
}

func TestJunctionsJoinTheDivisionsToTheBordersOfThePane(t *testing.T) {
	in := layout.Rect{X: 1, Y: 3, W: 60, H: 20}
	list := RunsTreeSpot(in, 20, 0, 10, 30)
	detail := Spot{ID: SettingsDetailID, Axis: Row, Hit: layout.Rect{X: 21, Y: 7, W: in.X + in.W + layout.Pad - 21, H: 1}}
	got := Junctions([]Spot{list, detail}, 2, 23)
	want := []Junction{{X: 21, Y: 2, Glyph: "┬"}, {X: 21, Y: 23, Glyph: "┴"}, {X: 21, Y: 7, Glyph: "├"}}
	if len(got) != len(want) {
		t.Fatalf("junctions = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("junctions = %+v, want %+v", got, want)
		}
	}
}

func TestTheDividerBesideTheDetailIsAFrameBorderWithNoJunction(t *testing.T) {
	l := layout.DivideWith(120, 30, true, 0)
	div := DetailSpot("run", l, 120)
	run := RunsRuleSpot(l.DetailInner(), 6)
	if got := Junctions([]Spot{div, run}, l.Main.Y, l.Main.Y+l.Main.H-1); len(got) != 0 {
		t.Fatalf("run junctions = %+v", got)
	}
}
