package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/statusbar"
	"github.com/monosmith-app/monosmith/apps/smith/tui/tabs"
)

// View implements tea.Model. It fills the whole terminal and can be called in
// any state, at any size: a panic while drawing is turned into a message
// instead of leaving the terminal in raw mode.
func (s Shell) View() (out string) {
	defer func() {
		if r := recover(); r != nil {
			out = fmt.Sprintf("smith: the interface could not be drawn (%v). Press Ctrl+C to leave.", r)
		}
	}()
	if s.w < 1 || s.h < 1 {
		return ""
	}
	if s.w < minWidth || s.h < minHeight {
		return s.tooSmall()
	}
	return strings.Join(s.screen(), "\n")
}

// The least terminal the screens are laid out for; below it a one-line hint
// replaces them instead of a jumble.
const (
	minWidth  = 72
	minHeight = 9
)

// tooSmall is the one-line hint shown on a terminal below the minimum, cut to
// the width when even the hint does not fit.
func (s Shell) tooSmall() string {
	msg := fmt.Sprintf("Terminal too small (%dx%d): enlarge it to at least %dx%d, or press Ctrl+C to leave.", s.w, s.h, minWidth, minHeight)
	if pane.Width(msg) > s.w {
		msg = fmt.Sprintf("Terminal too small: need %dx%d", minWidth, minHeight)
	}
	// One line of hint; the rest stays blank, so the screen still has exactly
	// the terminal's size.
	rows := []string{pane.Fit(msg, s.w)}
	for len(rows) < s.h {
		rows = append(rows, strings.Repeat(" ", s.w))
	}
	return strings.Join(rows, "\n")
}

// screen is the whole terminal as h lines of w cells: the header, the tabs,
// the content, the footer and the status bar.
func (s Shell) screen() []string {
	w, h := s.w, s.h
	rows := make([]string, 0, h)
	strip := s.tabStrip()
	rows = append(rows, s.headerRow())
	rows = append(rows, strip.Lines[:len(strip.Lines)-s.tabsInBody()]...)
	rows = append(rows, s.body(strip)...)
	rows = append(rows, s.footerRow())
	rows = append(rows, statusbar.Render(s.th, s.statusInfo(), w))
	rows = s.overlays(rows)
	for i := range rows {
		rows[i] = pane.Fit(rows[i], w)
	}
	return rows[:min(len(rows), h)]
}

// statusInfo is what the status bar shows: the exact command running or last
// run, its result, the branch and the version.
func (s Shell) statusInfo() statusbar.Info {
	info := statusbar.Info{Version: s.version, Hint: s.hint, Tick: s.tick, MouseOff: s.mouseOff}
	if s.run == nil {
		return info
	}
	info.Command = s.run.CommandLine()
	c := s.run.Counts()
	switch {
	case !s.run.Finished:
		info.State = statusbar.Running
		if c.Total > 0 {
			info.Result = fmtCount(c.Done, c.Total)
		}
	case s.run.Cancelled:
		info.State = statusbar.Cancelled
	case s.run.OK():
		info.State = statusbar.Passed
	default:
		info.State = statusbar.Failed
	}
	if s.run.Finished {
		info.Result = s.summary()
	}
	return info
}

// body draws the list, or the detail that took its place, and, on a wide
// screen, the detail of the selection in the same frame: one border all round,
// the list's right border dividing the two, and the top edge shared with the
// tabs when they are tall.
func (s Shell) body(strip tabs.Strip) []string {
	l := s.lay()
	page := s.page()
	in := l.Main.Inner()
	title := page.Title()
	if l.Tall {
		// The strip's last row is the pane's top edge, and the title sits in the strip.
		title = ""
	}
	reserve := s.toastRows(in.H)
	main := pane.Frame(s.th, title, page.View(in.W, in.H-reserve).lines, l.Main.W, l.Main.H, false)
	edge := func(r string) string { return s.th.Border(r, false) }
	if !l.HasDetail() {
		if l.Tall {
			main[0] = strip.Lines[len(strip.Lines)-1]
		}
		return main
	}
	din := l.DetailInner()
	side := pane.Block(s.detailContent(din.W, din.H-reserve).lines, din.W, din.H)
	cut := func(line string) string { return ansi.Truncate(line, l.Main.W-1, "") + "\x1b[m" }
	out := make([]string, 0, len(main))
	if l.Tall {
		out = append(out, strip.Lines[len(strip.Lines)-1])
	} else {
		out = append(out, cut(main[0])+edge("┬"+strings.Repeat("─", max(l.Detail.W-2, 0))+"╮"))
	}
	for y := 1; y < len(main)-1; y++ {
		out = append(out, main[y]+side[y-1]+edge("│"))
	}
	return append(out, cut(main[len(main)-1])+edge("┴"+strings.Repeat("─", max(l.Detail.W-2, 0))+"╯"))
}

// toastRows is how many rows at the foot of the content stay empty while a
// toast is shown, so the toast box never covers a line of the view. A small
// screen keeps all its rows, and the toast then sits over the last ones.
func (s Shell) toastRows(innerH int) int {
	const box = 3
	if !s.toast.Shown() || innerH < 2*box {
		return 0
	}
	return box
}

// modalOpen is true while a box that takes the keyboard is drawn over the screen.
func (s Shell) modalOpen() bool {
	return s.palOpen || s.help != nil || s.pkg != nil || s.confirm != nil || s.fix != nil || s.adopt != nil
}

// overlays composes the palette, dialogs and toast over the rendered
// screen by cell, so nothing underneath is pushed down. Behind a modal the
// content area is emptied (its frame stays), so no half word of the view shows
// at the modal's edges; the toast sits inside the content area, never on its
// frame.
func (s Shell) overlays(rows []string) []string {
	l := s.lay()
	if s.modalOpen() {
		rows = pane.Blank(rows, l.Main.Y, l.Footer.Y)
	}
	if s.palOpen {
		box := s.paletteBox()
		lines := pane.Frame(s.th, "Command palette", s.pal.Lines(s.th, box.W-2, box.H-2), box.W, box.H, true)
		rows = pane.Overlay(rows, lines, box.X, box.Y)
	}
	if s.help != nil {
		r := s.help.Rect()
		rows = pane.Overlay(rows, s.help.Lines(s.th), r.X, r.Y)
	}
	if s.pkg != nil {
		r := s.pkg.Rect()
		rows = pane.Overlay(rows, s.pkg.Lines(s.th), r.X, r.Y)
	}
	if s.confirm != nil {
		r := s.confirm.Rect()
		rows = pane.Overlay(rows, s.confirm.Lines(s.th), r.X, r.Y)
	}
	if s.fix != nil {
		r := s.fix.Rect()
		rows = pane.Overlay(rows, s.fix.Lines(s.th, s.tick), r.X, r.Y)
	}
	if s.adopt != nil {
		r := s.adopt.Rect()
		rows = pane.Overlay(rows, s.adopt.Lines(s.th, s.tick), r.X, r.Y)
	}
	if lines := s.toast.Lines(s.th, s.w-4); lines != nil {
		// The frame's bottom edge is the row above the footer: the toast sits
		// on the rows above it, one cell clear of the right edge.
		y := max(l.Footer.Y-1-len(lines), l.Main.Y+1)
		rows = pane.Overlay(rows, lines, max(s.w-2-pane.Width(lines[0]), 1), y)
	}
	return rows
}
