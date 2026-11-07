package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/splitter"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/logview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
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
			return s.detailContent(w, h)
		}
		return s.projectsContent(w, h)
	case commands.TabRun:
		return s.runContent(w, h)
	case commands.TabHome:
		return s.homeContent(w)
	case commands.TabDoctor:
		if s.detailFull() {
			return s.detailContent(w, h)
		}
		// Beside the list the detail tells a check in full; alone, the list expands it.
		doctor := s.doctor
		doctor.Waited, doctor.Covered = s.waited(), len(s.list.Projects())
		doctor.Under, doctor.Pressed = doctorview.FilterOf(s.actionAt(s.mouseX, s.mouseY)), doctorview.FilterOf(s.pressed)
		lines, rowOf := doctor.View(s.th, w, !s.lay().HasDetail())
		sel := 0
		for row, i := range rowOf {
			if doctor.IsCursorRow(i) {
				sel = row
			}
		}
		total := len(lines)
		top := s.doctorScroll.Show(sel, max(h, 1), total)
		for row, i := range rowOf {
			if row >= top && row < top+h {
				c.hits[row-top] = hit{kind: "finding", index: i}
			}
		}
		// The headline is the filter buttons: each answers a click with its own key.
		for line, spans := range doctor.HeadlineSpans(s.th, w) {
			if line < top || line >= top+h {
				continue
			}
			filters := hit{kind: "action"}
			for _, sp := range spans {
				filters.spans = append(filters.spans, hitSpan{from: sp.From, to: sp.To, id: sp.Filter.CommandID()})
			}
			c.hits[line-top] = filters
		}
		lines = lines[top:min(top+h, len(lines))]
		c.lines = append(c.lines, pane.Scrollbar(s.th, lines, w, top, total)...)
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
		lines, rowOf := deps.View(s.th, w, max(h-len(c.lines), 1))
		for row, i := range rowOf {
			c.hits[len(c.lines)+row] = hit{kind: "dep", index: i}
		}
		c.lines = append(c.lines, lines...)
	case commands.TabSettings:
		in := s.mainInner()
		c.lines = s.settings.Pulse(s.tick, s.settingsWaited()).View(s.th, w, h, s.mouseX-in.X, s.mouseY-in.Y)
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

// projectsContent is the table of projects; the keys that run tasks on the
// selection are in the footer, and its detail is beside it or opens over it.
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
	// The explanations wrap instead of running off the pane.
	noun := "projects"
	if len(projects) == 1 {
		noun = "project"
	}
	c.lines = append(c.lines, pane.Wrap(s.th.Muted(words.Say("tui.main.affected_by_the_changes", "len", len(projects), "noun", noun, "affected", affected, "dot", s.th.Icons.Dot)), w, 0)...)
	c.lines = append(c.lines, pane.Wrap(s.th.Muted(words.Say("tui.main.projects_it_needs_projects_that_use")), w, 0)...)
	c.hits[len(c.lines)] = hit{kind: "action", id: "file.add"}
	c.lines = append(c.lines, s.newProjectButton(under), "")

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
	head += words.Say("tui.main.tasks", "cell", pane.Cell("↑ ↓", projectlist.DepsW))
	c.lines = append(c.lines, s.th.Muted(pane.Fit(head, w)))

	avail := max(h-len(c.lines), 1)
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
		line += pane.Cell(words.Say("tui.main.text", "len", len(p.Deps), "len2", len(p.Dependents)), projectlist.DepsW)
		line += ansi.Truncate(taskList(p), cols.Tasks, "…")
		if p.Name == s.selectedProject() {
			line = s.th.Selected(pane.Fit(pane.Strip(line), w))
		}
		c.hits[len(c.lines)] = hit{kind: "project", id: p.Name, index: i}
		c.lines = append(c.lines, pane.Fit(line, w))
	}
	copy(c.lines[first:], pane.Scrollbar(s.th, c.lines[first:], w, top, len(projects)))
	return c
}

// projectColumns is how the table of projects spends w cells: which of the
// Kind, Language and Stack columns it draws. The detail asks it too, so it
// leaves out only the facts the table shows.
func (s Shell) projectColumns(w int) projectlist.Columns {
	nameW, tasksW, hasStack := 8, 0, false
	for _, p := range s.list.Projects() {
		hasStack = hasStack || p.Stack != ""
		nameW = max(nameW, pane.Width(p.Name)+2) // the room for the last run's mark
		tasksW = max(tasksW, pane.Width(taskList(p)))
	}
	return projectlist.ColumnsFor(w, nameW, tasksW, hasStack)
}

