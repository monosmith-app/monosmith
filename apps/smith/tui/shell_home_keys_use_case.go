package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dashboard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

// homeKey moves the keyboard through the buttons and runs of Home in reading
// order (Up and Left go back, Down and Right go on), scrolls a page taller than
// the pane, and presses the target the keyboard is on with Enter.
func (s *Shell) homeKey(name string) tea.Cmd {
	targets := s.homePage(s.mainInner().W, dashboard.Away).Targets
	if s.homeCommand() == "" || len(targets) == 0 {
		return nil
	}
	sel := min(max(s.homeSel, 0), len(targets)-1)
	switch act := commands.NavNamed(name); {
	case act == commands.NavConfirm:
		t := targets[sel]
		if t.Command != "" {
			return s.activate(t.Command)
		}
		s.recentAt = t.Run
		return s.openRecent()
	case act == commands.NavUp || name == "left":
		sel--
	case act == commands.NavDown || name == "right":
		sel++
	case act == commands.NavHome:
		sel = 0
	case act == commands.NavEnd:
		sel = len(targets) - 1
	case act == commands.NavPageUp:
		s.pageHome(-1)
		return nil
	case act == commands.NavPageDown:
		s.pageHome(1)
		return nil
	default:
		return nil
	}
	s.homeSel = min(max(sel, 0), len(targets)-1)
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

// keepHomeVisible scrolls Home just enough to show the target the keyboard is on.
func (s *Shell) keepHomeVisible() {
	page := s.homePage(s.mainInner().W, dashboard.Away)
	if len(page.Targets) == 0 {
		return
	}
	row := page.Targets[min(max(s.homeSel, 0), len(page.Targets)-1)].Row
	total, h := len(page.Lines), s.mainInner().H
	top := pane.KeepVisible(s.homeTop, row, total, h)
	// Show the headline above a button too, when that costs the button no row.
	// Near the top of the page, show the page from its first line.
	context := max(row-homeContext, 0)
	want := pane.KeepVisible(top, context, total, h)
	if context <= homeContext+homeSummaryLines {
		want = 0
	}
	if want < top && pane.KeepVisible(want, row, total, h) == want {
		top = want
	}
	s.homeTop = top
}

// homeSummaryLines is the lines of the summary of the workspace that Home may show above its first item.
const homeSummaryLines = 2

// homeContext is how many lines above a button its headline is: the title and
// its note, which the page shows with the button when it can.
const homeContext = 2

// homeScrolls is true while the page of Home is taller than its pane.
func (s Shell) homeScrolls() bool {
	in := s.mainInner()
	return len(s.homePage(in.W, dashboard.Away).Lines) > in.H
}
