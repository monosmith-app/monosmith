package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/splitter"
)

// detailDivider is the id of the divider beside the detail of the open tab.
func (s Shell) detailDivider() string { return splitter.DetailID(s.tab) }

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
		out = append(out, splitter.DetailSpot(s.tab, l, s.w))
	}
	in := l.Main.Inner()
	switch {
	case s.detailFull():
	case s.tab == commands.TabRun && s.run != nil:
		g := s.runGeom(in.W, in.H)
		out = append(out, splitter.RunListSpot(in, g.listTop, g.sepRow, g.listRows))
	case s.tab == commands.TabSettings:
		g := s.settings.Geometry(in.W, in.H)
		out = append(out, splitter.SettingsListSpot(in, g.ListW, g.ListMin, g.ListMax))
		if g.SepRow >= 0 {
			out = append(out, splitter.SettingsDetailSpot(in, g.ListW, g.SepRow, g.DetailRows, g.DetailMin, g.DetailMax))
		}
	}
	return out
}

// applySplits hands the sizes the person chose to the views that lay
// themselves out.
func (s *Shell) applySplits() {
	list, _ := s.split.Size(splitter.SettingsListID)
	rows, _ := s.split.Size(splitter.SettingsDetailID)
	s.settings = s.settings.WithSizes(list, rows)
}

// paintSplits draws the selected or held divider over the screen rows.
func (s Shell) paintSplits(rows []string) []string {
	for _, p := range s.spots() {
		rows = splitter.Paint(s.th, rows, p, s.split.Look(p.ID))
	}
	return rows
}
