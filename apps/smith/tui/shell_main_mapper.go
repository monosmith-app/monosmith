package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
)

// hit says what a click on a line of the main view means.
type hit struct {
	// kind is "project", "task", "finding", "release", "dep", "option" or "action".
	kind  string
	id    string
	index int
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
		if s.detailOpen {
			return s.detailContent(w, h)
		}
		return s.projectsContent(w, h)
	case commands.TabRun:
		return s.runContent(w, h)
	case commands.TabHome:
		return s.homeContent(w)
	case commands.TabDoctor:
		// Beside the list the detail tells a check in full; alone, the list expands it.
		lines, rowOf := s.doctor.View(s.th, w, !s.lay().HasDetail())
		for row, i := range rowOf {
			c.hits[len(c.lines)+row] = hit{kind: "finding", index: i}
		}
		c.lines = append(c.lines, lines...)
	case commands.TabRelease:
		lines, rowOf := s.release.View(s.th, w)
		for row, i := range rowOf {
			c.hits[len(c.lines)+row] = hit{kind: "release", index: i}
		}
		c.lines = append(c.lines, lines...)
		c.lines = append(c.lines, s.previewLines(w, h-len(c.lines))...)
	case commands.TabDeps:
		lines, rowOf := s.deps.View(s.th, w, max(h-len(c.lines), 1))
		for row, i := range rowOf {
			c.hits[len(c.lines)+row] = hit{kind: "dep", index: i}
		}
		c.lines = append(c.lines, lines...)
	case commands.TabGuide:
		lines, optionAt := s.guide.Body(s.th, w)
		for row, i := range optionAt {
			c.hits[row] = hit{kind: "option", index: i}
		}
		c.lines = lines
	}
	return c
}

// projectsContent is the table of projects; the keys that run tasks on the
// selection are in the footer, and its detail is beside it or opens over it.
func (s Shell) projectsContent(w, h int) content {
	c := content{hits: map[int]hit{}}
	if !s.loadedProjects {
		return withLines(c, s.th.Muted("Reading the workspace…"))
	}
	if s.projectsErr != nil {
		return withLines(c, s.th.Failure(s.projectsErr.Error()))
	}
	projects := s.list.Projects()
	if len(projects) == 0 {
		return withLines(c, s.th.Muted("No projects yet. Press "+commands.ShortcutOf("file.add")+" to add one."))
	}
	scope := "every project"
	if p := s.selectedProject(); p != "" {
		scope = p
	}
	affected := 0
	for _, p := range projects {
		if p.Affected {
			affected++
		}
	}
	// The explanations wrap instead of running off the pane.
	c.lines = append(c.lines, pane.Wrap(s.th.Muted("On: ")+s.th.Bold(scope), w, 0)...)
	c.lines = append(c.lines, pane.Wrap(s.th.Muted(fmt.Sprintf("%d projects · %d affected by the changes (%s)", len(projects), affected, s.th.Icons.Dot)), w, 0)...)
	c.lines = append(c.lines, pane.Wrap(s.th.Muted("↑ = projects it needs · ↓ = projects that use it"), w, 0)...)
	c.lines = append(c.lines, "")

	nameW, tasksW, hasStack := 8, 0, false
	for _, p := range projects {
		hasStack = hasStack || p.Stack != ""
		nameW = max(nameW, pane.Width(p.Name)+2) // the room for the last run's mark
		tasksW = max(tasksW, pane.Width(taskList(p)))
	}
	cols := columnsForProjects(w, nameW, tasksW, hasStack)
	head := "   " + pane.Cell("Project", cols.name+1)
	if cols.kind {
		head += pane.Cell("Kind", kindColW)
	}
	if cols.lang {
		head += pane.Cell("Language", langColW)
	}
	if cols.stack {
		head += pane.Cell("Stack", stackColW)
	}
	head += pane.Cell("↑ ↓", depsColW) + "Tasks"
	c.lines = append(c.lines, s.th.Muted(pane.Fit(head, w)))

	avail := max(h-len(c.lines), 1)
	sel := s.list.Cursor()
	top := min(max(sel-avail/2, 0), max(len(projects)-avail, 0))
	for i := top; i < len(projects) && i < top+avail; i++ {
		p := projects[i]
		dot := s.th.Muted(s.th.Icons.Queued)
		if p.Affected {
			dot = s.th.Warning(s.th.Icons.Dot)
		}
		line := " " + dot + " " + pane.Cell(p.Name+s.treeStatus(p.Name, ""), cols.name+1)
		if cols.kind {
			line += pane.Cell(p.Kind, kindColW)
		}
		if cols.lang {
			line += pane.Cell(p.Language, langColW)
		}
		if cols.stack {
			line += pane.Cell(p.Stack, stackColW)
		}
		line += pane.Cell(fmt.Sprintf("%d %d", len(p.Deps), len(p.Dependents)), depsColW)
		line += ansi.Truncate(taskList(p), cols.tasks, "…")
		if p.Name == s.selectedProject() {
			line = s.th.Selected(pane.Fit(pane.Strip(line), w))
		}
		c.hits[len(c.lines)] = hit{kind: "project", id: p.Name, index: i}
		c.lines = append(c.lines, pane.Fit(line, w))
	}
	return c
}

