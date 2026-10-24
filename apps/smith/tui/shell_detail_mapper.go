package tui

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/detail"
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

// detailRect is where the detail is drawn: beside the list, or over it.
func (s Shell) detailRect() layout.Rect {
	if s.detailOpen {
		return s.lay().Main.Inner()
	}
	return s.lay().DetailInner()
}

// detailContent is the detail drawn in w by h cells, with the action each of
// its lines stands for. The action under the pointer is highlighted, and, in
// the opened detail, the one the cursor is on.
func (s Shell) detailContent(w, h int) content {
	panel := s.detailPanel()
	_, rows := panel.View(s.th, w, "")
	hover := ""
	if r := s.detailRect(); s.detailOpen {
		if a := panel.Actions; s.detailCursor < len(a) {
			hover = a[s.detailCursor].ID
		}
	} else if r.Contains(s.mouseX, s.mouseY) {
		hover = rows[s.mouseY-r.Y]
	}
	lines, _ := panel.View(s.th, w, hover)
	c := content{lines: lines, hits: map[int]hit{}}
	for row, id := range rows {
		c.hits[row] = hit{kind: "action", id: id}
	}
	return c
}

// detailPanel describes what is selected in the open view.
func (s Shell) detailPanel() detail.Panel {
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
	if f.Fix != nil && f.Fix.Safe {
		rows = append(rows, detail.Row{Label: "command", Value: "smith doctor --fix --only '" + f.Check + "'"})
	} else if c := commandIn(f.Remedy); c != "" {
		rows = append(rows, detail.Row{Label: "command", Value: c})
	}
	secs := []detail.Section{{Heading: f.Check, Rows: rows}}
	if f.Detail != "" {
		secs = append(secs, detail.Section{Heading: "Problem", Items: []string{f.Detail}})
	}
	if f.Remedy != "" {
		secs = append(secs, detail.Section{Heading: "Remedy", Items: []string{f.Remedy}})
	}
	if f.Fix != nil {
		items := []string{f.Fix.Description}
		for _, e := range f.Fix.Files {
			verb := "writes "
			if e.Delete {
				verb = "removes "
			}
			items = append(items, "  "+verb+e.Path)
		}
		secs = append(secs, detail.Section{Heading: "Fix", Items: items})
	}
	return detail.Panel{Sections: secs, Actions: []detail.Action{
		{ID: "x:doctor-fix", Label: "Fix this…", Shortcut: "f", Disabled: s.doctor.FixArgv() == nil || s.running()},
		{ID: "x:reload", Label: "Check again", Shortcut: commands.ShortcutOf("view.reload")},
	}}
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
