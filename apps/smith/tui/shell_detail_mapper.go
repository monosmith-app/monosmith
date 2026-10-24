package tui

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
)

// hasDetail is true for the views whose selection has a detail: the project
// in Projects, the finding in Doctor, the task in Run. Beside the list on a
// wide screen; over it, opened with Enter, on a narrow one.
func (s Shell) hasDetail() bool {
	if !s.inWorkspace {
		return false
	}
	switch s.tab {
	case commands.TabProjects, commands.TabDoctor:
		return s.tab == commands.TabProjects || s.doctor.Report != nil
	case commands.TabRun:
		return s.run != nil
	}
	return false
}

// zoomed is true while the user made the detail fill the whole pane on a
// screen wide enough to show it beside the list.
func (s Shell) zoomed() bool {
	return s.zoom && s.hasDetail() && (s.tab == commands.TabProjects || s.tab == commands.TabDoctor)
}

// detailFull is true while the detail fills the pane instead of the list:
// opened on a narrow screen, or zoomed on a wide one.
func (s Shell) detailFull() bool { return s.detailOpen || s.zoomed() }

// detailVisible is true when the detail is on screen, beside the list or alone.
func (s Shell) detailVisible() bool { return s.detailFull() || s.lay().HasDetail() }

// detailRect is where the detail is drawn: beside the list, or over it.
func (s Shell) detailRect() layout.Rect {
	if s.detailFull() {
		return s.lay().Main.Inner()
	}
	return s.lay().DetailInner()
}

// detailContent is the detail drawn in w by h cells, scrolled to the user's
// place, with the action each of its lines stands for. The action under the
// pointer is highlighted, and, in the opened detail, the one the cursor is on.
func (s Shell) detailContent(w, h int) content {
	panel := s.detailPanel()
	all, rows := panel.View(s.th, w, "")
	_, hits := detail.Window(s.th, all, rows, s.detailTop, h)
	hover := ""
	if r := s.detailRect(); s.detailFull() {
		if a := panel.Actions; s.detailCursor < len(a) {
			hover = a[s.detailCursor].ID
		}
	} else if r.Contains(s.mouseX, s.mouseY) {
		hover = hits[s.mouseY-r.Y]
	}
	all, _ = panel.View(s.th, w, hover)
	lines, _ := detail.Window(s.th, all, rows, s.detailTop, h)
	c := content{lines: lines, hits: map[int]hit{}}
	for row, id := range hits {
		c.hits[row] = hit{kind: "action", id: id}
	}
	return c
}

// scrollDetail moves the detail by delta lines, never past its ends.
func (s *Shell) scrollDetail(delta int) {
	r := s.detailRect()
	lines, _ := s.detailPanel().View(s.th, r.W, "")
	s.detailTop = detail.ClampTop(s.detailTop+delta, len(lines), r.H)
}

// pageDetail scrolls the detail by a page up (dir -1) or down (dir 1).
func (s *Shell) pageDetail(dir int) {
	s.scrollDetail(dir * detail.PageStep(s.detailRect().H))
}

// detailIdentity changes when the selection the detail tells about changes,
// so the scroll and the expansion start over for the next one.
func (s Shell) detailIdentity() string {
	return fmt.Sprint(s.tab, "|", s.list.SelectedName(), "|", s.doctor.Cursor, s.doctor.Grouped, s.doctor.ShowPassed, "|", s.runSel)
}

// forgetDetailPlace starts the detail at its top, folded, when what it shows changed.
func (s *Shell) forgetDetailPlace() {
	if id := s.detailIdentity(); id != s.detailFor {
		s.detailFor, s.detailTop, s.detailAll = id, 0, false
	}
}

// resizeDetail makes the detail pane step columns wider (or narrower, when
// negative); the width is kept for the session.
func (s *Shell) resizeDetail(step int) {
	l := s.lay()
	if !l.HasDetail() {
		s.say("The detail is beside the list from 100 columns; z shows it alone")
		return
	}
	s.detailW = layout.ClampDetailWidth(s.w, l.Detail.W+step)
	s.say(fmt.Sprintf("Detail %d columns wide", s.lay().Detail.W))
}

// zoomDetail makes the detail fill the pane, or gives the list back.
func (s *Shell) zoomDetail() {
	switch {
	case s.detailFull():
		s.zoom, s.detailOpen = false, false
	case s.lay().Mode == layout.Narrow:
		s.detailOpen, s.detailCursor = true, 0
	default:
		s.zoom, s.detailCursor = true, 0
	}
	s.detailTop = 0
}

// detailPanel describes what is selected in the open view.
func (s Shell) detailPanel() detail.Panel {
	p := s.selectionPanel()
	p.Expanded = s.detailAll
	return p
}

