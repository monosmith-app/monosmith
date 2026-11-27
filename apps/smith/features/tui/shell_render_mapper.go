package tui

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/splitter"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/statusbar"
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

// screen is the whole terminal as h lines of w cells: the title bar, the strip
// of views (narrow) or the rail (wide), the content, the footer and the status bar.
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
	if s.fullRun() {
		// Only the top bar and the run: nothing else shares the terminal.
		return s.paintSplits(append([]string{s.fullBar()}, s.body()...))
	}
	l := s.lay()
	rows := make([]string, 0, s.h)
	if l.Header.H > 0 {
		rows = append(rows, s.headerRow())
	}
	if l.Tabs.H > 0 {
		rows = append(rows, s.tabStrip().Lines...)
	}
	blank := strings.Repeat(" ", s.w)
	for len(rows) < l.Main.Y {
		rows = append(rows, blank)
	}
	rows = append(rows, s.body()...)
	for len(rows) < l.Footer.Y {
		rows = append(rows, blank)
	}
	rows = append(rows, s.footerRow())
	rows = append(rows, statusbar.Render(s.th, s.statusInfo(), s.w))
	return s.paintSplits(rows)
}

// statusInfo is what the status bar shows: the exact command running or last
// run, its result and counts, the branch, the version and the doctor state.
func (s Shell) statusInfo() statusbar.Info {
	info := statusbar.Info{Version: s.version, Hint: s.hint, Tick: s.tick, MouseOff: s.mouseOff, Branch: s.branch, Health: s.healthSummary()}
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
		info.Failed, info.Passed, info.Took = c.Failed, c.Passed, shortDuration(s.run.RunTime(s.now()))
	}
	return info
}

// body draws the rows of the content: on a wide screen the rail of views,
// then the frame of the list (or of the detail that took its place) and the
// frame of the detail of the selection beside it. Every pane is a rounded box
// with its title in the top border and one cell of padding inside each side
// (see the layout's Frame).
func (s Shell) body() []string {
	l := s.lay()
	in, din := l.Main.Inner(), l.DetailInner()
	reserve := s.toastRows(in.H)
	spots := s.dividers()
	list := pane.Block(s.withRowHover(s.page().View(in.W, in.H-reserve), in).lines, in.W, in.H)
	focused := s.mainFocused()
	main := s.framed(s.mainTitle(), l.Main, in, list, focused, spots)
	var side []string
	if l.HasDetail() {
		lines := pane.Block(s.withRowHover(s.sideContent(din.W, din.H-reserve), din).lines, din.W, din.H)
		side = s.framed(s.detailTitle(), l.Detail, din, lines, !focused, spots)
	}
	railLines := s.renderRail(s.hoveredRail()).Lines
	rows := make([]string, l.Main.H)
	for i := range rows {
		row := ""
		if i < len(railLines) {
			row = railLines[i] + strings.Repeat(" ", max(l.Main.X-l.Rail.W, 0))
			if i == 0 && l.Main.X == l.Rail.W {
				// The rail's top rule runs into the frame: its corner becomes a tee.
				main[0] = pane.Splice(main[0], 0, 1, s.th.Border("┬", focused))
			}
		}
		row += main[i]
		if i < len(side) {
			row += side[i]
		}
		rows[i] = row
	}
	for _, j := range splitter.Junctions(spots, l.Main.Y, l.Main.Y+l.Main.H-1) {
		if at := j.Y - l.Main.Y; at >= 0 && at < len(rows) {
			rows[at] = pane.Splice(rows[at], j.X, 1, s.th.Border(j.Glyph, focused))
		}
	}
	return rows
}

// framed draws a pane at box: its lines, drawn for the inside in, with one
// cell of padding at each side, inside a rounded frame with the title in the
// top border. A horizontal division that runs through the padding is drawn
// there as part of its rule, in the frame's colour, and meets each side
// border with a tee.
func (s Shell) framed(title string, box, in layout.Rect, lines []string, focused bool, spots []splitter.Spot) []string {
	body := make([]string, len(lines))
	ruled := map[int]bool{}
	for i, line := range lines {
		y := in.Y + i
		left, right := s.padCell(spots, in.X-layout.Pad, y, focused), s.padCell(spots, in.X+in.W, y, focused)
		ruled[i] = left != " " && right != " "
		body[i] = left + line + right
	}
	rows := pane.Frame(s.th, pane.CutEllipsis(title, max(box.W-6, 1)), body, box.W, box.H, focused)
	for i := range ruled {
		if ruled[i] && i+1 < len(rows)-1 {
			rows[i+1] = pane.Splice(rows[i+1], 0, 1, s.th.Border("├", focused))
			rows[i+1] = pane.Splice(rows[i+1], box.W-1, 1, s.th.Border("┤", focused))
		}
	}
	return rows
}

// padCell is the padding cell at (x, y): blank, or part of the rule, in the
// colour of a frame that is focused or not, when a horizontal division runs
// through it to the border of the pane.
func (s Shell) padCell(spots []splitter.Spot, x, y int, focused bool) string {
	for _, p := range spots {
		if p.Axis == splitter.Row && p.Hit.Contains(x, y) {
			return s.th.Border("─", focused)
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
// inside of the content frames is emptied (the frames and the rail stay) so no
// half word of the view shows at the modal's edges, and the whole screen is
// drawn dimmed so nothing behind it looks interactive. The toast sits inside
// the content area.
func (s Shell) overlays(rows []string) []string {
	l := s.lay()
	if s.modalOpen() {
		for _, r := range []layout.Rect{l.Main.Inner(), l.DetailInner()} {
			rows = blankRect(rows, layout.Rect{X: r.X - layout.Pad, Y: r.Y, W: r.W + 2*layout.Pad, H: r.H})
		}
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
	if s.fix != nil && !s.inlineFix() {
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
// row above the bottom border of its frames: the toast sits on its last rows,
// inside the right border. An open dialog keeps its border whole: the toast goes under the
// dialog or over it when the rows there are free, and inside its border, over
// its first rows, when the dialog takes the whole height.
func (s Shell) toastSpot() (lines []string, x, y int, ok bool) {
	if lines = s.toast.Lines(s.th, s.w-4); lines == nil {
		return nil, 0, 0, false
	}
	l, h := s.lay(), len(lines)
	right := func(lines []string, edge int) int { return max(edge-pane.Width(lines[0]), 1) }
	bottom := l.Main.Y + l.Main.H - 1
	r, open := s.modalRect()
	switch {
	case !open || bottom-(r.Y+r.H) >= h:
		return lines, right(lines, s.w-1-layout.Pad), max(bottom-h, l.Main.Y+1), true
	case r.Y-s.mainInner().Y >= h:
		return lines, right(lines, s.w-1-layout.Pad), s.mainInner().Y, true
	}
	if lines = s.toast.Lines(s.th, r.W-2); lines == nil || r.H < len(lines)+2 {
		return nil, 0, 0, false
	}
	return lines, right(lines, r.X+r.W-1), r.Y + 1, true
}

// blankRect empties the cells of r in the rows (one string per screen row).
func blankRect(rows []string, r layout.Rect) []string {
	if r.W <= 0 {
		return rows
	}
	blank := strings.Repeat(" ", r.W)
	for y := max(r.Y, 0); y < r.Y+r.H && y < len(rows); y++ {
		rows[y] = pane.Splice(rows[y], r.X, r.W, blank)
	}
	return rows
}
