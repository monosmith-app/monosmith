package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/dashboard"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

// homeKey moves the keyboard through the list of Home: Up and Down go from
// item to item, Left and Right from button to button of the item, Home and End
// to the first and the last item. PgUp and PgDn scroll a page taller than the
// pane, and Enter presses the button the keyboard is on.
func (s *Shell) homeKey(name string) tea.Cmd {
	q := dashboard.Queue(s.dashboardFacts())
	if dashboard.Primary(q) == "" {
		return nil
	}
	item := min(max(s.homeSel, 0), len(q)-1)
	btn := min(max(s.homeBtn, 0), max(len(q[item].Buttons)-1, 0))
	switch act := commands.NavNamed(name); {
	case act == commands.NavConfirm:
		if len(q[item].Buttons) == 0 {
			return nil
		}
		return s.activate(q[item].Buttons[btn].CommandID)
	case act == commands.NavUp:
		item, btn = item-1, 0
	case act == commands.NavDown:
		item, btn = item+1, 0
	case name == "left":
		btn--
	case name == "right":
		btn++
	case act == commands.NavHome:
		item, btn = 0, 0
	case act == commands.NavEnd:
		item, btn = len(q)-1, 0
	case act == commands.NavPageUp:
		s.pageHome(-1)
		return nil
	case act == commands.NavPageDown:
		s.pageHome(1)
		return nil
	default:
		return nil
	}
	s.homeSel = min(max(item, 0), len(q)-1)
	s.homeBtn = min(max(btn, 0), max(len(q[s.homeSel].Buttons)-1, 0))
	s.keepHomeVisible()
	return nil
}

// pageHome scrolls Home by a page up (dir -1) or down (dir 1); the keyboard
// stays on its target.
func (s *Shell) pageHome(dir int) { s.scrollHome(dir * pane.PageStep(s.mainInner().H)) }

// scrollHome moves Home by delta lines, never past its ends.
func (s *Shell) scrollHome(delta int) {
	lines := len(s.homePage(s.mainInner().W, dashboard.Away).Lines)
	s.homeTop = pane.ClampTop(s.homeTop+delta, lines, s.mainInner().H)
}

// keepHomeVisible scrolls Home just enough to show both lines of the item the
// keyboard is on. With the first item, the page shows from its top when that
// still shows the item, so the boxes stay in view.
func (s *Shell) keepHomeVisible() {
	page := s.homePage(s.mainInner().W, dashboard.Away)
	if len(page.Items) == 0 {
		return
	}
	r := page.Items[min(max(s.homeSel, 0), len(page.Items)-1)]
	total, h := len(page.Lines), s.mainInner().H
	top := pane.KeepVisible(s.homeTop, r.To-1, total, h)
	top = pane.KeepVisible(top, r.From, total, h)
	if s.homeSel <= 0 && pane.KeepVisible(0, r.To-1, total, h) == 0 {
		top = 0
	}
	s.homeTop = top
}

// homeScrolls is true while the page of Home is taller than its pane.
func (s Shell) homeScrolls() bool {
	in := s.mainInner()
	return len(s.homePage(in.W, dashboard.Away).Lines) > in.H
}
