package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/tui/footer"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pointer"
	"github.com/monosmith-app/monosmith/apps/smith/tui/tabs"
)

// wheelStep is how many lines one notch of the wheel scrolls.
const wheelStep = 3

// paletteBox is where the command palette is drawn.
func (s Shell) paletteBox() layout.Rect {
	// Inside the content, whatever the size: never over the tabs or the footer.
	in := s.mainInner()
	w, h := min(max(s.w-8, 10), 100), min(max(in.H, 5), 28)
	return layout.Rect{X: (s.w - w) / 2, Y: in.Y + max((in.H-h)/3, 0), W: w, H: h}
}

// hoverDelay is the least time between two hover looks drawn while the pointer
// moves with no button held: about 40 frames a second.
var hoverDelay = 25 * time.Millisecond

// hoverMsg draws the hover look of the last pointer motion that waited for it.
type hoverMsg struct{}

// deferHover keeps a pointer motion with no button held for the next hover
// frame, so a pointer that sweeps the screen draws it once a frame and not once
// a cell. It reports whether it kept the motion, with the command that brings
// the frame (nil when one is on its way). Any other mouse event forgets a kept
// motion, so a late frame never moves the pointer back.
func (s *Shell) deferHover(e tea.MouseEvent) (tea.Cmd, bool) {
	if e.Action != tea.MouseActionMotion || e.Button != tea.MouseButtonNone ||
		s.pressed != "" || s.selection.Active || s.split.Dragging != "" {
		s.hoverAt = nil
		return nil, false
	}
	s.hoverAt = &e
	if s.hoverWaits {
		return nil, true
	}
	s.hoverWaits = true
	return tea.Tick(hoverDelay, func(time.Time) tea.Msg { return hoverMsg{} }), true
}

// applyHover handles the motion that waited for the hover frame.
func (s *Shell) applyHover() tea.Cmd {
	s.hoverWaits = false
	if s.hoverAt == nil {
		return nil
	}
	e := *s.hoverAt
	s.hoverAt = nil
	return s.mouse(tea.MouseMsg(e))
}

func (s *Shell) mouse(msg tea.MouseMsg) tea.Cmd {
	e := tea.MouseEvent(msg)
	if s.modalOpen() {
		return s.modalMouse(e)
	}
	s.mouseX, s.mouseY, s.mouseShift = e.X, e.Y, e.Shift
	if s.splitMouse(e) || s.selectMouse(e) {
		return nil
	}
	switch {
	case e.Action == tea.MouseActionMotion:
		if e.Button != tea.MouseButtonLeft {
			// A motion with no button held means the terminal lost the release.
			s.pressed = ""
		}
		s.hover(e.X, e.Y)
	case e.Action == tea.MouseActionRelease && s.pressed != "":
		return s.releaseButton(e.X, e.Y)
	case e.Action != tea.MouseActionPress:
		// Releases need no handling: everything happens on the press.
	case e.Button == tea.MouseButtonWheelUp:
		s.selection = s.selection.Clear()
		return s.wheel(e.X, e.Y, -1)
	case e.Button == tea.MouseButtonWheelDown:
		s.selection = s.selection.Clear()
		return s.wheel(e.X, e.Y, 1)
	case e.Button == tea.MouseButtonLeft:
		if s.pressButton(e.X, e.Y) {
			return nil
		}
		return s.pressAndSelect(e.X, e.Y)
	}
	return nil
}

