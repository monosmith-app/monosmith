package tui

import (
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/releaseview"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/runview"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// hit says what a click on a line of the main view means.
type hit struct {
	// kind is "project", "task", "recent", "nav", "finding", "release", "dep", "option", "homeitem" or "action".
	kind  string
	id    string
	index int
	// spans, when set, split the line into the parts that answer a click: a
	// click outside every span does nothing, and in one it takes that span's
	// id and index.
	spans []hitSpan
}

// hitSpan is a part of a line, the cells [from, to), that answers a click with
// its own id and index.
type hitSpan struct {
	from, to int
	id       string
	index    int
}

// at is the span under column x of the line.
func (h hit) at(x int) (hitSpan, bool) {
	for _, sp := range h.spans {
		if x >= sp.from && x < sp.to {
			return sp, true
		}
	}
	return hitSpan{}, false
}

// content is the main view's body: its lines, and what each line stands for.
type content struct {
	lines []string
	hits  map[int]hit
}

// mainInner is where the pane's body is drawn: inside the border.
func (s Shell) mainInner() layout.Rect { return s.lay().Main.Inner() }

// detailInner is where the detail beside the list is drawn.
func (s Shell) detailInner() layout.Rect { return s.lay().DetailInner() }

// mainContent builds the body of the active tab for an area of w by h cells.
func (s Shell) mainContent(w, h int) content {
	c := content{hits: map[int]hit{}}
	switch s.tab {
	case commands.TabProjects:
		if s.detailFull() {
			return s.sideContent(w, h)
		}
		return s.projectsContent(w, h)
	case commands.TabHome:
		return s.homeContent(w, h)
	case commands.TabRuns:
		return s.runsContent(w, h)
	case commands.TabDoctor:
		if s.detailFull() {
			return s.sideContent(w, h)
		}
		// Beside the list the detail tells a check in full; alone, the list expands it.
		doctor := s.doctor
		doctor.Waited, doctor.Covered = s.waited(), len(s.list.Projects())
		under := s.actionAt(s.mouseX, s.mouseY)
		doctor.Under, doctor.Pressed = doctorview.FilterOf(under), doctorview.FilterOf(s.pressed)
		doctor.UnderReload, doctor.PressedReload = under == doctorview.ReloadID, s.pressed == doctorview.ReloadID
		lines, rowOf := doctor.View(s.th, w, !s.lay().HasDetail())
		sel := 0
		for row, i := range rowOf {
			if doctor.IsCursorRow(i) {
				sel = row
			}
		}
		// The headline (the filter buttons and its blank line) stays at the top;
		// only the checks below it scroll.
		fixed := min(doctor.HeadHeight(s.th, w), max(h-1, 0))
		body := max(h-fixed, 1)
		total := len(lines) - fixed
		top := s.doctorScroll.Show(max(sel-fixed, 0), body, total)
		for row, i := range rowOf {
			if r := row - fixed - top; r >= 0 && r < body {
				c.hits[fixed+r] = hit{kind: "finding", index: i}
			}
		}
		// The headline is the filter buttons: each answers a click with its own key.
		for line, spans := range doctor.HeadlineSpans(s.th, w) {
			if line >= fixed {
				continue
			}
			filters := hit{kind: "action"}
			for _, sp := range spans {
				if sp.ID != "" {
					filters.spans = append(filters.spans, hitSpan{from: sp.From, to: sp.To, id: sp.ID})
				}
			}
			c.hits[line] = filters
		}
		scrolled := lines[fixed+top : min(fixed+top+body, len(lines))]
		c.lines = append(c.lines, lines[:fixed]...)
		c.lines = append(c.lines, pane.Scrollbar(s.th, scrolled, w, top, total)...)
	case commands.TabRelease:
		if s.detailFull() {
			// On a narrow screen Enter opened the selected package in the frame.
			return s.sideContent(w, h)
		}
		release := s.release
		release.Waited, release.Projects, release.Tick = s.waited(), len(s.list.Projects()), s.tick
		release.Button, release.FirstButton = "", s.releaseViewButton("rel.first", "tui.release_empty.first_button")
		lines, rowOf := release.View(s.th, w)
		for row, i := range rowOf {
			switch i {
			case releaseview.RowFirst:
				c.hits[row] = hit{kind: "action", id: "rel.first"}
			case releaseview.RowWhy:
				c.hits[row] = hit{kind: "action", id: "rel.why"}
			default:
				c.hits[row] = hit{kind: "release", index: i}
			}
		}
		c.lines = lines
		for _, row := range s.releaseRows(w) {
			if row.hit.kind != "" {
				c.hits[len(c.lines)] = row.hit
			}
			c.lines = append(c.lines, row.line)
		}
		c.lines = append(c.lines, s.previewLines(w, h-len(c.lines))...)
	case commands.TabDeps:
		if s.detailFull() {
			// On a narrow screen Enter opened the selected package in the frame.
			return s.sideContent(w, h)
		}
		deps := s.deps
		deps.Waited, deps.Projects, deps.Tick, deps.Links = s.waited(), len(s.list.Projects()), s.tick, s.hyperlinks
		deps.Beside = s.lay().HasDetail()
		deps.Under, deps.Pressed = depsview.FilterOf(s.actionAt(s.mouseX, s.mouseY)), depsview.FilterOf(s.pressed)
		box := s.depsBox(w, s.actionAt(s.mouseX, s.mouseY))
		listH := h
		if len(box) > 0 {
			listH = max(h-len(box)-1, 3)
		}
		lines, rowOf := deps.View(s.th, w, listH)
		for row, i := range rowOf {
			c.hits[row] = hit{kind: "dep", index: i}
		}
		// The chips at the top answer a click with their own key.
		for line, spans := range deps.HeadlineSpans(s.th, w) {
			if len(spans) == 0 {
				continue
			}
			chips := hit{kind: "action"}
			for _, sp := range spans {
				chips.spans = append(chips.spans, hitSpan{from: sp.From, to: sp.To, id: sp.Filter.CommandID()})
			}
			c.hits[line] = chips
		}
		c.lines = lines
		if len(box) > 0 {
			// The box of the actions stands at the foot of the pane, under the list.
			for len(c.lines) < h-len(box) {
				c.lines = append(c.lines, "")
			}
			for _, row := range box {
				if row.hit.kind != "" {
					c.hits[len(c.lines)] = row.hit
				}
				c.lines = append(c.lines, row.line)
			}
		}
	case commands.TabSettings:
		// A wide screen lists the sections in this frame and the open section in
		// the frame beside it; a narrow one shows one of them at a time.
		in := s.mainInner()
		tab := s.settings.Pulse(s.tick, s.settingsWaited())
		if s.lay().HasDetail() {
			c.lines = tab.SectionsPane(s.th, w, h, s.mouseX-in.X, s.mouseY-in.Y)
		} else {
			c.lines = tab.Narrow(s.th, w, h, s.mouseX-in.X, s.mouseY-in.Y)
		}
	case commands.TabGuide:
		broken := s.brokenConfigBanner(w)
		banner := append(broken, s.adoptBanner(w)...)
		for i := len(broken); i < len(banner); i++ {
			c.hits[i] = hit{kind: "action", id: "file.adopt"}
		}
		lines, optionAt := s.guide.Window(s.th, w, max(h-len(banner), 1))
		for row, i := range optionAt {
			c.hits[len(banner)+row] = hit{kind: "option", index: i}
		}
		c.lines = append(banner, lines...)
	}
	return c
}

// newProjectButton is the row that starts the Add wizard from the list of projects.
func (s Shell) newProjectButton(under string) string {
	return button.Render(s.th, words.Say("tui.main.new_project"), commands.ShortcutOf("file.add"), s.buttonState("file.add", under))
}

func withLines(c content, lines ...string) content {
	c.lines = append(c.lines, lines...)
	return c
}

// statusRank orders the states by how much they matter to a project's mark.
var statusRank = map[runview.Status]int{runview.Failed: 0, runview.Running: 1, runview.Queued: 2, runview.Cancelled: 3, runview.Skipped: 3, runview.Cached: 4, runview.Passed: 5}

// statusIcon is the symbol of a task's state; the running one spins.
func (s Shell) statusIcon(st runview.Status) string {
	switch st {
	case runview.Running:
		return s.th.Info(s.th.Spin(s.tick))
	case runview.Passed:
		return s.th.Success(s.th.Icons.Passed)
	case runview.Cached:
		return s.th.Success(s.th.Icons.Cached)
	case runview.Failed:
		return s.th.Failure(s.th.Icons.Failed)
	case runview.Skipped:
		return s.th.Warning(s.th.Icons.Skipped)
	case runview.Cancelled:
		return s.th.Muted(s.th.Icons.Cancelled)
	}
	return s.th.Muted(s.th.Icons.Queued)
}

// shortDuration is a duration for a list: milliseconds under a second, then tenths of a second.
func shortDuration(d time.Duration) string {
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	return d.Round(100 * time.Millisecond).String()
}