// taskList is a project's task names on one line.
func taskList(p projectlist.Project) string {
	var names []string
	for _, t := range p.Tasks {
		names = append(names, t.Name)
	}
	return strings.Join(names, " ")
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
var statusRank = map[runview.Status]int{runview.Failed: 0, runview.Running: 1, runview.Queued: 2, runview.Skipped: 3, runview.Cached: 4, runview.Passed: 5}

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
	}
	return s.th.Muted(s.th.Icons.Queued)
}

// ---- the Run tab ----------------------------------------------------------

// runGeom is where the parts of the Run tab sit, in lines of its body.
type runGeom struct {
	cmdRow, progRow   int
	listTop, listRows int
	sepRow            int
	logTop, logRows   int
}

func (s Shell) runGeom(w, h int) runGeom {
	g := runGeom{}
	g.cmdRow, g.progRow, g.listTop = 0, 1, 2
	n := len(s.runIDs())
	avail := max(h-g.listTop-2, 1)
	g.listRows = min(n, max(min(avail, 2), avail*2/5))
	g.sepRow = g.listTop + g.listRows
	g.logTop = g.sepRow + 1
	g.logRows = max(h-g.logTop, 1)
	return g
}

// logHeight is the number of log lines visible in the Run tab.
func (s Shell) logHeight() int {
	in := s.mainInner()
	return s.runGeom(in.W, in.H).logRows
}

// keepRunVisible scrolls the task list so the selected task shows.
func (s *Shell) keepRunVisible() {
	in := s.mainInner()
	g := s.runGeom(in.W, in.H)
	at := 0
	for i, id := range s.runIDs() {
		if id == s.runSel {
			at = i
		}
	}
	if at < s.runTop {
		s.runTop = at
	}
	if at >= s.runTop+g.listRows {
		s.runTop = at - g.listRows + 1
	}
	s.runTop = max(min(s.runTop, len(s.runIDs())-g.listRows), 0)
}