// newProjectButton is the row that starts the Add wizard from the list of projects.
func (s Shell) newProjectButton(under string) string {
	return button.Render(s.th, words.Say("tui.main.new_project"), commands.ShortcutOf("file.add"), s.buttonState("file.add", under))
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

// ---- the Run tab ----------------------------------------------------------

// runGeom is where the parts of the Run tab sit, in lines of its body.
type runGeom struct {
	navRow, cmdRow    int
	progRow           int
	listTop, listRows int
	sepRow            int
	logTop, logRows   int
}

func (s Shell) runGeom(w, h int) runGeom {
	g := runGeom{}
	g.navRow, g.cmdRow, g.progRow, g.listTop = 0, 1, 2, 3
	n := len(s.runIDs())
	avail := max(h-g.listTop-2, 1)
	g.listRows = min(n, max(min(avail, 2), avail*2/5))
	if given, ok := s.split.Size(splitter.RunListID); ok {
		g.listRows = splitter.RunListRows(given, avail)
	}
	g.sepRow = g.listTop + g.listRows
	g.logTop = g.sepRow + 1
	g.logRows = max(h-g.logTop, 1)
	return g
}

// logHeight is the number of log lines visible in the Run tab.
// logTotal is how many rows the log shown takes at the width of the Run tab.
func (s Shell) logTotal() int {
	return len(logview.Rows(s.th, s.logLines(), s.mainInner().W))
}

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
	if s.run == nil || s.runHome {
		under := s.actionAt(s.mouseX, s.mouseY)
		if s.run == nil {
			c.lines = append(c.lines, "", s.th.Muted(words.Say("tui.main.nothing_has_run_yet")))
		} else {
			c.lines = append(c.lines, "", s.th.Muted(words.Say("tui.main.last_run", "commandLine", s.lastRun().CommandLine())))
		}
		c.lines = append(c.lines, "")
		// The scope of every key is spelled out; the buttons run on every project.
		scope := words.Say("tui.main.press_to_build", "shortcutOf", commands.ShortcutOf("run.build"))
		if name := s.state().RunProject(); name != "" {
			scope += words.Say("tui.main.the_project_selected_on_projects", "name", name)
		} else {
			scope += words.Say("tui.main.every_project_2")
		}
		c.lines = append(c.lines, pane.Wrap(s.th.Muted(words.Say("tui.main.to_test_or_start_one_on", "scope", scope, "shortcutOf", commands.ShortcutOf("run.test"))), w, 0)...)
		c.lines = append(c.lines, "")
		for _, b := range []struct{ id, label string }{
			{"run.verify", words.Say("tui.main.verify_everything_lint_typecheck_test")},
			{"run.build-all", words.Say("tui.main.build_everything", "len", s.list.Len())},
			{"run.test-all", words.Say("tui.main.test_everything", "len", s.list.Len())},
			{"run.lint-all", words.Say("tui.main.lint_everything", "len", s.list.Len())},
			{"run.typecheck-all", words.Say("tui.main.typecheck_everything", "len", s.list.Len())},
		} {
			c.hits[len(c.lines)] = hit{kind: "action", id: b.id}
			c.lines = append(c.lines, "  "+button.Render(s.th, b.label, commands.ShortcutOf(b.id), s.buttonState(b.id, under)))
		}
		c.lines = append(c.lines, "")
		if len(s.runs) > 0 && h-len(c.lines) > 1 {
			c.lines = append(c.lines, pane.Fit(s.th.Title(words.Say("tui.main.recent_runs"))+"  "+s.th.Muted(words.Say("tui.main.up_down_choose_enter_or_double")), w))
			rows, first := s.recentRows(w, h-len(c.lines))
			for i := range rows {
				c.hits[len(c.lines)] = hit{kind: "recent", index: first + i}
				c.lines = append(c.lines, rows[i])
			}
		}
		return c
	}
	g := s.runGeom(w, h)
	counts := s.run.Counts()
	now := s.now()
	took := shortDuration(s.run.RunTime(now))
	state := s.th.Info(words.Say("tui.main.running", "spin", s.th.Spin(s.tick)))
	switch {
	case s.stopping && s.running():
		state = s.th.Warning(words.Say("tui.main.stopping", "spin", s.th.Spin(s.tick)))
	case s.run.Cancelled:
		state = s.th.Muted(words.Say("tui.main.cancelled", "cancelled", s.th.Icons.Cancelled))
	case s.run.Finished && s.run.OK():
		state = s.th.Success(words.Say("tui.main.passed", "passed", s.th.Icons.Passed))
	case s.run.Finished:
		state = s.th.Failure(words.Say("tui.main.failed", "failed", s.th.Icons.Failed))
	}
	c.hits[len(c.lines)] = hit{kind: "nav"}
	c.lines = append(c.lines, s.navLine(w))
	scope := ""
	if label := commands.RunScopeLabel(s.run.Command, s.list.Len()); label != "" {
		scope = s.th.Title(label) + "  "
	}
	c.lines = append(c.lines, pane.Fit(" "+scope+s.th.Faint("$")+" "+s.th.Bold(s.run.CommandLine())+"   "+state+"  "+s.th.Muted(took), w))
	if counts.Total > 0 {
		bar := progress.Draw(s.th, counts.Done, counts.Total, min(max(w/2, 20), 40), counts.Failed > 0)
		if s.run.Cancelled {
			// A stopped run is not complete: the bar counts what ended and what was left.
			bar = progress.DrawCancelled(s.th, counts.Passed+counts.Cached+counts.Failed, counts.Skipped+counts.Cancelled, counts.Total, min(max(w/2, 20), 40))
		}
		c.lines = append(c.lines, pane.Fit(" "+bar+"  "+s.countsText(counts), w))
	} else if name, ok := addedProject(s.run.Command); ok && s.run.Finished && s.run.OK() {
		c.lines = append(c.lines, pane.Fit(addedNextSteps(s.th, name), w))
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
	copy(c.lines[g.listTop:g.sepRow], pane.Scrollbar(s.th, c.lines[g.listTop:g.sepRow], w, s.runTop, len(ids)))
	title := words.Say("tui.main.all_output")
	if s.runSel != "" {
		title = s.runSel
	}
	follow := "following"
	if !s.log.Following() {
		follow = "paused"
	}
	sep := words.Say("tui.main.log", "title", title, "follow", follow)
	c.lines = append(c.lines, s.th.Faint(pane.Fit(sep+strings.Repeat("─", max(w-pane.Width(sep), 0)), w)))
	c.lines = append(c.lines, s.log.Render(s.th, logview.Rows(s.th, s.logLines(), w), w, g.logRows)...)
	return c
}

// countsText is "✓ 4  ≈ 1  × 1  » 2  ● 2  ○ 3" without the zeros.
func (s Shell) countsText(c runview.Counts) string {
	var parts []string
	add := func(n int, st runview.Status) {
		if n > 0 {
			parts = append(parts, words.Say("tui.main.text_2", "statusIcon", s.statusIcon(st), "n", n, "st", st))
		}
	}
	add(c.Failed, runview.Failed)
	add(c.Running, runview.Running)
	add(c.Passed, runview.Passed)
	add(c.Cached, runview.Cached)
	add(c.Skipped, runview.Skipped)
	add(c.Cancelled, runview.Cancelled)
	add(c.Queued, runview.Queued)
	return strings.Join(parts, "  ")
}

// taskLine is one row of the task list.
func (s Shell) taskLine(id string, w int, now time.Time) string {
	if id == "" {
		line := words.Say("tui.main.all_output_2", "pointer", s.th.Icons.Pointer, "say", s.th.Muted(words.Say("tui.main.lines", "len", len(s.run.Lines))))
		if s.runSel == "" {
			return s.th.Selected(pane.Fit(pane.Strip(line), w))
		}
		return pane.Fit(line, w)
	}
	t := s.run.Task(id)
	if t == nil {
		return ""
	}
	nameW := taskNameW(w)
	state, took, reason := s.taskFacts(t, now)
	detail := state
	if took != "" {
		detail += "  " + took
	}
	if reason != "" {
		detail += "  " + reason
	}
	if t.Status == runview.Running && strings.HasSuffix(t.ID, ":lint") && s.run.Elapsed(t, now) >= slowLintAfter {
		detail += words.Say("tui.main.lint_checks_every_file_it_is")
	}
	line := " " + s.statusIcon(t.Status) + " " + pad(t.ID, nameW) + " " + s.th.Muted(detail)
	if id == s.runSel {
		line = s.th.Selected(pane.Fit(pane.Strip(line), w))
	}
	return pane.Fit(line, w)
}

// taskNameW is the width of the name column of the task list in w cells.
func taskNameW(w int) int { return min(max(w/3, 12), 28) }

// taskFacts are the words a task row tells after its name: the state, the time
// when there is one, and the reason unless the task passed.
func (s Shell) taskFacts(t *runview.Task, now time.Time) (state, took, reason string) {
	state = t.Status.String()
	if d := s.run.Elapsed(t, now); d > 0 {
		took = shortDuration(d)
	}
	if t.Reason != "" && t.Status != runview.Passed {
		reason = t.Reason
	}
	return state, took, reason
}

// taskRowShows says which facts the row of a task draws whole in a list w cells
// wide: the state, the time and the reason. A row cut short by a narrow list
// does not show the fact it lost, so the detail keeps it. The last cell is the
// scrollbar's.
func (s Shell) taskRowShows(t *runview.Task, w int, now time.Time) (state, took, reason bool) {
	st, tk, rs := s.taskFacts(t, now)
	end := 1 + pane.Width(s.statusIcon(t.Status)) + 1 + taskNameW(w) + 1 + pane.Width(st)
	state = end <= w-1
	if tk != "" {
		end += 2 + pane.Width(tk)
		took = end <= w-1
	}
	if rs != "" {
		end += 2 + pane.Width(rs)
		reason = end <= w-1
	}
	return state, took, reason
}

// slowLintAfter is how long a lint task runs before the list says why it takes
// time.
const slowLintAfter = 5 * time.Second

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