// modalMouse handles a mouse event while a modal is open. Only the modal sees
// it: nothing behind it gets a hover, a press, a release, a wheel step, a drag
// or a click, and the pointer position the screen draws its hover looks from
// stays where it was. A press outside the modal does what the modal does with
// it (the palette closes; the other dialogs ignore it or follow their own rule).
func (s *Shell) modalMouse(e tea.MouseEvent) tea.Cmd {
	s.mouseShift = e.Shift
	// A press or a drag that began before the modal opened ends here.
	s.pressed = ""
	if s.modalSelect(e) {
		return nil
	}
	if e.Action == tea.MouseActionPress {
		// The highlight of a copied selection stays until a press or the wheel.
		s.selection = s.selection.Clear()
	}
	s.letGoSplit()
	switch {
	case e.Action == tea.MouseActionMotion:
		s.hover(e.X, e.Y)
	case e.Action != tea.MouseActionPress:
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
	case s.compose != nil:
		m := s.compose.Hover(x, y)
		s.compose = &m
	case s.cache != nil:
		m := s.cache.Hover(s.th, x, y)
		s.cache = &m
	case s.tools != nil:
		m := s.tools.Hover(s.th, x, y)
		s.tools = &m
	case s.result != nil:
		m := s.result.Hover(s.th, x, y)
		s.result = &m
	case s.adopt != nil:
		m := s.adopt.Hover(x, y)
		s.adopt = &m
	case s.fix != nil:
		w := s.fix.Hover(x, y)
		s.fix = &w
	case s.pkg != nil:
		p := s.pkg.Hover(x, y)
		s.pkg = &p
	case s.relpick != nil:
		m := s.relpick.Hover(x, y)
		s.relpick = &m
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
	case s.compose != nil:
		return s.composeClick(x, y)
	case s.cache != nil:
		return s.cacheClick(x, y)
	case s.tools != nil:
		m, out := s.tools.Click(s.th, x, y)
		s.tools = &m
		return s.afterTools(out)
	case s.result != nil:
		return s.resultClick(x, y)
	case s.adopt != nil:
		m, out := s.adopt.Click(x, y)
		s.adopt = &m
		return s.afterAdopt(out)
	case s.fix != nil:
		w, out := s.synced(*s.fix).Click(x, y)
		s.fix = &w
		return s.afterFix(out)
	case s.pkg != nil:
		p, out := s.pkg.Click(x, y)
		s.pkg = &p
		return s.afterPackages(p, out)
	case s.relpick != nil:
		m, out := s.relpick.Click(x, y)
		s.relpick = &m
		return s.afterReleasePick(m, out)
	case s.confirm != nil:
		c, out := s.confirm.Click(x, y)
		s.confirm = &c
		return s.afterConfirm(c, out)
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
			s.sideFocus = false
			return s.contentClick(x-in.X, y-in.Y, double)
		}
	case layout.Detail:
		in := s.detailInner()
		if in.Contains(x, y) {
			// The click lands on the rows as they were drawn: the cursor mark of a detail
			// that has the keyboard changes how its notes wrap. The keyboard moves after.
			cmd := s.clickHit(s.sideContent(in.W, in.H-s.toastRows(in.H)), x-in.X, y-in.Y, double)
			s.sideFocus = s.beside()
			return cmd
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
	if s.tab == commands.TabSettings {
		return s.settingsClick(in.W, in.H, rx, ry, double)
	}
	return s.clickHit(s.mainContent(in.W, in.H-s.toastRows(in.H)), rx, ry, double)
}

// clickHit handles a click at (rx, ry) in a body c that was drawn: on a project,
// a task, a recent run, a part of the run header, a finding, a dependency, an
// option or a button.
func (s *Shell) clickHit(c content, rx, ry int, double bool) tea.Cmd {
	h, ok := c.hits[ry]
	if !ok {
		return nil
	}
	if len(h.spans) > 0 {
		sp, ok := h.at(rx)
		if !ok {
			return nil
		}
		h.id, h.index = sp.id, sp.index
	}
	switch h.kind {
	case "project":
		// A project clicked shows its own detail: a run that was open goes on in
		// the background and stays in the list of recent runs.
		s.runOpen, s.recentFocus = false, false
		s.list = s.list.SelectProject(h.id)
		s.detailCursor = 0
		if double {
			return s.paneKey(tea.KeyMsg{Type: tea.KeyEnter})
		}
	case "task":
		s.runSel = h.id
		s.keepRunVisible()
	case "recent":
		s.recentAt = h.index
		if s.tab == commands.TabHome {
			// A run on Home opens at one click, in the run panel of Projects.
			return s.openRecent()
		}
		s.recentFocus = true
		if double {
			return s.openRecent()
		}
	case "nav":
		switch s.navAction(rx) {
		case "older":
			s.stepRun(-1)
		case "newer":
			s.stepRun(1)
		}
	case "finding":
		s.doctor = s.doctor.SelectRow(h.index)
	case "release":
		s.release = s.release.SelectRow(h.index)
	case "dep":
		inBox := rx >= depsview.BoxFrom && rx < depsview.BoxTo
		switch {
		case h.index == depsview.HeadingRow:
			if inBox {
				s.deps = s.deps.TickAll()
			}
		case inBox:
			s.deps = s.deps.TickRow(h.index, s.mouseShift)
		default:
			s.deps = s.deps.ChooseRow(h.index, s.mouseShift)
		}
	case "option":
		// One click picks in a single-choice list. In a list of tick boxes a click
		// ticks; in the path finder it selects. There a double click goes on.
		if s.guide.SingleChoice() {
			s.guide = s.guide.SelectRow(h.index)
			return s.paneKey(tea.KeyMsg{Type: tea.KeyEnter})
		}
		if double {
			return s.paneKey(tea.KeyMsg{Type: tea.KeyEnter})
		}
		s.guide = s.guide.SelectRow(h.index)
	case "choose":
		// A button above the rule of a split detail chooses; the run button under it runs.
		for i, a := range s.detailActions() {
			if a.ID == h.id {
				s.chooseDetail(i)
			}
		}
	case "action":
		if h.id == detail.RunID {
			s.detailOnRun = true
			return s.runChosen()
		}
		if s.detailFull() {
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
	case s.compose != nil:
		c := s.compose.Wheel(dir * wheelStep)
		s.compose = &c
		return nil
	case s.cache != nil:
		m := s.cache.Wheel(dir * wheelStep)
		s.cache = &m
		return nil
	case s.tools != nil, s.result != nil:
		return nil
	case s.adopt != nil:
		m := s.adopt.Wheel(dir * wheelStep)
		s.adopt = &m
		return nil
	case s.fix != nil:
		w := s.synced(*s.fix).Wheel(dir * wheelStep)
		s.fix = &w
		return nil
	case s.pkg != nil, s.relpick != nil, s.confirm != nil:
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
	case layout.Detail:
		s.wheelSide(y, dir)
	case layout.Main:
		if s.detailFull() {
			s.wheelSide(y, dir)
			return nil
		}
		switch s.tab {
		case commands.TabHome:
			s.scrollHome(dir * wheelStep)
		case commands.TabDeps:
			s.deps = s.deps.Wheel(dir * wheelStep)
		case commands.TabProjects:
			if s.projScroll == nil {
				s.projScroll = &pane.Scroll{}
			}
			s.projScroll.By(dir * wheelStep)
		case commands.TabSettings:
			s.settings = s.settings.Wheel(dir)
		case commands.TabDoctor:
			if s.doctorScroll == nil {
				s.doctorScroll = &pane.Scroll{}
			}
			s.doctorScroll.By(dir * wheelStep)
		case commands.TabGuide:
			// The wheel scrolls the list of the question; the selection stays.
			s.guide = s.guide.Wheel(dir * wheelStep)
		case commands.TabRelease:
			key := tea.KeyMsg{Type: tea.KeyDown}
			if dir < 0 {
				key = tea.KeyMsg{Type: tea.KeyUp}
			}
			return s.paneKey(key)
		}
	}
	return nil
}

// wheelSide scrolls what the right panel shows: the log of the run (or its task
// list, above the log) while a run is open, else the detail.
func (s *Shell) wheelSide(y, dir int) {
	if !s.runShown() {
		s.scrollDetail(dir * wheelStep)
		return
	}
	r := s.detailRect()
	g := s.runGeom(r.W, r.H)
	if y-r.Y >= g.logTop {
		s.log = s.log.Scroll(dir*wheelStep, s.logTotal(), g.logRows)
		return
	}
	s.runTop = max(min(s.runTop+dir, len(s.runIDs())-g.listRows), 0)
}
