package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/splitter"
)

// detailDivider is the id of the divider beside the detail of the open tab.
func (s Shell) detailDivider() string { return splitter.DetailID(s.tab) }

// detailWidth is the width the person gave the detail of the open tab; 0 is
// automatic. A run open beside the projects takes half the screen unless the
// person chose a width: its log needs the room.
func (s Shell) detailWidth() int {
	n, _ := s.split.Size(s.detailDivider())
	if n == 0 && s.runWanted() {
		return layout.ClampDetailWidth(s.w, s.w/2)
	}
	return n
}

// splitBlocked is true while something covers the screen, so its dividers can
// be neither seen nor moved.
func (s Shell) splitBlocked() bool { return !s.inWorkspace || s.overlaid() }

// overlaid is true while a modal covers the screen, or the terminal is too
// small to draw it.
func (s Shell) overlaid() bool { return s.belowMinimum() || s.modalOpen() }

// modalOpen is true while a blocking overlay is open: a dialog, a wizard, a
// box or the palette. It is the one list of them. While it is true the mouse
// acts only on the modal and the screen behind it is drawn dimmed.
func (s Shell) modalOpen() bool {
	return s.palOpen || s.help != nil || s.pkg != nil || s.relpick != nil || s.confirm != nil ||
		s.fix != nil || s.adopt != nil || s.compose != nil || s.tools != nil || s.result != nil
}

// spots are the dividers of the open tab that can be moved now, for the splitter.
func (s Shell) spots() []splitter.Spot {
	if s.splitBlocked() {
		return nil
	}
	return s.dividers()
}

// dividers are the divisions of the open tab as drawn, whether or not they can
// be moved now: the vertical line between a list and its detail, and the
// horizontal lines between stacked panels.
func (s Shell) dividers() []splitter.Spot {
	var out []splitter.Spot
	l := s.lay()
	if l.HasDetail() {
		out = append(out, splitter.DetailSpot(s.tab, l, s.w))
	}
	in := l.Main.Inner()
	switch {
	case s.detailFull():
	case s.runShown():
		r := s.detailRect()
		g := s.runGeom(r.W, r.H)
		out = append(out, splitter.RunListSpot(r, g.listTop, g.sepRow, g.listRows))
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

// paintSplits draws the selected or held divider, and the junctions that touch
// it, over the screen rows.
func (s Shell) paintSplits(rows []string) []string {
	spots := s.spots()
	for _, p := range spots {
		rows = splitter.Paint(s.th, rows, p, s.split.Look(p.ID))
	}
	// The junctions follow, so a tee over a lit divider is lit with it.
	js := splitter.Junctions(spots, s.lay().Main.Y)
	for _, p := range spots {
		rows = splitter.PaintJunctions(s.th, rows, js, p, s.split.Look(p.ID))
	}
	return rows
}
