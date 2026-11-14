package tui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/splitter"
	"github.com/monosmith-app/monosmith/apps/smith/tui/statusbar"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// lastFrame is the screen View drew last, shared by every copy of a Shell, so
// an event that changed nothing (a pointer motion that waits for the next
// hover frame) costs no drawing.
type lastFrame struct {
	out string
	ok  bool
}

// View implements tea.Model. It fills the whole terminal and can be called in
// any state, at any size: a panic while drawing is turned into a message
// instead of leaving the terminal in raw mode.
func (s Shell) View() string {
	if s.frame == nil {
		return s.render()
	}
	if !s.reuseFrame || !s.frame.ok {
		s.frame.out, s.frame.ok = s.render(), true
	}
	return s.frame.out
}

// render draws the screen.
func (s Shell) render() (out string) {
	defer func() {
		if r := recover(); r != nil {
			out = words.Say("tui.render.smith_the_interface_could_not_be", "r", r)
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

// belowMinimum is true while the terminal is too small to draw the screens.
// Its size is known (a size of zero is the start, before the first size).
func (s Shell) belowMinimum() bool {
	return s.w >= 1 && s.h >= 1 && (s.w < minWidth || s.h < minHeight)
}

// tooSmall is the one-line hint shown on a terminal below the minimum, cut to
// the width when even the hint does not fit.
func (s Shell) tooSmall() string {
	msg := words.Say("tui.render.terminal_too_small_x_enlarge_it", "w", s.w, "h", s.h, "minWidth", minWidth, "minHeight", minHeight)
	if pane.Width(msg) > s.w {
		msg = words.Say("tui.render.terminal_too_small_need_x", "minWidth", minWidth, "minHeight", minHeight)
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
	var rows []string
	if s.modalOpen() {
		// A selection inside a dialog is drawn over the dialog.
		rows = s.paintSelection(s.overlays(s.baseRows()))
	} else {
		rows = s.overlays(s.paintSelection(s.baseRows()))
	}
	for i := range rows {
		rows[i] = pane.Fit(rows[i], w)
	}
	return rows[:min(len(rows), h)]
}

// baseRows are the screen rows before the selection and the overlays are drawn
// over them: the text a drag selects.
func (s Shell) baseRows() []string {
	rows := make([]string, 0, s.h)
	strip := s.tabStrip()
	if s.inlineHeader() {
		rows = append(rows, s.tabRow(strip))
	} else {
		rows = append(rows, s.headerRow())
		rows = append(rows, strip.Lines...)
	}
	rows = append(rows, s.body()...)
	rows = append(rows, s.footerRow())
	rows = append(rows, statusbar.Render(s.th, s.statusInfo(), s.w))
	return s.paintSplits(rows)
}

// statusInfo is what the status bar shows: the exact command running or last
// run, its result, the branch and the version.
func (s Shell) statusInfo() statusbar.Info {
	info := statusbar.Info{Version: s.version, Hint: s.hint, Tick: s.tick, MouseOff: s.mouseOff}
	if s.updateOffered() {
		info.Update = s.upd.Latest
	}
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
// screen, the detail of the selection beside it. The main screen has no box: one
// rule under the tabs across the whole width, a one-cell margin at each side,
// and the divisions between the panels (see the layout's Frame).
func (s Shell) body() []string {
	l := s.lay()
	in, din := l.Main.Inner(), l.DetailInner()
	reserve := s.toastRows(in.H)
	spots := s.dividers()
	list := pane.Block(s.withRowHover(s.page().View(in.W, in.H-reserve), in).lines, in.W, in.H)
	var side []string
	if l.HasDetail() {
		side = pane.Block(s.withRowHover(s.sideContent(din.W, din.H-reserve), din).lines, din.W, din.H)
	}
	rows := make([]string, 0, l.Main.H)
	rows = append(rows, s.th.Border(strings.Repeat("─", l.Main.W+l.Detail.W), false))
	for i := range in.H {
		y := in.Y + i
		row := s.marginCell(spots, l.Main.X, y) + list[i]
		if l.HasDetail() {
			// A scrollbar in the list's last column moves onto the divider, so the
			// screen shows one line there, not two side by side.
			divider := s.th.Border("│", false)
			if glyph, ok := s.scrollbarCell(list[i], in.W); ok {
				row, divider = s.marginCell(spots, l.Main.X, y)+pane.Splice(list[i], in.W-1, 1, " "), glyph
			}
			row += divider + side[i] + s.marginCell(spots, din.X+din.W, y)
		} else {
			row += s.marginCell(spots, in.X+in.W, y)
		}
		rows = append(rows, row)
	}
	for _, j := range splitter.Junctions(spots, l.Main.Y) {
		if at := j.Y - l.Main.Y; at >= 0 && at < len(rows) {
			rows[at] = pane.Splice(rows[at], j.X, 1, s.th.Border(j.Glyph, false))
		}
	}
	return rows
}

// scrollbarCell is the last cell of a list line w cells wide, with its style,
// when that cell is a scrollbar glyph (the track or the thumb).
func (s Shell) scrollbarCell(line string, w int) (string, bool) {
	if w < 2 {
		return "", false
	}
	cell := ansi.Cut(line, w-1, w)
	glyph := pane.Strip(cell)
	return cell, glyph == s.th.Icons.Track || glyph == s.th.Icons.Thumb
}

// marginCell is the cell of a margin at (x, y): blank, or part of the rule when a
// horizontal division runs through it to the edge of the screen.
func (s Shell) marginCell(spots []splitter.Spot, x, y int) string {
	for _, p := range spots {
		if p.Axis == splitter.Row && p.Hit.Contains(x, y) {
			return s.th.Border("─", false)
		}
	}
	return " "
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

// overlays composes the palette, dialogs and toast over the rendered
// screen by cell, so nothing underneath is pushed down. Behind a modal the
// content area is emptied (its rule and dividers stay) so no half word of the
// view shows at the modal's edges, and the whole screen is drawn dimmed so
// nothing behind it looks interactive. The toast sits inside the content area.
func (s Shell) overlays(rows []string) []string {
	l := s.lay()
	if s.modalOpen() {
		keep := []int{}
		if l.HasDetail() {
			keep = append(keep, l.Detail.X-1)
		}
		rows = pane.Blank(rows, l.Main.Y, l.Footer.Y, keep...)
		for i := range rows {
			rows[i] = s.th.Dim(rows[i])
		}
	}
	if s.palOpen {
		box := s.paletteBox()
		lines := pane.Frame(s.th, words.Say("tui.render.command_palette"), s.pal.Lines(s.th, box.W-2, box.H-2), box.W, box.H, true)
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
	if s.relpick != nil {
		r := s.relpick.Rect()
		rows = pane.Overlay(rows, s.relpick.Lines(s.th), r.X, r.Y)
	}
	if s.confirm != nil {
		r := s.confirm.Rect()
		rows = pane.Overlay(rows, s.confirm.Pulse(s.tick, s.waitedForPlan()).Lines(s.th), r.X, r.Y)
	}
	if s.fix != nil {
		r := s.fix.Rect()
		fix := s.synced(*s.fix)
		rows = pane.Overlay(rows, fix.Lines(s.th, s.tick), r.X, r.Y)
	}
	if s.adopt != nil {
		r := s.adopt.Rect()
		rows = pane.Overlay(rows, s.adopt.Lines(s.th, s.tick), r.X, r.Y)
	}
	if s.compose != nil {
		r := s.compose.Rect()
		rows = pane.Overlay(rows, s.compose.Lines(s.th), r.X, r.Y)
	}
	if s.tools != nil {
		r := s.tools.Rect()
		rows = pane.Overlay(rows, s.tools.Lines(s.th), r.X, r.Y)
	}
	if s.cache != nil {
		r := s.cache.Rect()
		rows = pane.Overlay(rows, s.cache.Lines(s.th, s.tick, s.now()), r.X, r.Y)
	}
	if s.result != nil {
		r := s.result.Rect()
		rows = pane.Overlay(rows, s.result.Pulse(s.tick, s.outsideElapsed()).Lines(s.th), r.X, r.Y)
	}
	// The result dialog says the same in its own words: a toast over its bottom
	// edge would cut the frame.
	if s.result == nil {
		if lines, x, y, ok := s.toastSpot(); ok {
			rows = pane.Overlay(rows, lines, x, y)
		}
	}
	return rows
}

// toastSpot is the toast as drawn and where it sits. The content ends on the
// row above the footer: the toast sits on its last rows, one cell clear of the
// right edge. An open dialog keeps its border whole: the toast goes under the
// dialog or over it when the rows there are free, and inside its border, over
// its first rows, when the dialog takes the whole height.
func (s Shell) toastSpot() (lines []string, x, y int, ok bool) {
	if lines = s.toast.Lines(s.th, s.w-4); lines == nil {
		return nil, 0, 0, false
	}
	l, h := s.lay(), len(lines)
	right := func(lines []string, edge int) int { return max(edge-pane.Width(lines[0]), 1) }
	r, open := s.modalRect()
	switch {
	case !open || l.Footer.Y-(r.Y+r.H) >= h:
		return lines, right(lines, s.w-2), max(l.Footer.Y-h, l.Main.Y+1), true
	case r.Y-s.mainInner().Y >= h:
		return lines, right(lines, s.w-2), s.mainInner().Y, true
	}
	if lines = s.toast.Lines(s.th, r.W-2); lines == nil || r.H < len(lines)+2 {
		return nil, 0, 0, false
	}
	return lines, right(lines, r.X+r.W-1), r.Y + 1, true
}