func (s Shell) selectionPanel() detail.Panel {
	if !s.inWorkspace {
		return detail.Panel{
			Sections: []detail.Section{{Heading: "No workspace", Items: []string{"No smith.toml here.", "Press " + commands.ShortcutOf("file.new") + " to create a workspace."}}},
			Actions:  []detail.Action{s.action("file.palette", "Palette")},
		}
	}
	switch s.tab {
	case commands.TabDoctor:
		// A check is about the workspace or one project, not about the project
		// last selected in the list.
		return s.doctorPanel()
	case commands.TabRun:
		return s.runPanel()
	}
	return s.projectPanel()
}

// runPanel describes the task selected in the Run view, or the run as a whole.
func (s Shell) runPanel() detail.Panel {
	if s.run == nil {
		return s.workspacePanel()
	}
	if t := s.run.Task(s.runSel); t != nil && s.runSel != "" {
		rows := []detail.Row{{Label: "state", Value: s.statusIcon(t.Status) + " " + t.Status.String()}}
		if d := s.run.Elapsed(t, s.now()); d > 0 {
			rows = append(rows, detail.Row{Label: "time", Value: shortDuration(d)})
		}
		if t.Command != "" {
			rows = append(rows, detail.Row{Label: "command", Value: t.Command})
		}
		if t.Reason != "" {
			rows = append(rows, detail.Row{Label: "reason", Value: t.Reason})
		}
		rows = append(rows, detail.Row{Label: "log", Value: fmt.Sprintf("%d lines", len(t.Log))})
		return detail.Panel{
			Sections: []detail.Section{{Heading: t.ID, Rows: rows}},
			Actions: []detail.Action{
				{ID: "x:rerun", Label: "Run again", Disabled: s.running() || s.lastArg == nil},
				{ID: "x:projects", Label: "Show the projects"},
			},
		}
	}
	counts := s.run.Counts()
	rows := []detail.Row{{Label: "command", Value: s.run.CommandLine()}}
	if counts.Total > 0 {
		rows = append(rows, detail.Row{Label: "tasks", Value: s.countsText(counts)})
	}
	rows = append(rows, detail.Row{Label: "output", Value: fmt.Sprintf("%d lines", len(s.run.Lines))})
	return detail.Panel{
		Sections: []detail.Section{{Heading: "This run", Rows: rows}},
		Actions: []detail.Action{
			{ID: "x:rerun", Label: "Run again", Disabled: s.running() || s.lastArg == nil},
			{ID: "x:projects", Label: "Show the projects"},
		},
	}
}

// projectPanel describes the selected project, its tasks and what it can do.
func (s Shell) projectPanel() detail.Panel {
	p, ok := s.list.Selected()
	if !ok {
		return s.workspacePanel()
	}
	rows := []detail.Row{{Label: "kind", Value: p.Kind}, {Label: "language", Value: p.Language}}
	if p.Stack != "" {
		rows = append(rows, detail.Row{Label: "stack", Value: p.Stack})
	}
	rows = append(rows, detail.Row{Label: "root", Value: p.Root})
	if p.Version != "" {
		rows = append(rows, detail.Row{Label: "version", Value: p.Version})
	}
	rows = append(rows, detail.Row{Label: "affected", Value: yesNoMark(s, p.Affected)})
	secs := []detail.Section{{Heading: p.Name, Rows: rows}}
	if len(p.Tasks) > 0 {
		secs = append(secs, detail.Section{Heading: "Tasks", Items: taskItems(p), OneLine: true})
	}
	if len(p.Deps) > 0 {
		secs = append(secs, detail.Section{Heading: "Depends on", Items: p.Deps})
	}
	if len(p.Dependents) > 0 {
		secs = append(secs, detail.Section{Heading: "Used by", Items: p.Dependents})
	}
	var actions []detail.Action
	for _, t := range p.Tasks {
		actions = append(actions, s.taskAction(t.Name))
	}
	actions = append(actions,
		detail.Action{ID: "x:explain", Label: "Explain"},
		detail.Action{ID: "x:graph", Label: "Graph: needs and used by"},
	)
	return detail.Panel{Sections: secs, Actions: actions}
}

// taskItems is one line per task: its name, its command and what matters to a cache.
func taskItems(p projectlist.Project) []string {
	var out []string
	for _, t := range p.Tasks {
		line := t.Name
		if t.Command != "" {
			line += "  " + t.Command
		}
		var notes []string
		if t.Cache {
			notes = append(notes, "cached")
		}
		if len(t.DependsOn) > 0 {
			notes = append(notes, "needs "+strings.Join(t.DependsOn, ", "))
		}
		if len(notes) > 0 {
			line += "  (" + strings.Join(notes, "; ") + ")"
		}
		out = append(out, line)
	}
	return out
}

