package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/buttons"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/inspector"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/navigator"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
)

// hit says what a click on a line of the main view means.
type hit struct {
	// kind is "project", "task", "finding", "setting", "option", "tree" or "action".
	kind  string
	id    string
	index int
}

// content is the main view's body: rows of buttons on top, then lines, and
// what each line stands for.
type content struct {
	lines []string
	bars  []buttons.Bar
	hits  map[int]hit
}

// mainInner is where the main view's body is drawn: inside the tab row and the border.
func (s Shell) mainInner() layout.Rect {
	m := s.lay.Main
	return layout.Rect{X: m.X + 1, Y: m.Y + 2, W: max(m.W-2, 0), H: max(m.H-3, 0)}
}

func (s Shell) navigatorInner() layout.Rect { return s.lay.Navigator.Inner() }

func (s Shell) inspectorInner() layout.Rect { return s.lay.Inspector.Inner() }

// cmdButton is a button for a command, with its shortcut and its availability.
func (s Shell) cmdButton(id, label, icon string) buttons.Button {
	c, _ := commands.ByID(id)
	b := buttons.Button{ID: id, Label: label, Icon: icon, Shortcut: c.Shortcut, Disabled: c.Unavailable(s.state()) != ""}
	if s.running() && len(s.run.Command) > 1 {
		switch c.Kind {
		case commands.RunTask:
			b.Busy = s.run.Command[0] == "run" && s.run.Command[1] == c.Task && (!c.All || len(s.run.Command) == 2)
		case commands.Execute:
			b.Busy = strings.Join(s.run.Command, " ") == strings.Join(c.Argv, " ")
		}
	}
	return b
}

// addBars puts rows of buttons at the top of the content, then a blank line.
func (s Shell) addBars(c *content, bar buttons.Bar, w int) {
	c.bars = bar.Wrap(s.th, w)
	for range c.bars {
		c.lines = append(c.lines, "")
	}
	c.lines = append(c.lines, "")
}

// mainContent builds the body of the active tab for an area of w by h cells.
func (s Shell) mainContent(w, h int) content {
	c := content{hits: map[int]hit{}}
	switch s.tab {
	case tabNavigator:
		return s.navigatorContent(w, h)
	case commands.TabProjects:
		return s.projectsContent(w, h)
	case commands.TabRun:
		return s.runContent(w, h)
	case commands.TabDoctor:
		fixOne := buttons.Button{ID: "x:doctor-fix", Label: "Fix this…", Icon: "⚒", Shortcut: "F", Disabled: s.doctor.FixArgv() == nil || s.running()}
		s.addBars(&c, buttons.Bar{s.cmdButton("ws.doctor", "Re-check", "↻"), fixOne, s.cmdButton("ws.fix", "Fix all…", "⚒")}, w)
		lines, rowOf := s.doctor.View(s.th, w)
		for row, i := range rowOf {
			c.hits[len(c.lines)+row] = hit{kind: "finding", index: i}
		}
		c.lines = append(c.lines, lines...)
	case commands.TabRelease:
		rel := s.cmdButton("rel.release", "Release…", "⇧")
		rel.Disabled = rel.Disabled || !s.release.CanRelease()
		notes := buttons.Button{ID: "x:release-notes", Label: "Notes", Icon: "≡", Shortcut: "N", Disabled: !s.release.CanRelease() || s.running()}
		s.addBars(&c, buttons.Bar{s.cmdButton("rel.preview", "Preview", "▶"), notes, rel}, w)
		lines, rowOf := s.release.View(s.th, w)
		for row, i := range rowOf {
			c.hits[len(c.lines)+row] = hit{kind: "release", index: i}
		}
		c.lines = append(c.lines, lines...)
		c.lines = append(c.lines, s.previewLines(w, h-len(c.lines))...)
	case commands.TabPlugins:
		toggle := buttons.Button{ID: "x:plugin-toggle", Label: "Turn on/off…", Icon: "⏻", Shortcut: "E", Disabled: s.plugins.ToggleArgv() == nil || s.running()}
		s.addBars(&c, buttons.Bar{toggle, buttons.Button{ID: "x:reload", Label: "Reload", Icon: "↻", Shortcut: "R"}}, w)
		lines, rowOf := s.plugins.View(s.th, w, max(h-len(c.lines), 1))
		for row, i := range rowOf {
			c.hits[len(c.lines)+row] = hit{kind: "plugin", index: i}
		}
		c.lines = append(c.lines, lines...)
	case commands.TabDeps:
		up := buttons.Button{ID: "x:deps-up", Label: "Update…", Icon: "⇧", Shortcut: "U", Disabled: s.deps.UpArgv() == nil || s.running()}
		align := buttons.Button{ID: "x:deps-align", Label: "Align versions…", Icon: "=", Shortcut: "V", Disabled: s.deps.AlignArgv() == nil || s.running()}
		s.addBars(&c, buttons.Bar{up, align, buttons.Button{ID: "x:reload", Label: "Re-check", Icon: "↻", Shortcut: "R"}}, w)
		lines, rowOf := s.deps.View(s.th, w, max(h-len(c.lines), 1))
		for row, i := range rowOf {
			c.hits[len(c.lines)+row] = hit{kind: "dep", index: i}
		}
		c.lines = append(c.lines, lines...)
	case commands.TabRecipes:
		argv, _ := s.recipes.Argv()
		create := buttons.Button{ID: "x:recipes-create", Label: "Create…", Icon: "+", Shortcut: "C", Disabled: argv == nil || s.running()}
		s.addBars(&c, buttons.Bar{create, buttons.Button{ID: "x:reload", Label: "Reload", Icon: "↻", Shortcut: "R"}}, w)
		lines, rowOf := s.recipes.View(s.th, w)
		for row, i := range rowOf {
			c.hits[len(c.lines)+row] = hit{kind: "recipe", index: i}
		}
		c.lines = append(c.lines, lines...)
	case commands.TabSettings:
		lines, rowOf := s.settings.View(s.th, w, h)
		for row, i := range rowOf {
			c.hits[row] = hit{kind: "setting", index: i}
		}
		c.lines = lines
	case commands.TabGuide:
		lines, optionAt := s.guide.Body(s.th, w)
		for row, i := range optionAt {
			c.hits[row] = hit{kind: "option", index: i}
		}
		c.lines = lines
	}
	return c
}

