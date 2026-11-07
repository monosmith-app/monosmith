package splitter

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
)

// The ids of the dividers a screen has besides the one beside the detail, which
// has one per tab (DetailID) so each view keeps its own width.
const (
	RunListID         = "run.list"
	SettingsListID    = "settings.sections"
	SettingsDetailID  = "settings.detail"
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
// layout l of a screen w cells wide. It runs under the rule below the tabs.
func DetailSpot(tab string, l layout.Layout, w int) Spot {
	in := l.Main.Inner()
	return Spot{
		ID: DetailID(tab), Axis: Column,
		Hit:  layout.Rect{X: l.Detail.X - 1, Y: in.Y, W: 1, H: in.H},
		Size: l.Detail.W, Min: layout.MinDetail, Max: max(w-layout.MinList, layout.MinDetail),
		Origin: w - 1, Sign: -1,
	}
}

// rightEdge is the cell after the right margin of the pane whose content is in.
func rightEdge(in layout.Rect) int { return in.X + in.W + layout.MainFrame.Right }

// RunListSpot is the divider between the task list and the log of a run, in the
// area in; listTop is the row of the list, sepRow the row of the divider. It
// runs from edge to edge of the pane.
func RunListSpot(in layout.Rect, listTop, sepRow, listRows int) Spot {
	avail := max(in.H-listTop-2, 1)
	x := in.X - layout.MainFrame.Left
	return Spot{
		ID: RunListID, Axis: Row,
		Hit:  layout.Rect{X: x, Y: in.Y + sepRow, W: rightEdge(in) - x, H: 1},
		Size: listRows, Min: MinRunListRows, Max: max(avail-MinLogRows, MinRunListRows),
		Origin: in.Y + listTop, Sign: 1,
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
