package tui

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/projectpage"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/runhistory"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/runview"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// The Projects view (package projectpage draws it): the list of projects
// grouped by kind, and the detail of the selected one with its sections. These
// are the places the shell turns its state into what the pages draw.

// The buttons of the detail of a project.
const (
	runMenuID    = "x:run-menu"
	openFolderID = "x:open-folder"
)

// projectsContent is the list of projects: the filter line, the counts, the
// projects under their kind headings, and the legend on the last row.
func (s Shell) projectsContent(w, h int) content {
	c := content{hits: map[int]hit{}}
	if !s.loadedProjects {
		return withLines(c, s.th.Muted(words.Say("tui.main.reading_the_workspace")))
	}
	if s.projectsErr != nil {
		return withLines(c, s.th.Failure(s.projectsErr.Error()))
	}
	under := s.actionAt(s.mouseX, s.mouseY)
	if s.list.Len() == 0 {
		c.hits[0], c.hits[1] = hit{kind: "action", id: "file.add"}, hit{kind: "action", id: "file.compose"}
		return withLines(c, s.newProjectButton(under), button.Render(s.th, words.Say("tui.main.compose_a_stack"), commands.ShortcutOf("file.compose"), s.buttonState("file.compose", under)),
			"", s.th.Muted(words.Say("tui.main.no_projects_yet_add_one_or")))
	}
	in := s.listInput()
	c.lines = append(projectpage.ListHead(s.th, in, w), "")
	body, rowOf, sel := projectpage.ListBody(s.th, in, w)
	box := s.pickBox(w, under)
	if h-len(c.lines)-2-len(box) < 3 {
		// A short list keeps its rows: the keys still run the picked projects.
		box = nil
	}
	avail := max(h-len(c.lines)-2-len(box), 1)
	top := s.projScroll.Show(max(sel, 0), avail, len(body))
	first := len(c.lines)
	for i := top; i < len(body) && i < top+avail; i++ {
		if idx, ok := rowOf[i]; ok {
			c.hits[len(c.lines)] = hit{kind: "project", id: s.list.Projects()[idx].Name, index: idx}
		}
		c.lines = append(c.lines, body[i])
	}
	copy(c.lines[first:], pane.Scrollbar(s.th, c.lines[first:], w, top, len(body)))
	for len(c.lines) < h-1-len(box) {
		c.lines = append(c.lines, "")
	}
	for _, b := range box {
		if len(b.hit.spans) > 0 {
			c.hits[len(c.lines)] = b.hit
		}
		c.lines = append(c.lines, b.line)
	}
	c.lines = append(c.lines, projectpage.Legend(s.th, w))
	return c
}

// listInput is what the list of projects shows: the projects the filter keeps,
// how each last ran, and the counts of every kind.
func (s Shell) listInput() projectpage.ListInput {
	in := projectpage.ListInput{Totals: map[string]int{}, Filter: s.list.Filter(), Typing: s.filtering, Selected: -1, Tick: s.tick}
	for _, p := range s.list.Projects() {
		if in.Totals[p.Kind] == 0 {
			in.Kinds = append(in.Kinds, p.Kind)
		}
		in.Totals[p.Kind]++
		if p.Affected {
			in.Changed++
		}
	}
	if _, ok := s.list.Selected(); ok {
		in.Selected = s.list.Cursor()
	}
	listed, indexes := s.list.Listed()
	for i, p := range listed {
		in.Rows = append(in.Rows, projectpage.ListRow{
			Index: indexes[i], Name: p.Name, Kind: p.Kind, Language: p.Language, Version: s.versionOf(p, projectlist.VersionW-1),
			Mark: s.projectMark(p.Name), Changed: p.Affected, Picked: s.picked[p.Name],
		})
	}
	in.Picked = len(s.pickedNames())
	return in
}

// lastTask is how a task of a project last ran in this session: the newest run
// that has it, and the task in that run; nil when none ran it.
func (s Shell) lastTask(id string) (*runview.Run, *runview.Task) {
	for i := len(s.runs) - 1; i >= 0; i-- {
		if t := s.runs[i].Task(id); t != nil {
			return s.runs[i], t
		}
	}
	return nil, nil
}

// projectMark is how the tasks of a project ran in the newest run that ran
// one of them: the worst state among them.
func (s Shell) projectMark(name string) projectpage.Mark {
	prefix := name + ":"
	for i := len(s.runs) - 1; i >= 0; i-- {
		best, found := runview.Queued, false
		for _, t := range s.runs[i].Tasks {
			if strings.HasPrefix(t.ID, prefix) && (!found || statusRank[t.Status] < statusRank[best]) {
				best, found = t.Status, true
			}
		}
		if found {
			return markOf(best)
		}
	}
	return projectpage.NotRun
}

