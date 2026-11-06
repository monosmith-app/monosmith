package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/splitter"
)

// The ids of the dividers. The one beside the detail has one per tab, so each
// view keeps its own width.
const (
	splitRunList        = "run.list"
	splitSettingsList   = "settings.sections"
	splitSettingsDetail = "settings.detail"
	detailDividerSuffix = ".detail"
	minLogRows          = 2
	minRunListRows      = 1
)

// detailDivider is the id of the divider beside the detail of the open tab.
func (s Shell) detailDivider() string { return s.tab + detailDividerSuffix }

// detailWidth is the width the person gave the detail of the open tab; 0 is automatic.
func (s Shell) detailWidth() int {
	n, _ := s.split.Size(s.detailDivider())
	return n
}

// splitBlocked is true while something covers the screen, so its dividers can
// be neither seen nor moved.
func (s Shell) splitBlocked() bool {
	return !s.inWorkspace || s.belowMinimum() || s.compose != nil || s.tools != nil || s.result != nil ||
		s.adopt != nil || s.fix != nil || s.pkg != nil || s.relpick != nil || s.confirm != nil || s.help != nil || s.palOpen
}

// spots are the dividers of the open tab as drawn now, for the splitter.
func (s Shell) spots() []splitter.Spot {
	if s.splitBlocked() {
		return nil
	}
	var out []splitter.Spot
	l := s.lay()
	if l.HasDetail() {
		out = append(out, splitter.Spot{
			ID: s.detailDivider(), Axis: splitter.Column,
			Hit:  layout.Rect{X: l.Detail.X - 1, Y: l.Main.Y, W: 1, H: l.Main.H},
			Size: l.Detail.W, Min: layout.MinDetail, Max: max(s.w-layout.MinList, layout.MinDetail),
			Origin: s.w - 1, Sign: -1,
		})
	}
	in := l.Main.Inner()
	switch {
	case s.detailFull():
	case s.tab == commands.TabRun && s.run != nil:
		g := s.runGeom(in.W, in.H)
		avail := max(in.H-g.listTop-2, 1)
		out = append(out, splitter.Spot{
			ID: splitRunList, Axis: splitter.Row,
			Hit:  layout.Rect{X: in.X, Y: in.Y + g.sepRow, W: in.W, H: 1},
			Size: g.listRows, Min: minRunListRows, Max: max(avail-minLogRows, minRunListRows),
			Origin: in.Y + g.listTop, Sign: 1,
		})
	case s.tab == commands.TabSettings:
		g := s.settings.Geometry(in.W, in.H)
		out = append(out, splitter.Spot{
			ID: splitSettingsList, Axis: splitter.Column,
			Hit:  layout.Rect{X: in.X + g.ListW, Y: in.Y, W: 1, H: in.H},
			Size: g.ListW, Min: g.ListMin, Max: g.ListMax,
			Origin: in.X, Sign: 1,
		})
		if g.SepRow >= 0 {
			out = append(out, splitter.Spot{
				ID: splitSettingsDetail, Axis: splitter.Row,
				Hit:  layout.Rect{X: in.X + g.ListW + 1, Y: in.Y + g.SepRow, W: in.W - g.ListW - 1, H: 1},
				Size: g.DetailRows, Min: g.DetailMin, Max: g.DetailMax,
				Origin: in.Y + g.SepRow + g.DetailRows, Sign: -1,
			})
		}
	}
	return out
}

// applySplits hands the sizes the person chose to the views that lay
// themselves out.
func (s *Shell) applySplits() {
	list, _ := s.split.Size(splitSettingsList)
	rows, _ := s.split.Size(splitSettingsDetail)
	s.settings = s.settings.WithSizes(list, rows)
}

// paintSplits draws the selected or held divider over the screen rows.
func (s Shell) paintSplits(rows []string) []string {
	for _, p := range s.spots() {
		rows = splitter.Paint(s.th, rows, p, s.split.Look(p.ID))
	}
	return rows
}
