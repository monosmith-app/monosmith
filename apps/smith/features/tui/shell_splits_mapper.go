package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/splitter"
)

// detailDivider is the id of the divider beside the detail of the open tab.
func (s Shell) detailDivider() string { return splitter.DetailID(s.tab) }

// projectsListMax is the most columns the list of projects takes beside the
// detail while the person has not moved the divider.
const projectsListMax = 44

// detailWidth is the width the person gave the detail of the open tab; 0 is
// automatic. The detail of a project takes what a list of at most 44 columns
// leaves.
func (s Shell) detailWidth() int {
	n, _ := s.split.Size(s.detailDivider())
	cw := layout.ContentWidth(s.w)
	switch {
	case n == 0 && s.tab == commands.TabProjects && s.inWorkspace:
		// The list of projects needs little room; the sections of the detail
		// (the table of tasks) take the rest.
		return layout.ClampDetailWidth(cw, cw-pane.Clamp(cw*2/5, layout.MinList, projectsListMax))
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
		s.fix != nil || s.adopt != nil || s.compose != nil || s.tools != nil || s.result != nil || s.cache != nil
}

// spots are the dividers of the open tab that can be moved now, for the splitter.
func (s Shell) spots() []splitter.Spot {
	if s.splitBlocked() {
		return nil
	}
	var out []splitter.Spot
	for _, p := range s.dividers() {
		// The rule under the header of the Runs view is drawn, never moved.
		if p.ID != splitter.RunsRuleID {
			out = append(out, p)
		}
	}
	return out
}

// dividers are the divisions of the open tab as drawn, whether or not they can
// be moved now: the vertical line between a list and its detail, and the
// horizontal lines between stacked panels.
func (s Shell) dividers() []splitter.Spot {
	var out []splitter.Spot
	l := s.lay()
	// Settings has dividers of its own: they come first, so the first key selects the one inside the tab.
	if l.HasDetail() && s.tab != commands.TabSettings {
		out = append(out, splitter.DetailSpot(s.tab, l, s.w))
	}
	in := l.Main.Inner()
	switch {
	case s.runShown():
		w, h := s.runRect()
		g := s.runsGeom(w, h)
		out = append(out, splitter.RunsRuleSpot(in, g.ruleRow))
		if !g.narrow {
			out = append(out, splitter.RunsTreeSpot(in, g.treeW, g.bodyTop, runsTreeMin, max(w-runsLogMin, runsTreeMin)))
		}
	case s.detailFull():
	case s.tab == commands.TabSettings:
		g := s.settings.WithSideDetail(l.HasDetail()).Geometry(in.W, in.H)
		out = append(out, splitter.SettingsListSpot(in, g.ListW, g.ListMin, g.ListMax))
		if g.SepRow >= 0 {
			out = append(out, splitter.SettingsDetailSpot(in, g.ListW, g.SepRow, g.DetailRows, g.DetailMin, g.DetailMax))
		}
	}
	if l.HasDetail() && s.tab == commands.TabSettings {
		out = append(out, splitter.DetailSpot(s.tab, l, s.w))
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
	l := s.lay()
	js := splitter.Junctions(spots, l.Main.Y, l.Main.Y+l.Main.H-1)
	for _, p := range spots {
		rows = splitter.PaintJunctions(s.th, rows, js, p, s.split.Look(p.ID))
	}
	return rows
}