// taskAction is the action that runs a task on the selected project; the
// common four are the commands with their keys.
func (s Shell) taskAction(task string) detail.Action {
	for _, id := range []string{"run.build", "run.test", "run.lint", "run.typecheck"} {
		if c, _ := commands.ByID(id); c.Task == task {
			return s.action(id, "Run "+task)
		}
	}
	return detail.Action{ID: "x:run:" + task, Label: "Run " + task, Disabled: s.running()}
}

// doctorPanel tells the selected check in full: the list beside it has one
// line per check, so the problem, the remedy and the fix are said here, once.
func (s Shell) doctorPanel() detail.Panel {
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
	rows := []detail.Row{{Label: "state", Value: state}}
	if f.Source != "" {
		rows = append(rows, detail.Row{Label: "from", Value: f.Source})
	}
	planned, hasFix := s.doctor.FixFor(f)
	if hasFix {
		rows = append(rows, detail.Row{Label: "command", Value: "smith doctor --fix --only '" + f.Check + "'"})
	} else if c := commandIn(f.Remedy); c != "" {
		rows = append(rows, detail.Row{Label: "command", Value: c})
	}
	// What to do comes first, with the actions: the problem can be as long as
	// a page and must never push them off the screen.
	secs := []detail.Section{{Heading: f.Check, Rows: rows}}
	if f.Remedy != "" {
		secs = append(secs, detail.Section{Heading: "How to fix", Items: []string{f.Remedy}})
	}
	var after []detail.Section
	if hasFix {
		after = append(after, detail.Section{Heading: "Fix", Items: fixItems(planned), Collapse: 8})
	}
	if problems := doctorview.SplitProblem(f.Detail); len(problems) > 0 {
		after = append(after, detail.Section{Heading: "Problem", Items: problems, Bullets: len(problems) > 1, Collapse: 5})
	}
	// A check without a fix shows no fix action: it is remedy-only, and the
	// remedy above says what to do.
	var actions []detail.Action
	if hasFix {
		actions = append(actions, detail.Action{ID: "x:doctor-fix", Label: doctorview.FixLabel(planned), Shortcut: "f", Disabled: s.running()})
	}
	actions = append(actions, detail.Action{ID: "x:reload", Label: "Check again", Shortcut: commands.ShortcutOf("view.reload")})
	return detail.Panel{Sections: secs, After: after, Actions: actions}
}

// fixItems tell a fix in the detail: what it does, the projects it covers, the
// files it writes, the commands it runs and the questions it asks first.
func fixItems(p doctorfix.Planned) []string {
	items := []string{p.Description}
	if scope := doctorfix.Scope(p.AppliesTo); scope != "" {
		items = append(items, scope)
	}
	for _, e := range p.Files {
		verb := "writes "
		if e.Delete {
			verb = "removes "
		}
		items = append(items, "  "+verb+e.Path)
	}
	for _, st := range p.Steps {
		items = append(items, "  runs $ "+strings.Join(st.Argv, " "))
	}
	for _, c := range p.Choices {
		items = append(items, "  asks: "+c.Question)
	}
	return items
}

// commandIn is the first `command` a remedy spells out, or "".
func commandIn(remedy string) string {
	_, rest, ok := strings.Cut(remedy, "`")
	if !ok {
		return ""
	}
	c, _, ok := strings.Cut(rest, "`")
	if !ok {
		return ""
	}
	return c
}

func (s Shell) workspacePanel() detail.Panel {
	projects := s.list.Projects()
	affected := 0
	for _, p := range projects {
		if p.Affected {
			affected++
		}
	}
	rows := []detail.Row{{Label: "workspace", Value: s.header}, {Label: "projects", Value: fmt.Sprint(len(projects))}, {Label: "affected", Value: fmt.Sprint(affected)}}
	if s.branch != "" {
		rows = append(rows, detail.Row{Label: "branch", Value: s.branch})
	}
	return detail.Panel{
		Sections: []detail.Section{{Heading: "Workspace", Rows: rows}},
		Actions:  []detail.Action{s.action("ws.doctor", "Doctor"), s.action("rel.preview", "Preview the release")},
	}
}

// action is a detail action for a command, with its shortcut and availability.
func (s Shell) action(id, label string) detail.Action {
	c, _ := commands.ByID(id)
	return detail.Action{ID: id, Label: label, Shortcut: c.Shortcut, Disabled: c.Unavailable(s.state()) != ""}
}

func (s Shell) detailActions() []detail.Action { return s.detailPanel().Actions }
