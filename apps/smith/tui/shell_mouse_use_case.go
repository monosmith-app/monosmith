package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/footer"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/navigator"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pointer"
)

// wheelStep is how many lines one notch of the wheel scrolls.
const wheelStep = 3

// paletteBox is where the command palette is drawn.
func (s Shell) paletteBox() layout.Rect {
	w, h := min(max(s.w-4, 10), 76), min(max(s.h-4, 5), 20)
	return layout.Rect{X: (s.w - w) / 2, Y: max((s.h-h)/3, 1), W: w, H: h}
}

func (s *Shell) mouse(msg tea.MouseMsg) tea.Cmd {
	e := tea.MouseEvent(msg)
	s.mouseX, s.mouseY = e.X, e.Y
	switch {
	case e.Action == tea.MouseActionMotion:
		s.hover(e.X, e.Y)
	case e.Action != tea.MouseActionPress:
		// Releases need no handling: everything happens on the press.
	case e.Button == tea.MouseButtonWheelUp:
		return s.wheel(e.X, e.Y, -1)
	case e.Button == tea.MouseButtonWheelDown:
		return s.wheel(e.X, e.Y, 1)
	case e.Button == tea.MouseButtonLeft:
		return s.click(e.X, e.Y)
	}
	return nil
}

// hover highlights what is under the pointer in the boxes that keep their own
// cursor; panes read the pointer position when they are drawn.
func (s *Shell) hover(x, y int) {
	switch {
	case s.confirm != nil:
		c := s.confirm.Hover(x, y)
		s.confirm = &c
	case s.palOpen:
		box := s.paletteBox()
		if box.Contains(x, y) {
			s.pal = s.pal.Hover(s.pal.RowAt(y-box.Y-2, max(box.H-5, 1)))
		}
	}
}

func (s *Shell) click(x, y int) tea.Cmd {
	now := s.now()
	press := pointer.Click{X: x, Y: y, At: now}
	double := pointer.IsDouble(s.lastClick, press)
	s.lastClick = press
	if double {
		s.lastClick = pointer.Click{}
	}
	switch {
	case s.confirm != nil:
		c, out := s.confirm.Click(x, y)
		s.confirm = &c
		return s.afterConfirm(c, out)
	case s.about != nil:
		if s.about.Click(x, y).Done {
			s.about = nil
		}
		return nil
	case s.help != nil:
		if s.help.Click(x, y).Done {
			s.help = nil
		}
		return nil
	case s.palOpen:
		box := s.paletteBox()
		if !box.Contains(x, y) {
			s.palOpen = false
			s.pal = s.pal.Reset()
			return nil
		}
		if i := s.pal.RowAt(y-box.Y-2, max(box.H-5, 1)); i >= 0 {
			p, res := s.pal.Choose(i)
			s.pal = p
			return s.afterPalette(res)
		}
		return nil
	}
	switch s.lay.AreaAt(x, y) {
	case layout.Footer:
		if h, ok := footer.At(s.footerHints(), x); ok && h.Key != "" {
			return s.key(pressedKey(h.Key))
		}
		return nil
	case layout.Navigator:
		s.focus = layout.Navigator
		return s.navigatorClick(x, y, double)
	case layout.Main:
		s.focus = layout.Main
		if y == s.lay.Main.Y {
			for _, sp := range s.tabSpans() {
				if x >= sp.x0 && x < sp.x1 {
					return s.gotoTab(sp.id)
				}
			}
			return nil
		}
		if in := s.mainInner(); in.Contains(x, y) {
			return s.contentClick(x-in.X, y-in.Y, double)
		}
	case layout.Inspector:
		s.focus = layout.Inspector
		in := s.inspectorInner()
		_, hits := s.inspectorPanel().View(s.th, in.W, "")
		if id, ok := hits[y-in.Y]; ok {
			for i, a := range s.inspectorActions() {
				if a.ID == id {
					s.inspCursor = i
				}
			}
			return s.activate(id)
		}
	}
	return nil
}

