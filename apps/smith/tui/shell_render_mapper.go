package tui

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/statusbar"
)

// View implements tea.Model. It fills the whole terminal and can be called in
// any state, at any size: a panic while drawing is turned into a message
// instead of leaving the terminal in raw mode.
func (s Shell) View() (out string) {
	defer func() {
		if r := recover(); r != nil {
			out = fmt.Sprintf("smith: the interface could not be drawn (%v). Press Ctrl+Q to leave.", r)
		}
	}()
	if s.w < 1 || s.h < 1 {
		return ""
	}
	return strings.Join(s.screen(), "\n")
}

// screen is the whole terminal as h lines of w cells.
func (s Shell) screen() []string {
	w, h := s.w, s.h
	rows := make([]string, 0, h)
	rows = append(rows, s.bar.Row(s.th, s.header, w))
	if h > 2 {
		rows = append(rows, s.body()...)
	}
	if h > 1 {
		rows = append(rows, statusbar.Render(s.th, s.statusInfo(), w))
	}
	rows = s.overlays(rows)
	for i := range rows {
		rows[i] = pane.Fit(rows[i], w)
	}
	return rows[:min(len(rows), h)]
}

// statusInfo is what the status bar shows: the exact command running or last
// run, its result, the branch and the version.
func (s Shell) statusInfo() statusbar.Info {
	info := statusbar.Info{Branch: s.branch, Version: s.version, Hint: s.hint, Tick: s.tick}
	if tip := s.bar.Tooltip(); tip != "" {
		info.Hint = tip
	}
	if tip := s.popup.Tooltip(); tip != "" {
		info.Hint = tip
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

// body draws the navigator, the main view and the inspector side by side.
func (s Shell) body() []string {
	h := s.lay.Main.H
	var nav, insp []string
	if s.lay.HasNavigator() {
		r := s.lay.Navigator
		tree := s.tree.SetHeight(max(r.H-2, 1))
		hover := -1
		if in := s.navigatorInner(); in.Contains(s.mouseX, s.mouseY) {
			hover = tree.RowAt(s.mouseY - in.Y)
		}
		lines := tree.View(s.th, max(r.W-2, 0), hover, s.focus == layout.Navigator, s.treeStatus)
		if !s.inWorkspace {
			lines = []string{s.th.Muted(" No workspace here."), s.th.Muted(" File > New workspace…")}
		}
		nav = pane.Frame(s.th, "Workspace", lines, r.W, r.H, s.focus == layout.Navigator)
	}
	if s.lay.HasInspector() {
		r := s.lay.Inspector
		lines, _ := s.inspectorPanel().View(s.th, max(r.W-2, 0), s.inspectorHover())
		insp = pane.Frame(s.th, "Inspector", lines, r.W, r.H, s.focus == layout.Inspector)
	}
	main := s.mainColumn()
	out := make([]string, 0, h)
	for y := 0; y < h; y++ {
		var b strings.Builder
		if nav != nil {
			b.WriteString(nav[y])
		}
		b.WriteString(main[y])
		if insp != nil {
			b.WriteString(insp[y])
		}
		out = append(out, b.String())
	}
	return out
}

// inspectorHover is the id of the inspector action under the pointer.
func (s Shell) inspectorHover() string {
	in := s.inspectorInner()
	if !in.Contains(s.mouseX, s.mouseY) {
		if s.focus == layout.Inspector {
			if a := s.inspectorActions(); s.inspCursor < len(a) {
				return a[s.inspCursor].ID
			}
		}
		return ""
	}
	_, hits := s.inspectorPanel().View(s.th, in.W, "")
	return hits[s.mouseY-in.Y]
}

// mainColumn is the tab row and, under it, the framed body of the active tab.
func (s Shell) mainColumn() []string {
	r := s.lay.Main
	rows := []string{s.tabRow(r.W)}
	in := s.mainInner()
	c := s.mainContent(in.W, in.H)
	lines := s.renderContent(c, in)
	rows = append(rows, pane.Frame(s.th, s.mainTitle(), lines, r.W, r.H-1, s.focus == layout.Main)...)
	return rows
}

// renderContent draws the button rows (with hover) over the content's blank
// lines.
func (s Shell) renderContent(c content, in layout.Rect) []string {
	lines := append([]string(nil), c.lines...)
	for i, bar := range c.bars {
		hover := ""
		if in.Contains(s.mouseX, s.mouseY) && s.mouseY-in.Y == i {
			if id, ok := bar.At(s.th, in.W, s.mouseX-in.X); ok {
				hover = id
			}
		}
		if i < len(lines) {
			lines[i] = bar.Render(s.th, s.tick, hover, in.W)
		}
	}
	return lines
}

// mainTitle is the heading of the framed body.
func (s Shell) mainTitle() string {
	switch s.tab {
	case commands.TabProjects:
		return fmt.Sprintf("Projects (%d)", len(s.tree.Projects()))
	case commands.TabRun:
		if s.run != nil {
			return s.run.CommandLine()
		}
	}
	return tabTitles[s.tab]
}

// tabSpan is where a tab's label sits in the tab row.
type tabSpan struct {
	id     string
	x0, x1 int
}

func (s Shell) tabLabel(id string) string {
	label := " " + tabTitles[id]
	switch id {
	case commands.TabRun:
		if s.running() {
			label += " " + s.th.Info(s.th.Spin(s.tick))
		} else if s.run != nil {
			if n := s.run.Counts().Failed; n > 0 {
				label += " " + s.th.Failure(fmt.Sprintf("%s%d", s.th.Icons.Failed, n))
			} else if s.run.OK() {
				label += " " + s.th.Success(s.th.Icons.Passed)
			}
		}
	case commands.TabDoctor:
		if s.doctor.report != nil && s.doctor.report.Failed > 0 {
			label += " " + s.th.Failure(fmt.Sprintf("%s%d", s.th.Icons.Failed, s.doctor.report.Failed))
		}
	}
	return label + " "
}

func (s Shell) tabSpans() []tabSpan {
	var spans []tabSpan
	x := 1
	for _, id := range s.tabs {
		w := pane.Width(s.tabLabel(id))
		spans = append(spans, tabSpan{id: id, x0: s.lay.Main.X + x, x1: s.lay.Main.X + x + w})
		x += w + 1
	}
	return spans
}

func (s Shell) tabRow(w int) string {
	var b strings.Builder
	b.WriteString(" ")
	for _, id := range s.tabs {
		label := s.tabLabel(id)
		switch {
		case id == s.tab:
			label = s.th.Selected(pane.Strip(label))
		case s.hoveredTab() == id:
			label = s.th.Hover(label)
		default:
			label = s.th.Muted(label)
		}
		b.WriteString(label + " ")
	}
	return pane.Fit(b.String(), w)
}

// hoveredTab is the tab under the pointer, or "".
func (s Shell) hoveredTab() string {
	if s.mouseY != s.lay.Main.Y {
		return ""
	}
	for _, sp := range s.tabSpans() {
		if s.mouseX >= sp.x0 && s.mouseX < sp.x1 {
			return sp.id
		}
	}
	return ""
}

// overlays composes the menus, popups, dialogs and toast over the rendered
// screen by cell, so nothing underneath is pushed down.
func (s Shell) overlays(rows []string) []string {
	if lines, x, y, ok := s.bar.Box(s.th); ok {
		rows = pane.Overlay(rows, lines, x, y)
	}
	if s.palOpen {
		box := s.paletteBox()
		lines := pane.Frame(s.th, "Command palette", s.pal.Lines(s.th, box.W-2, box.H-2), box.W, box.H, true)
		rows = pane.Overlay(rows, lines, box.X, box.Y)
	}
	if lines, x, y, ok := s.popup.Box(s.th); ok {
		rows = pane.Overlay(rows, lines, x, y)
	}
	if s.help != nil {
		r := s.help.Rect()
		rows = pane.Overlay(rows, s.help.Lines(s.th), r.X, r.Y)
	}
	if s.confirm != nil {
		r := s.confirm.Rect()
		rows = pane.Overlay(rows, s.confirm.Lines(s.th), r.X, r.Y)
	}
	if lines := s.toast.Lines(s.th, s.w-2); lines != nil {
		rows = pane.Overlay(rows, lines, max(s.w-pane.Width(lines[0])-1, 0), max(s.h-1-len(lines), 1))
	}
	return rows
}
