package splitter

import (
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/layout"
)

// The ids of the dividers a screen has besides the one beside the detail, which
// has one per tab (DetailID) so each view keeps its own width. RunListID,
// MinLogRows and MinRunListRows belong to the run panel that Projects had before
// the Runs view; no screen draws that divider now, and they stay for callers of
// this package.
const (
	RunListID        = "run.list"
	SettingsListID   = "settings.sections"
	SettingsDetailID = "settings.detail"
	// RunsTreeID is the divider between the tree and the log of the Runs view;
	// RunsRuleID the rule under its header, which does not move.
	RunsTreeID        = "runs.tree"
	RunsRuleID        = "runs.rule"
	detailDividerTail = ".detail"
	// MinLogRows is the rows the run log keeps; MinRunListRows the rows of the run list.
	MinLogRows     = 2
	MinRunListRows = 1
)

// DetailID is the id of the divider beside the detail of the tab.
func DetailID(tab string) string { return tab + detailDividerTail }

// RunListRows keeps the rows the person gave the run list between its least and
// what leaves the log its rows, with avail rows for both.
func RunListRows(given, avail int) int {
	return min(max(given, MinRunListRows), max(avail-MinLogRows, MinRunListRows))
}

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

// RunListSpot is the divider between the task list and the log of a run, in the
// area in; listTop is the row of the list, sepRow the row of the divider. It
// runs over the padding from border to border of the pane.
func RunListSpot(in layout.Rect, listTop, sepRow, listRows int) Spot {
	avail := max(in.H-listTop-2, 1)
	x := in.X - layout.Pad
	return Spot{
		ID: RunListID, Axis: Row,
		Hit:  layout.Rect{X: x, Y: in.Y + sepRow, W: rightEdge(in) - x, H: 1},
		Size: listRows, Min: MinRunListRows, Max: max(avail-MinLogRows, MinRunListRows),
		Origin: in.Y + listTop, Sign: 1,
	}
}

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

// SettingsListSpot is the column between the sections and the rest of the
// settings, in the area in.
func SettingsListSpot(in layout.Rect, listW, lo, hi int) Spot {
	return Spot{
		ID: SettingsListID, Axis: Column,
		Hit:  layout.Rect{X: in.X + listW, Y: in.Y, W: 1, H: in.H},
		Size: listW, Min: lo, Max: hi,
		Origin: in.X, Sign: 1,
	}
}

// SettingsDetailSpot is the row above the detail of a setting, in the area in.
// It starts at the column of the sections' divider (where the two meet) and
// runs to the right edge of the pane.
func SettingsDetailSpot(in layout.Rect, listW, sepRow, detailRows, lo, hi int) Spot {
	x := in.X + listW
	return Spot{
		ID: SettingsDetailID, Axis: Row,
		Hit:  layout.Rect{X: x, Y: in.Y + sepRow, W: rightEdge(in) - x, H: 1},
		Size: detailRows, Min: lo, Max: hi,
		Origin: in.Y + sepRow + detailRows, Sign: -1,
	}
}