// navigatorClick selects the row clicked, opens or closes it on its arrow, and
// runs a task or toggles a branch on a double click.
func (s *Shell) navigatorClick(x, y int, double bool) tea.Cmd {
	in := s.navigatorInner()
	i := s.tree.SetHeight(max(in.H, 1)).RowAt(y - in.Y)
	if i < 0 {
		return nil
	}
	s.tree = s.tree.Select(i)
	s.inspCursor = 0
	row := s.tree.Rows()[i]
	if double {
		return s.openSelected()
	}
	arrowX := in.X + 1 + 2*row.Depth
	if row.Expandable && x >= arrowX && x < arrowX+2 {
		s.tree = s.tree.Toggle()
	}
	return nil
}

// openSelected is the double click and Enter on a navigator row: a task runs,
// a project or group opens or closes.
func (s *Shell) openSelected() tea.Cmd {
	if s.tree.Selected().Kind == navigator.TaskRow {
		return s.activate("run.selection")
	}
	s.tree = s.tree.Toggle()
	return nil
}

// contentClick handles a click at (rx, ry) in the main body.
func (s *Shell) contentClick(rx, ry int, double bool) tea.Cmd {
	in := s.mainInner()
	c := s.mainContent(in.W, in.H)
	h, ok := c.hits[ry]
	if !ok {
		return nil
	}
	switch h.kind {
	case "project":
		s.tree = s.tree.SelectProject(h.id)
		s.inspCursor = 0
	case "task":
		s.runSel = h.id
		s.keepRunVisible()
	case "finding":
		s.doctor = s.doctor.SelectRow(h.index)
	case "recipe":
		s.recipes = s.recipes.SelectRow(h.index)
	case "release":
		s.release = s.release.SelectRow(h.index)
	case "plugin":
		s.plugins = s.plugins.SelectRow(h.index)
	case "dep":
		s.deps = s.deps.SelectRow(h.index)
	case "setting":
		s.settings = s.settings.SelectRow(h.index)
		if double {
			return s.paneKey(tea.KeyMsg{Type: tea.KeyEnter})
		}
	case "option":
		if double {
			return s.paneKey(tea.KeyMsg{Type: tea.KeyEnter})
		}
		s.guide = s.guide.SelectRow(h.index)
	case "tree":
		s.tree = s.tree.SetHeight(max(in.H, 1))
		if i := s.tree.RowAt(ry); i >= 0 {
			s.tree = s.tree.Select(i)
			if double {
				return s.openSelected()
			}
		}
	case "action":
		return s.activate(h.id)
	}
	return nil
}

func (s *Shell) wheel(x, y, dir int) tea.Cmd {
	switch {
	case s.confirm != nil || s.about != nil:
		return nil
	case s.help != nil:
		h := s.help.Wheel(dir*wheelStep, s.help.Total(s.th))
		s.help = &h
		return nil
	case s.palOpen:
		s.pal = s.pal.Scroll(dir * wheelStep)
		return nil
	}
	switch s.lay.AreaAt(x, y) {
	case layout.Navigator:
		s.tree = s.tree.Scroll(dir * wheelStep)
	case layout.Main:
		in := s.mainInner()
		switch s.tab {
		case tabNavigator:
			s.tree = s.tree.Scroll(dir * wheelStep)
		case commands.TabRun:
			if s.run == nil {
				return nil
			}
			g := s.runGeom(in.W, in.H)
			if y-in.Y >= g.logTop {
				s.log = s.log.Scroll(dir*wheelStep, len(s.logLines()), g.logRows)
			} else {
				s.runTop = max(min(s.runTop+dir, len(s.runIDs())-g.listRows), 0)
			}
		case commands.TabProjects:
			s.tree = s.tree.Move(dir)
		case commands.TabDoctor, commands.TabSettings, commands.TabGuide, commands.TabPlugins, commands.TabDeps, commands.TabRelease, commands.TabRecipes:
			key := tea.KeyMsg{Type: tea.KeyDown}
			if dir < 0 {
				key = tea.KeyMsg{Type: tea.KeyUp}
			}
			return s.paneKey(key)
		}
	}
	return nil
}
