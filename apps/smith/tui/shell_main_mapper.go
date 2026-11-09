package tui

import (
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// hit says what a click on a line of the main view means.
type hit struct {
	// kind is "project", "task", "recent", "nav", "finding", "release", "dep", "option" or "action".
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
	case commands.TabDoctor:
		if s.detailFull() {
			return s.detailContent(w, h)
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
		release := s.release
		release.Waited, release.Projects = s.waited(), len(s.list.Projects())
		release.Button, release.FirstButton = s.releaseViewButton("rel.release", "tui.release_confirm.run"), s.releaseViewButton("rel.first", "tui.release_empty.first_button")
		lines, rowOf := release.View(s.th, w)
		for row, i := range rowOf {
			switch i {
			case releaseview.RowRelease:
				c.hits[len(c.lines)+row] = hit{kind: "action", id: "rel.release"}
			case releaseview.RowFirst:
				c.hits[len(c.lines)+row] = hit{kind: "action", id: "rel.first"}
			case releaseview.RowWhy:
				c.hits[len(c.lines)+row] = hit{kind: "action", id: "rel.why"}
			default:
				c.hits[len(c.lines)+row] = hit{kind: "release", index: i}
			}
		}
		c.lines = append(c.lines, lines...)
		c.lines = append(c.lines, s.previewLines(w, h-len(c.lines))...)
	case commands.TabDeps:
		deps := s.deps
		deps.Waited, deps.Projects, deps.Tick = s.waited(), len(s.list.Projects()), s.tick
		deps.Beside = s.lay().HasDetail()
		deps.Under, deps.Pressed = depsview.FilterOf(s.actionAt(s.mouseX, s.mouseY)), depsview.FilterOf(s.pressed)
		lines, rowOf := deps.View(s.th, w, max(h-len(c.lines), 1))
		for row, i := range rowOf {
			c.hits[len(c.lines)+row] = hit{kind: "dep", index: i}
		}
		// The chips at the top answer a click with their own key.
		for line, spans := range deps.HeadlineSpans(s.th, w) {
			chips := hit{kind: "action"}
			for _, sp := range spans {
				chips.spans = append(chips.spans, hitSpan{from: sp.From, to: sp.To, id: sp.Filter.CommandID()})
			}
			c.hits[len(c.lines)+line] = chips
		}
		c.lines = append(c.lines, lines...)
	case commands.TabSettings:
		in := s.mainInner()
		c.lines = s.settings.WithSideDetail(s.lay().HasDetail()).Pulse(s.tick, s.settingsWaited()).View(s.th, w, h, s.mouseX-in.X, s.mouseY-in.Y)
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

// projectsContent is the table of projects, with the list of recent runs under
// it; the keys that run tasks on the selection are in the footer, and its detail
// (or the run that is open) is beside it or opens over it.
func (s Shell) projectsContent(w, h int) content {
	c := content{hits: map[int]hit{}}
	if !s.loadedProjects {
		return withLines(c, s.th.Muted(words.Say("tui.main.reading_the_workspace")))
	}
	if s.projectsErr != nil {
		return withLines(c, s.th.Failure(s.projectsErr.Error()))
	}
	projects := s.list.Projects()
	under := s.actionAt(s.mouseX, s.mouseY)
	if len(projects) == 0 {
		c.hits[0], c.hits[1] = hit{kind: "action", id: "file.add"}, hit{kind: "action", id: "file.compose"}
		return withLines(c, s.newProjectButton(under), button.Render(s.th, words.Say("tui.main.compose_a_stack"), commands.ShortcutOf("file.compose"), s.buttonState("file.compose", under)),
			"", s.th.Muted(words.Say("tui.main.no_projects_yet_add_one_or")))
	}
	affected := 0
	for _, p := range projects {
		if p.Affected {
			affected++
		}
	}
	// The summary and the legend share a line, and wrap onto the next when the pane is narrow.
	noun := "projects"
	if len(projects) == 1 {
		noun = "project"
	}
	summary := words.Say("tui.main.affected_by_the_changes", "len", len(projects), "noun", noun, "affected", affected, "dot", s.th.Icons.Dot)
	c.lines = append(c.lines, pane.Wrap(s.th.Muted(summary+"   "+words.Say("tui.main.dependencies_dependants_legend")), w, 0)...)
	// Every action on the workspace sits on one line, or on as few as the width needs.
	for _, row := range s.workspaceButtons(w, under) {
		c.hits[len(c.lines)] = row.hit
		c.lines = append(c.lines, row.line)
	}
	c.lines = append(c.lines, "")

	cols := s.projectColumns(w)
	head := "   " + pane.Cell(words.Say("tui.main.project"), cols.Name+1)
	if cols.Kind {
		head += pane.Cell(words.Say("tui.main.kind"), projectlist.KindW)
	}
	if cols.Lang {
		head += pane.Cell(words.Say("tui.main.language"), projectlist.LangW)
	}
	if cols.Stack {
		head += pane.Cell(words.Say("tui.main.stack"), projectlist.StackW)
	}
	if cols.Version {
		head += pane.Cell(words.Say("tui.main.version"), projectlist.VersionW)
	}
	head += pane.Cell("↑ ↓", projectlist.DepsW)
	c.lines = append(c.lines, s.th.Muted(pane.Fit(head, w)))

	// The block of recent runs sits on the bottom rows of the pane; the list of
	// projects keeps at least minProjectRows rows above it.
	const minProjectRows = 3
	recent := s.recentSection(w, h-len(c.lines)-minProjectRows)
	avail := max(h-len(c.lines)-len(recent.lines), 1)
	sel := s.list.Cursor()
	top := s.projScroll.Show(sel, avail, len(projects))
	first := len(c.lines)
	for i := top; i < len(projects) && i < top+avail; i++ {
		p := projects[i]
		dot := s.th.Muted(s.th.Icons.Queued)
		if p.Affected {
			dot = s.th.Warning(s.th.Icons.Dot)
		}
		line := " " + dot + " " + pane.Cell(p.Name+s.treeStatus(p.Name, ""), cols.Name+1)
		if cols.Kind {
			line += pane.Cell(p.Kind, projectlist.KindW)
		}
		if cols.Lang {
			line += pane.Cell(p.Language, projectlist.LangW)
		}
		if cols.Stack {
			line += pane.Cell(p.Stack, projectlist.StackW)
		}
		if cols.Version {
			line += pane.Cell(s.versionOf(p, projectlist.VersionW-1), projectlist.VersionW)
		}
		line += pane.Cell(words.Say("tui.main.text", "len", len(p.Deps), "len2", len(p.Dependents)), projectlist.DepsW)
		if p.Name == s.selectedProject() {
			line = s.th.Selected(pane.Fit(pane.Strip(line), w))
		}
		c.hits[len(c.lines)] = hit{kind: "project", id: p.Name, index: i}
		c.lines = append(c.lines, pane.Fit(line, w))
	}
	copy(c.lines[first:], pane.Scrollbar(s.th, c.lines[first:], w, top, len(projects)))
	// Blank rows keep the block on the bottom of the pane.
	for len(recent.lines) > 0 && len(c.lines) < h-len(recent.lines) {
		c.lines = append(c.lines, "")
	}
	for i, line := range recent.lines {
		if h, ok := recent.hits[i]; ok {
			c.hits[len(c.lines)] = h
		}
		c.lines = append(c.lines, line)
	}
	return c
}

// projectColumns is how the table of projects spends w cells: which of the
// Kind, Language and Stack columns it draws. The detail asks it too, so it
// leaves out only the facts the table shows.
func (s Shell) projectColumns(w int) projectlist.Columns {
	nameW, hasStack := 8, false
	for _, p := range s.list.Projects() {
		hasStack = hasStack || p.Stack != ""
		nameW = max(nameW, pane.Width(p.Name)+2) // the room for the last run's mark
	}
	return projectlist.ColumnsFor(w, nameW, hasStack, s.hasVersions())
}

// newProjectButton is the row that starts the Add wizard from the list of projects.
func (s Shell) newProjectButton(under string) string {
	return button.Render(s.th, words.Say("tui.main.new_project"), commands.ShortcutOf("file.add"), s.buttonState("file.add", under))
}

func withLines(c content, lines ...string) content {
	c.lines = append(c.lines, lines...)
	return c
}

func pad(s string, w int) string { return pane.Fit(s, w) }

// treeStatus is the mark drawn after a project or task: how the last run did there.
func (s Shell) treeStatus(project, task string) string {
	if s.run == nil {
		return ""
	}
	if task != "" {
		if t := s.run.Task(project + ":" + task); t != nil {
			return " " + s.statusIcon(t.Status)
		}
		return ""
	}
	best, found := runview.Queued, false
	prefix := project + ":"
	for _, t := range s.run.Tasks {
		if !strings.HasPrefix(t.ID, prefix) {
			continue
		}
		if !found || statusRank[t.Status] < statusRank[best] {
			best, found = t.Status, true
		}
	}
	if !found {
		return ""
	}
	return " " + s.statusIcon(best)
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

func yesNoMark(s Shell, b bool) string {
	if b {
		return words.Say("tui.main.yes", "dot", s.th.Warning(s.th.Icons.Dot))
	}
	return "no"
}

// shortDuration is a duration for a list: milliseconds under a second, then tenths of a second.
func shortDuration(d time.Duration) string {
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	return d.Round(100 * time.Millisecond).String()
}