// navigatorContent is the tree as a tab, below 80 columns.
func (s Shell) navigatorContent(w, h int) content {
	c := content{hits: map[int]hit{}}
	tree := s.tree.SetHeight(max(h, 1))
	c.lines = tree.View(s.th, w, -1, true, s.treeStatus)
	for i := range c.lines {
		c.hits[i] = hit{kind: "tree", index: i}
	}
	return c
}

// projectsContent is the table of projects with the buttons that run tasks on
// the selection. When the inspector is folded away it is drawn below.
func (s Shell) projectsContent(w, h int) content {
	c := content{hits: map[int]hit{}}
	// The scope is stated once, in the "On:" line below: the buttons stay short.
	s.addBars(&c, buttons.Bar{
		s.cmdButton("run.build", "Build", "▶"), s.cmdButton("run.test", "Test", "▶"), s.cmdButton("run.lint", "Lint", "▶"),
		s.cmdButton("run.typecheck", "Typecheck", "▶"), s.cmdButton("run.build-all", "Build all", "▶"),
		s.cmdButton("run.affected", "What changed", "⟳"),
	}, w)
	if !s.loadedProjects {
		return withLines(c, s.th.Muted("Reading the workspace…"))
	}
	if s.projectsErr != nil {
		return withLines(c, s.th.Failure(s.projectsErr.Error()))
	}
	projects := s.tree.Ordered()
	if len(projects) == 0 {
		return withLines(c, s.th.Muted("No projects yet. Use File > Add project… (Ctrl+N)."))
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
		nameW = max(nameW, pane.Width(p.Name))
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

	foldedLines := 0
	var folded []string
	var foldedHits map[int]string
	if !s.lay.HasInspector() {
		folded, foldedHits = s.inspectorPanel().View(s.th, w, "")
		foldedLines = min(len(folded)+1, max(h/2, 0))
	}
	avail := max(h-len(c.lines)-foldedLines, 1)
	sel := 0
	for i, p := range projects {
		if p.Name == s.selectedProject() {
			sel = i
		}
	}
	top := min(max(sel-avail/2, 0), max(len(projects)-avail, 0))
	for i := top; i < len(projects) && i < top+avail; i++ {
		p := projects[i]
		dot := s.th.Muted(s.th.Icons.Queued)
		if p.Affected {
			dot = s.th.Warning(s.th.Icons.Dot)
		}
		line := " " + dot + " " + pane.Cell(p.Name, cols.name+1)
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
	if foldedLines > 0 {
		c.lines = append(c.lines, s.th.Faint(strings.Repeat("─", w)))
		base := len(c.lines)
		for row, id := range foldedHits {
			c.hits[base+row] = hit{kind: "action", id: id}
		}
		c.lines = append(c.lines, folded...)
	}
	return c
}

// taskList is a project's task names on one line.
func taskList(p navigator.Project) string {
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

// treeStatus is the mark drawn after a project or task in the navigator: how
// the last run did there.
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
	bars, cmdRow, progRow int
	listTop, listRows     int
	sepRow                int
	logTop, logRows       int
}

func (s Shell) runButtons() buttons.Bar {
	follow := buttons.Button{ID: "run.follow", Label: "Pause log", Icon: "⏸", Shortcut: "Ctrl+F"}
	if !s.log.Following() {
		follow.Label, follow.Icon = "Follow log", "⏵"
	}
	again := buttons.Button{ID: "x:rerun", Label: "Run again", Icon: "↻", Disabled: s.lastArg == nil || s.running()}
	if !s.running() {
		// Stop has nothing to stop: it is shown only while a run goes.
		return buttons.Bar{follow, again}
	}
	return buttons.Bar{s.cmdButton("run.stop", "Stop", "■"), follow, again}
}

func (s Shell) runGeom(w, h int) runGeom {
	g := runGeom{bars: max(len(s.runButtons().Wrap(s.th, w)), 1)}
	g.cmdRow, g.progRow, g.listTop = g.bars, g.bars+1, g.bars+2
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
	c.bars = s.runButtons().Wrap(s.th, w)
	if s.run == nil {
		for range c.bars {
			c.lines = append(c.lines, "")
		}
		c.lines = append(c.lines, "", s.th.Muted("Nothing has run yet."))
		c.lines = append(c.lines, pane.Wrap(s.th.Muted("Pick a task in the navigator and press Ctrl+R, or press Ctrl+B to build."), w, 0)...)
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
	for range c.bars {
		c.lines = append(c.lines, "")
	}
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

// ---- the inspector --------------------------------------------------------

// inspectorPanel describes what is selected: a task of the last run in the Run
// tab, otherwise what the navigator has selected.
func (s Shell) inspectorPanel() inspector.Panel {
	if !s.inWorkspace {
		return inspector.Panel{
			Sections: []inspector.Section{{Heading: "No workspace", Items: []string{"No smith.toml here.", "File > New workspace… creates one."}}},
			Actions:  []inspector.Action{s.action("file.palette", "Command palette")},
		}
	}
	if s.tab == commands.TabDoctor {
		// A check is about the workspace or one project, not about the project
		// last selected in the navigator.
		return s.doctorPanel()
	}
	if s.tab == commands.TabRun && s.run != nil && s.runSel != "" {
		if t := s.run.Task(s.runSel); t != nil {
			rows := []inspector.Row{{Label: "state", Value: s.statusIcon(t.Status) + " " + t.Status.String()}}
			if d := s.run.Elapsed(t, s.now()); d > 0 {
				rows = append(rows, inspector.Row{Label: "time", Value: shortDuration(d)})
			}
			if t.Command != "" {
				rows = append(rows, inspector.Row{Label: "command", Value: t.Command})
			}
			if t.Reason != "" {
				rows = append(rows, inspector.Row{Label: "reason", Value: t.Reason})
			}
			rows = append(rows, inspector.Row{Label: "log", Value: fmt.Sprintf("%d lines", len(t.Log))})
			return inspector.Panel{
				Sections: []inspector.Section{{Heading: t.ID, Rows: rows}},
				Actions: []inspector.Action{
					{ID: "x:rerun", Label: "Run again", Disabled: s.running() || s.lastArg == nil},
					{ID: "x:projects", Label: "Show the project"},
				},
			}
		}
	}
	sel := s.tree.Selected()
	switch {
	case !sel.Valid:
		return s.workspacePanel()
	case sel.Kind == navigator.GroupRow:
		n := 0
		for _, p := range s.tree.Projects() {
			if p.Kind == sel.Group {
				n++
			}
		}
		return inspector.Panel{
			Sections: []inspector.Section{{Heading: sel.Group + "s", Rows: []inspector.Row{{Label: "projects", Value: fmt.Sprint(n)}}}},
			Actions:  []inspector.Action{s.action("run.build", "Build everything"), s.action("run.test", "Test everything")},
		}
	}
	p, ok := s.tree.Project(sel.Project)
	if !ok {
		return s.workspacePanel()
	}
	if sel.Kind == navigator.TaskRow {
		t, _ := p.Task(sel.Task)
		rows := []inspector.Row{{Label: "command", Value: t.Command}, {Label: "cached", Value: yesNo(t.Cache)}}
		if len(t.DependsOn) > 0 {
			rows = append(rows, inspector.Row{Label: "needs", Value: strings.Join(t.DependsOn, ", ")})
		}
		if t.Description != "" {
			rows = append(rows, inspector.Row{Label: "about", Value: t.Description})
		}
		if r := s.treeStatus(p.Name, t.Name); r != "" && s.run != nil {
			if rt := s.run.Task(p.Name + ":" + t.Name); rt != nil {
				rows = append(rows, inspector.Row{Label: "last run", Value: rt.Status.String()})
			}
		}
		return inspector.Panel{
			Sections: []inspector.Section{{Heading: p.Name + ":" + t.Name, Rows: rows}},
			Actions: []inspector.Action{
				s.action("run.selection", "Run"),
				{ID: "x:explain", Label: "Explain"},
				{ID: "x:why-miss", Label: "Why a cache miss"},
				{ID: "x:graph", Label: "Graph around the project"},
			},
		}
	}
	rows := []inspector.Row{{Label: "kind", Value: p.Kind}, {Label: "language", Value: p.Language}}
	if p.Stack != "" {
		rows = append(rows, inspector.Row{Label: "stack", Value: p.Stack})
	}
	rows = append(rows, inspector.Row{Label: "root", Value: p.Root})
	if p.Version != "" {
		rows = append(rows, inspector.Row{Label: "version", Value: p.Version})
	}
	rows = append(rows, inspector.Row{Label: "affected", Value: yesNoMark(s, p.Affected)})
	secs := []inspector.Section{{Heading: p.Name, Rows: rows}}
	if len(p.Deps) > 0 {
		secs = append(secs, inspector.Section{Heading: "Depends on", Items: p.Deps})
	}
	if len(p.Dependents) > 0 {
		secs = append(secs, inspector.Section{Heading: "Used by", Items: p.Dependents})
	}
	return inspector.Panel{Sections: secs, Actions: []inspector.Action{
		s.action("run.build", "Build"), s.action("run.test", "Test"), s.action("run.lint", "Lint"), s.action("run.typecheck", "Typecheck"),
		{ID: "x:explain", Label: "Explain"},
		{ID: "x:graph", Label: "Graph: needs and used by"},
	}}
}

// doctorPanel describes the selected check; with none selected it describes
// the workspace.
func (s Shell) doctorPanel() inspector.Panel {
	f, ok := s.doctor.Selected()
	if !ok {
		return s.workspacePanel()
	}
	state := s.th.Success(s.th.Icons.Passed) + " passed"
	switch {
	case !f.OK && f.Warning:
		state = s.th.Warning(s.th.Icons.Warning) + " warning"
	case !f.OK:
		state = s.th.Failure(s.th.Icons.Failed) + " failed"
	}
	rows := []inspector.Row{{Label: "state", Value: state}, {Label: "source", Value: f.Source}}
	if f.Detail != "" {
		rows = append(rows, inspector.Row{Label: "problem", Value: f.Detail})
	}
	if f.Remedy != "" {
		rows = append(rows, inspector.Row{Label: "remedy", Value: f.Remedy})
	}
	sec := inspector.Section{Heading: f.Check, Rows: rows}
	if f.Fix != nil {
		sec.Items = append(sec.Items, "fix: "+f.Fix.Description)
		for _, e := range f.Fix.Files {
			sec.Items = append(sec.Items, "  "+e.Path)
		}
	}
	return inspector.Panel{Sections: []inspector.Section{sec}, Actions: []inspector.Action{
		{ID: "x:doctor-fix", Label: "Fix this…", Shortcut: "F", Disabled: s.doctor.FixArgv() == nil || s.running()},
		s.action("ws.doctor", "Re-check"),
	}}
}

func (s Shell) workspacePanel() inspector.Panel {
	n, affected := len(s.tree.Projects()), 0
	for _, p := range s.tree.Projects() {
		if p.Affected {
			affected++
		}
	}
	rows := []inspector.Row{{Label: "workspace", Value: s.header}, {Label: "projects", Value: fmt.Sprint(n)}, {Label: "affected", Value: fmt.Sprint(affected)}}
	if s.branch != "" {
		rows = append(rows, inspector.Row{Label: "branch", Value: s.branch})
	}
	return inspector.Panel{
		Sections: []inspector.Section{{Heading: "Workspace", Rows: rows}},
		Actions:  []inspector.Action{s.action("ws.doctor", "Doctor"), s.action("rel.preview", "Preview the release")},
	}
}

// action is an inspector action for a command, with its shortcut and availability.
func (s Shell) action(id, label string) inspector.Action {
	c, _ := commands.ByID(id)
	return inspector.Action{ID: id, Label: label, Shortcut: c.Shortcut, Disabled: c.Unavailable(s.state()) != ""}
}

func (s Shell) inspectorActions() []inspector.Action { return s.inspectorPanel().Actions }

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
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
