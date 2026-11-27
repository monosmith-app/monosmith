package splitter

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/layout"
)

func TestTheTreeDividerOfTheRunsViewMeetsItsRuleAndTheBottomBorderInTees(t *testing.T) {
	in := layout.Rect{X: 2, Y: 3, W: 60, H: 20}
	rule := RunsRuleSpot(in, 5)
	tree := RunsTreeSpot(in, 20, 6, 10, 30)
	js := Junctions([]Spot{rule, tree}, in.Y-1, in.Y+in.H)
	want := map[[2]int]string{{in.X + 20, in.Y + 5}: "┬", {in.X + 20, in.Y + in.H}: "┴"}
	if len(js) != len(want) {
		t.Fatalf("junctions %+v", js)
	}
	for _, j := range js {
		if want[[2]int{j.X, j.Y}] != j.Glyph {
			t.Errorf("junction %+v, want %v", j, want)
		}
	}
	if rule.Min != rule.Max {
		t.Fatalf("the rule moves: %+v", rule)
	}
}