// markOf is the mark of a task's state.
func markOf(st runview.Status) projectpage.Mark {
	switch st {
	case runview.Passed:
		return projectpage.Passed
	case runview.Cached:
		return projectpage.Cached
	case runview.Failed:
		return projectpage.Failed
	case runview.Running:
		return projectpage.Running
	case runview.Skipped:
		return projectpage.Skipped
	case runview.Cancelled:
		return projectpage.Cancelled
	}
	return projectpage.Queued
}

// resultWords say how a task ended, without its mark.
func resultWords(st runview.Status) string {
	switch st {
	case runview.Passed, runview.Cached:
		return words.Say("tui.projects.result.passed")
	case runview.Failed:
		return words.Say("tui.projects.result.failed")
	case runview.Running:
		return words.Say("tui.projects.result.running")
	case runview.Skipped:
		return words.Say("tui.projects.result.skipped")
	case runview.Cancelled:
		return words.Say("tui.projects.result.stopped")
	}
	return words.Say("tui.projects.result.queued")
}

// projectPage is true while the detail shows the sections of the selected
// project: on Projects, with a project selected, no run open and no Run… menu.
func (s Shell) projectPage() bool {
	if s.tab != commands.TabProjects || !s.inWorkspace || s.runShown() || s.projMenu {
		return false
	}
	_, ok := s.list.Selected()
	return ok
}

// projectKeys is true while the keyboard is in the sections of the detail.
func (s Shell) projectKeys() bool { return s.projectPage() && s.sideKeys() }

// taskRows are the rows of the Tasks table of a project.
func (s Shell) taskRows(p projectlist.Project) []projectpage.TaskRow {
	keys := s.taskKeys(p)
	var rows []projectpage.TaskRow
	for _, t := range p.Tasks {
		key := keys[t.Name]
		if a := s.taskAction(t.Name, key, p.Kind); a.Shortcut != "" {
			key = a.Shortcut
		}
		r := projectpage.TaskRow{Name: t.Name, Key: key, Cached: t.Cache, Result: words.Say("tui.projects.result.not_run"),
			OwnTerminal: commands.NeedsOwnTerminal(commands.RunArgv(t.Name, p.Name))}
		if run, task := s.lastTask(p.Name + ":" + t.Name); task != nil {
			r.Mark, r.Result = markOf(task.Status), resultWords(task.Status)
			if d := run.Elapsed(task, s.now()); d > 0 {
				r.Time = shortDuration(d)
			}
			r.Cache = cacheWords(task)
		}
		rows = append(rows, r)
	}
	return rows
}

// cacheWords say whether the cache answered a task that ran: hit (and from
// where), miss, or nothing while it has not ended.
func cacheWords(t *runview.Task) string {
	switch t.Status {
	case runview.Cached:
		if t.Remote {
			return words.Say("tui.projects.cache.remote")
		}
		return words.Say("tui.projects.cache.hit")
	case runview.Passed, runview.Failed:
		if t.Miss != "" {
			return words.Say("tui.projects.cache.miss_why", "why", t.Miss)
		}
		return words.Say("tui.projects.cache.miss")
	}
	return ""
}

// projectRuns are the runs of the session that ran a task of the project,
// newest first, each with its index in s.runs.
func (s Shell) projectRuns(name string) (rows []projectpage.RunRow, at []int) {
	prefix := name + ":"
	now := s.now()
	for i := len(s.runs) - 1; i >= 0; i-- {
		r := s.runs[i]
		var tasks []string
		best, found := runview.Queued, false
		failed := 0
		for _, t := range r.Ordered() {
			if !strings.HasPrefix(t.ID, prefix) {
				continue
			}
			tasks = append(tasks, strings.TrimPrefix(t.ID, prefix))
			if !found || statusRank[t.Status] < statusRank[best] {
				best, found = t.Status, true
			}
			if t.Status == runview.Failed {
				failed++
			}
		}
		if !found {
			continue
		}
		result := resultWords(best)
		if failed > 1 {
			result = words.Say("tui.projects.result.failed_count", "count", failed)
		}
		ago := words.Say("tui.projects.ago.now")
		if r.Finished {
			ago = runhistory.Ago(now.Sub(r.Ended))
		}
		rows = append(rows, projectpage.RunRow{Mark: markOf(best), Task: strings.Join(tasks, ", "), Result: result, Time: shortDuration(r.RunTime(now)), Ago: ago})
		at = append(at, i)
	}
	return rows, at
}