func (s Shell) runContent(w, h int) content {
	c := content{hits: map[int]hit{}}
	if s.run == nil && len(s.runs) > 0 {
		c.lines = append(c.lines, "", s.th.Title("Runs"))
		c.lines = append(c.lines, s.inlineRunLines(w, h)...)
		c.lines = append(c.lines, "")
		c.lines = append(c.lines, pane.Wrap(s.th.Muted("The output of a run stays in your terminal's scrollback. "+commands.ShortcutOf("run.again")+" runs the last one again."), w, 0)...)
		return c
	}
	if s.run == nil {
		c.lines = append(c.lines, "", s.th.Muted("Nothing has run yet."))
		c.lines = append(c.lines, pane.Wrap(s.th.Muted("Pick a project in the list and press "+commands.ShortcutOf("run.build")+" to build it, or "+commands.ShortcutOf("run.test")+" to test."), w, 0)...)
		c.lines = append(c.lines, "")
		if len(s.history) > 0 {
			c.lines = append(c.lines, s.th.Title("Command log"))
			for i := len(s.history) - 1; i >= 0 && len(c.lines) < h; i-- {
				c.lines = append(c.lines, "  "+s.history[i])
			}
		}
		return c
	}
	g := s.runGeom(w, h)
	counts := s.run.Counts()
	now := s.now()
	took := shortDuration(s.run.RunTime(now))
	state := s.th.Info(s.th.Spin(s.tick) + " running")
	switch {
	case s.stopping && s.running():
		state = s.th.Warning(s.th.Spin(s.tick) + " stopping")
	case s.run.Cancelled:
		state = s.th.Warning(s.th.Icons.Skipped + " cancelled")
	case s.run.Finished && s.run.OK():
		state = s.th.Success(s.th.Icons.Passed + " passed")
	case s.run.Finished:
		state = s.th.Failure(s.th.Icons.Failed + " failed")
	}
	c.lines = append(c.lines, pane.Fit(" "+s.th.Faint("$")+" "+s.th.Bold(s.run.CommandLine())+"   "+state+"  "+s.th.Muted(took), w))
	if counts.Total > 0 {
		bar := progress.Draw(s.th, counts.Done, counts.Total, min(max(w/2, 20), 40), counts.Failed > 0)
		c.lines = append(c.lines, pane.Fit(" "+bar+"  "+s.countsText(counts), w))
	} else {
		c.lines = append(c.lines, "")
	}
	ids := s.runIDs()
	for i := s.runTop; i < len(ids) && i < s.runTop+g.listRows; i++ {
		c.hits[len(c.lines)] = hit{kind: "task", id: ids[i], index: i}
		c.lines = append(c.lines, s.taskLine(ids[i], w, now))
	}
	for len(c.lines) < g.sepRow {
		c.lines = append(c.lines, "")
	}
	title := "all output"
	if s.runSel != "" {
		title = s.runSel
	}
	follow := "following"
	if !s.log.Following() {
		follow = "paused"
	}
	sep := "── Log: " + title + " (" + follow + ") "
	c.lines = append(c.lines, s.th.Faint(pane.Fit(sep+strings.Repeat("─", max(w-pane.Width(sep), 0)), w)))
	c.lines = append(c.lines, s.log.Render(s.th, s.logLines(), w, g.logRows)...)
	return c
}

// countsText is "✔ 4  ⚡ 1  ✖ 1  ↷ 2  ● 2  ○ 3" without the zeros.
func (s Shell) countsText(c runview.Counts) string {
	var parts []string
	add := func(n int, st runview.Status) {
		if n > 0 {
			parts = append(parts, fmt.Sprintf("%s %d %s", s.statusIcon(st), n, st))
		}
	}
	add(c.Failed, runview.Failed)
	add(c.Running, runview.Running)
	add(c.Passed, runview.Passed)
	add(c.Cached, runview.Cached)
	add(c.Skipped, runview.Skipped)
	add(c.Queued, runview.Queued)
	return strings.Join(parts, "  ")
}

// taskLine is one row of the task list.
func (s Shell) taskLine(id string, w int, now time.Time) string {
	if id == "" {
		line := fmt.Sprintf(" %s All output  %s", s.th.Icons.Pointer, s.th.Muted(fmt.Sprintf("%d lines", len(s.run.Lines))))
		if s.runSel == "" {
			return s.th.Selected(pane.Fit(pane.Strip(line), w))
		}
		return pane.Fit(line, w)
	}
	t := s.run.Task(id)
	if t == nil {
		return ""
	}
	nameW := min(max(w/3, 12), 28)
	detail := t.Status.String()
	if d := s.run.Elapsed(t, now); d > 0 {
		detail += "  " + shortDuration(d)
	}
	if t.Reason != "" && t.Status != runview.Passed {
		detail += "  " + t.Reason
	}
	line := " " + s.statusIcon(t.Status) + " " + pad(t.ID, nameW) + " " + s.th.Muted(detail)
	if id == s.runSel {
		line = s.th.Selected(pane.Fit(pane.Strip(line), w))
	}
	return pane.Fit(line, w)
}

func yesNoMark(s Shell, b bool) string {
	if b {
		return s.th.Warning(s.th.Icons.Dot) + " yes"
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
