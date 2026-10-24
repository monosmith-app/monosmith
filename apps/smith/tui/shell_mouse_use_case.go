package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/footer"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pointer"
	"github.com/monosmith-app/monosmith/apps/smith/tui/tabs"
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
	switch s.lay().AreaAt(x, y) {
	case layout.Footer:
		if h, ok := footer.At(s.footerHints(), x); ok && h.Key != "" {
			return s.key(pressedKey(h.Key))
		}
		return nil
	case layout.Tabs:
		return s.tabClick(x)
	case layout.Main:
		if in := s.mainInner(); in.Contains(x, y) {
			return s.contentClick(x-in.X, y-in.Y, double)
		}
	case layout.Detail:
		in := s.detailInner()
		if in.Contains(x, y) {
			if h, ok := s.detailContent(in.W, in.H).hits[y-in.Y]; ok {
				return s.activate(h.id)
			}
		}
	}
	return nil
}

// tabClick opens the tab at column x of the strip; a scroll marker moves to
// the previous or the next one.
func (s *Shell) tabClick(x int) tea.Cmd {
	id := s.renderStrip("").At(x)
	at := -1
	for i, t := range s.tabs {
		if t == s.tab {
			at = i
		}
	}
	switch id {
	case "":
		return nil
	case tabs.Prev:
		return s.gotoTab(s.tabs[max(at-1, 0)])
	case tabs.Next:
		return s.gotoTab(s.tabs[min(at+1, len(s.tabs)-1)])
	}
	return s.gotoTab(id)
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
		s.list = s.list.SelectProject(h.id)
		s.detailCursor = 0
		if double {
			return s.paneKey(tea.KeyMsg{Type: tea.KeyEnter})
		}
	case "task":
		s.runSel = h.id
		s.keepRunVisible()
	case "finding":
		s.doctor = s.doctor.SelectRow(h.index)
	case "release":
		s.release = s.release.SelectRow(h.index)
	case "dep":
		s.deps = s.deps.SelectRow(h.index)
	case "option":
		if double {
			return s.paneKey(tea.KeyMsg{Type: tea.KeyEnter})
		}
		s.guide = s.guide.SelectRow(h.index)
	case "action":
		if s.detailOpen {
			for i, a := range s.detailActions() {
				if a.ID == h.id {
					s.detailCursor = i
				}
			}
		}
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
	switch s.lay().AreaAt(x, y) {
	case layout.Main:
		in := s.mainInner()
		switch s.tab {
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
			if s.detailOpen {
				return s.detailKey(map[int]string{-1: "up", 1: "down"}[dir])
			}
			s.list = s.list.Move(dir * wheelStep)
		case commands.TabDoctor, commands.TabGuide, commands.TabDeps, commands.TabRelease:
			key := tea.KeyMsg{Type: tea.KeyDown}
			if dir < 0 {
				key = tea.KeyMsg{Type: tea.KeyUp}
			}
			return s.paneKey(key)
		}
	}
	return nil
}