// projectHead is the top of the detail of a project.
func (s Shell) projectHead(p projectlist.Project, w int) ([]string, map[int][]projectpage.Span) {
	buttons := []projectpage.Button{{ID: runMenuID, Label: words.Say("tui.projects.run_menu"), Key: s.shortcutHere("projects.run-menu"), Primary: true}}
	if s.sources.Open != nil {
		buttons = append(buttons, projectpage.Button{ID: openFolderID, Label: words.Say("tui.projects.open_folder")})
	}
	in := projectpage.HeadInput{
		Path: p.Root, Kind: p.Kind, Language: p.Language, Stack: stackIfAny(p.Stack), Version: s.versionOf(p, 1<<10), Changed: p.Affected,
		Buttons: buttons, Hover: s.actionAt(s.mouseX, s.mouseY), Pressed: s.pressed, Section: s.projSection, Narrow: s.lay().Mode == layout.Narrow,
	}
	return projectpage.Head(s.th, in, w)
}

// stackIfAny is the text of a stack, or "" for a project without one.
func stackIfAny(stack string) string {
	if stack == "" {
		return ""
	}
	return stackText(stack)
}

// projectDetailContent is the detail of the selected project in w by h cells:
// the top that never scrolls, then the open section.
func (s Shell) projectDetailContent(w, h int) content {
	p, _ := s.list.Selected()
	c := content{hits: map[int]hit{}}
	head, spans := s.projectHead(p, w)
	for line, list := range spans {
		h := hit{kind: "action"}
		for _, sp := range list {
			h.spans = append(h.spans, hitSpan{from: sp.From, to: sp.To, id: sp.ID})
		}
		c.hits[line] = h
	}
	c.lines = head
	bodyH := max(h-len(head), 1)
	if s.projSection == projectpage.Settings {
		r := s.detailRect()
		px, py := s.mouseX-r.X, s.mouseY-r.Y-len(head)
		for i, line := range s.settings.ProjectPane(s.th, w, bodyH, px, py) {
			c.hits[len(c.lines)] = hit{kind: "psettings", index: i}
			c.lines = append(c.lines, line)
		}
		return c
	}
	body, bodyHits, cursor := s.sectionBody(p, w)
	top := pane.ClampTop(s.detailTop, len(body), bodyH)
	if cursor >= 0 {
		top = pane.ClampTop(pane.KeepVisible(top, cursor, len(body), bodyH), len(body), bodyH)
	}
	end := min(top+bodyH, len(body))
	for i := top; i < end; i++ {
		if bh, ok := bodyHits[i]; ok {
			c.hits[len(c.lines)] = bh
		}
		c.lines = append(c.lines, body[i])
	}
	copy(c.lines[len(head):], pane.Scrollbar(s.th, c.lines[len(head):], w, top, len(body)))
	return c
}

// sectionBody is the open section of the detail of a project, apart from
// Settings: its lines, what a click on each does, and the line of the row the
// keyboard is on (-1 for none).
func (s Shell) sectionBody(p projectlist.Project, w int) (lines []string, hits map[int]hit, cursor int) {
	hits, cursor = map[int]hit{}, -1
	focused := s.projectKeys()
	switch s.projSection {
	case projectpage.Dependencies:
		lines = projectpage.Uses(s.th, p.Deps, p.Dependents, w)
		lines = append(lines, "")
		under := s.actionAt(s.mouseX, s.mouseY)
		row := hit{kind: "action"}
		var b strings.Builder
		x := 1
		b.WriteString(" ")
		for i, a := range []detail.Action{
			{ID: "x:install", Label: words.Say("tui.detail.install_a_package"), Shortcut: commands.ShortcutOf("projects.install-package")},
			{ID: "x:uninstall", Label: words.Say("tui.detail.remove_a_package"), Shortcut: commands.ShortcutOf("projects.remove-package")},
		} {
			if i > 0 {
				b.WriteString("  ")
				x += 2
			}
			bw := button.Width(s.th, a.Label, a.Shortcut)
			row.spans = append(row.spans, hitSpan{from: x, to: x + bw, id: a.ID})
			b.WriteString(button.Render(s.th, a.Label, a.Shortcut, s.buttonState(a.ID, under)))
			x += bw
		}
		hits[len(lines)] = row
		lines = append(lines, pane.Fit(b.String(), w))
		if sec, ok := s.derivedSection(p); ok {
			derived, _ := detail.Panel{Sections: []detail.Section{sec}}.View(s.th, w, "")
			lines = append(append(lines, ""), derived...)
		}
	case projectpage.Runs:
		rows, _ := s.projectRuns(p.Name)
		var rowOf map[int]int
		lines, rowOf = projectpage.RunList(s.th, rows, w, s.projRunAt, focused, s.tick)
		for line, i := range rowOf {
			hits[line] = hit{kind: "prun", index: i}
			if i == s.projRunAt {
				cursor = line
			}
		}
	default:
		rows := s.taskRows(p)
		var rowOf map[int]int
		lines, rowOf = projectpage.TaskTable(s.th, rows, w, s.taskAt, focused, s.tick)
		for line, i := range rowOf {
			hits[line] = hit{kind: "ptask", id: rows[i].Name, index: i}
			if i == s.taskAt {
				cursor = line
			}
		}
	}
	return lines, hits, cursor
}
