package splitter

import (
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/layout"
)

// The ids of the dividers a screen has besides the one beside the detail, which
// has one per tab (DetailID) so each view keeps its own width.
const (
	SettingsListID   = "settings.sections"
	SettingsDetailID = "settings.detail"
	// RunsTreeID is the divider between the tree and the log of the Runs view;
	// RunsRuleID the rule under its header, which does not move.
	RunsTreeID        = "runs.tree"
	RunsRuleID        = "runs.rule"
	detailDividerTail = ".detail"
)

// DetailID is the id of the divider beside the detail of the tab.
func DetailID(tab string) string { return tab + detailDividerTail }

// DetailSpot is the divider between the list and the detail of the tab, for the
// layout l of a screen w cells wide: the right border of the list's frame,
// between its top and bottom corners.
func DetailSpot(tab string, l layout.Layout, w int) Spot {
	in := l.Main.Inner()
	return Spot{
		ID: DetailID(tab), Axis: Column,
		Hit:  layout.Rect{X: l.Detail.X - 1, Y: in.Y, W: 1, H: in.H},
		Size: l.Detail.W, Min: layout.MinDetail, Max: max(l.Main.W+l.Detail.W-layout.MinList, layout.MinDetail),
		Origin: w - 1, Sign: -1,
	}
}

// rightEdge is the cell after the right padding of the pane whose content is
// in: the column of its right border.
func rightEdge(in layout.Rect) int { return in.X + in.W + layout.Pad }

// RunsRuleSpot is the rule under the header of the Runs view, in the area in,
// at row ruleRow of it. It runs over the padding from border to border and does
// not move: its size is fixed.
func RunsRuleSpot(in layout.Rect, ruleRow int) Spot {
	x := in.X - layout.Pad
	return Spot{
		ID: RunsRuleID, Axis: Row,
		Hit:  layout.Rect{X: x, Y: in.Y + ruleRow, W: rightEdge(in) - x, H: 1},
		Size: ruleRow, Min: ruleRow, Max: ruleRow,
		Origin: in.Y, Sign: 1,
	}
}

// RunsTreeSpot is the column between the tree of tasks and the log of the Runs
// view, in the area in: treeW cells from its left edge, from the row under the
// rule (bodyTop) to the bottom. lo and hi bound the cells of the tree.
func RunsTreeSpot(in layout.Rect, treeW, bodyTop, lo, hi int) Spot {
	return Spot{
		ID: RunsTreeID, Axis: Column,
		Hit:  layout.Rect{X: in.X + treeW, Y: in.Y + bodyTop, W: 1, H: max(in.H-bodyTop, 1)},
		Size: treeW, Min: lo, Max: hi,
		Origin: in.X, Sign: 1,
	}
}
