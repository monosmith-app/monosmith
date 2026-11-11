package tui

import (
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// hasDetail is true for the views whose selection has a detail: the project
// in Projects (or the run that is open there), the finding in Doctor, the
// one-version policy in Deps and the option in Settings. Beside the list on a
// wide screen. On a narrow one Projects and Doctor open it over the list with
// Enter; Deps and Settings keep their text under the list.
func (s Shell) hasDetail() bool {
	if !s.inWorkspace {
		return false
	}
	switch s.tab {
	case commands.TabProjects, commands.TabDoctor:
		return s.tab == commands.TabProjects || s.doctor.Report != nil
	case commands.TabDeps:
		return s.deps.Loaded && s.deps.Err == nil
	case commands.TabSettings:
		return s.settings.Available()
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
// place, with the action each of its lines stands for. The button under the
// pointer shows its hover look, and its active look while the mouse holds it. In
// the opened detail the action the keyboard cursor is on has the cursor mark.
func (s Shell) detailContent(w, h int) content {
	panel := s.detailPanel()
	panel.Pressed = s.pressed
	if a := s.detailActions(); s.sideKeys() && s.detailCursor < len(a) {
		panel.Cursor = a[s.detailCursor].ID
	}
	// The fixed top of a project's detail (see detailSplit) is drawn above the body that scrolls.
	top, h := s.detailSplit(w, h)
	all, rows := panel.View(s.th, w, "")
	_, hits := pane.WindowMarked(s.th, all, rows, s.detailTop, h)
	hover := ""
	if r := s.detailRect(); r.Contains(s.mouseX, s.mouseY) {
		hover = hits[s.mouseY-r.Y-len(top)]
	}
	all, _ = panel.View(s.th, w, hover)
	lines, _ := pane.WindowMarked(s.th, all, rows, s.detailTop, h)
	lines = pane.Scrollbar(s.th, lines, w, pane.ClampTop(s.detailTop, len(all), h), len(all))
	c := content{lines: append(top, lines...), hits: map[int]hit{}}
	for row, id := range hits {
		c.hits[len(top)+row] = hit{kind: "action", id: id}
	}
	return c
}

// sideContent is what the right panel shows: the run while one is open in
// Projects, else the detail of the selection.
func (s Shell) sideContent(w, h int) content {
	if s.runShown() {
		return s.runPanelContent(w, h)
	}
	return s.detailContent(w, h)
}

// scrollDetail moves the detail by delta lines, never past its ends.
func (s *Shell) scrollDetail(delta int) {
	r := s.detailRect()
	lines, _ := s.detailPanel().View(s.th, r.W, "")
	_, h := s.detailSplit(r.W, r.H)
	s.detailTop = pane.ClampTop(s.detailTop+delta, len(lines), h)
}

// pageDetail scrolls the detail by a page up (dir -1) or down (dir 1).
func (s *Shell) pageDetail(dir int) {
	s.scrollDetail(dir * pane.PageStep(s.detailRect().H))
}

// detailIdentity changes when the selection the detail tells about changes,
// so the scroll and the expansion start over for the next one.
func (s Shell) detailIdentity() string {
	return fmt.Sprint(s.tab, "|", s.list.SelectedName(), "|", s.doctor.Cursor, s.doctor.Grouped, s.doctor.All, len(s.doctor.Folds), s.doctor.ShowPassed, s.doctor.HideFailed, s.doctor.HideWarnings, "|", s.runSel, "|", s.settings.Identity())
}

// forgetDetailPlace starts the detail at its top, folded, when what it shows changed.
func (s *Shell) forgetDetailPlace() {
	if id := s.detailIdentity(); id != s.detailFor {
		s.detailFor, s.detailTop, s.detailAll = id, 0, false
	}
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
	if s.splitsDetail() {
		_, p = splitPanel(p)
	}
	p.Expanded = s.detailAll
	p.ListVisible = !s.detailFull() && s.lay().HasDetail()
	p.Actions = s.withKeys(p.Actions)
	p.More = s.withKeys(p.More)
	return p
}

func (s Shell) selectionPanel() detail.Panel {
	if !s.inWorkspace {
		return detail.Panel{
			Sections: []detail.Section{{Heading: words.Say("tui.detail.no_workspace"), Items: []string{words.Say("tui.detail.no_toml"), words.Say("tui.detail.press_to_create_a_workspace", "shortcutOf", commands.ShortcutOf("file.new"))}}},
			Actions:  []detail.Action{s.action("file.palette", words.Say("tui.detail.palette"))},
		}
	}
	switch s.tab {
	case commands.TabDeps:
		return detail.Panel{Sections: []detail.Section{{Heading: depsview.PolicyTitle(), Items: s.deps.PolicyItems(s.th)}}}
	case commands.TabSettings:
		return s.settingsPanel()
	case commands.TabDoctor:
		// A check is about the workspace or one project, not about the project
		// last selected in the list.
		return s.doctorPanel()
	}
	return s.projectPanel()
}

// settingsPanel tells the selected option of Settings in the panel beside the
// tab: what it is, what it does and which key changes it.
func (s Shell) settingsPanel() detail.Panel {
	lines := s.settings.DetailLines(s.th, max(s.detailInner().W-4, 20))
	if len(lines) == 0 {
		return detail.Panel{}
	}
	return detail.Panel{Sections: []detail.Section{{Items: lines}}}
}

// projectPanel describes the selected project, its tasks and what it can do.
func (s Shell) projectPanel() detail.Panel {
	p, ok := s.list.Selected()
	if !ok {
		return s.workspacePanel()
	}
	// The table beside the detail has the Kind, Language and Stack columns when
	// it has the room for them.
	cols := s.projectColumns(s.mainInner().W)
	rows := []detail.Row{{Label: "kind", Value: p.Kind, ShownByList: cols.Kind}, {Label: "language", Value: p.Language, ShownByList: cols.Lang}}
	if p.Stack != "" {
		rows = append(rows, detail.Row{Label: "stack", Value: stackText(p.Stack), ShownByList: cols.Stack})
	}
	rows = append(rows, detail.Row{Label: "root", Value: p.Root})
	if v := s.versionOf(p, 1<<10); v != "" {
		// The pending release of the Release tab shows as "current → next".
		rows = append(rows, detail.Row{Label: "version", Value: v})
	}
	rows = append(rows, detail.Row{Label: "affected", Value: yesNoMark(s, p.Affected)})
	// The neighbours are names in a line that wraps, which takes fewer rows than a list.
	if len(p.Deps) > 0 {
		rows = append(rows, detail.Row{Label: words.Say("tui.detail.row.dependencies"), Value: strings.Join(p.Deps, ", ")})
	}
	if len(p.Dependents) > 0 {
		rows = append(rows, detail.Row{Label: words.Say("tui.detail.row.dependants"), Value: strings.Join(p.Dependents, ", ")})
	}
	secs := []detail.Section{{Heading: p.Name, Rows: rows}}
	var after []detail.Section
	if sec, ok := s.derivedSection(p); ok {
		after = append(after, sec)
	}
	var actions []detail.Action
	keys := s.taskKeys(p)
	for _, t := range p.Tasks {
		a := s.taskAction(t.Name, keys[t.Name], p.Kind)
		a.Hint = taskHint(t)
		if _, named := taskVerbs[t.Name]; named {
			// The button names what the task does to this kind; the task keeps its own name here.
			a.Hint += words.Say("tui.detail.task_command", "task", t.Name)
		}
		actions = append(actions, a)
	}
	actions = append(actions,
		detail.Action{ID: "x:install", Label: words.Say("tui.detail.install_a_package"), Hint: words.Say("tui.detail.adds_packages_to_with_its_own", "name", p.Name)},
		detail.Action{ID: "x:uninstall", Label: words.Say("tui.detail.remove_a_package")},
		detail.Action{ID: "x:explain", Label: words.Say("tui.detail.explain"), Hint: words.Say("tui.detail.explain_hint")},
		detail.Action{ID: "x:why", Label: words.Say("tui.detail.why_who_owns_these_files"), Hint: words.Say("tui.detail.what_produced_the_project_s_files")},
		detail.Action{ID: "x:graph", Label: words.Say("tui.detail.graph_needs_and_used_by")},
	)
	actions = append(actions, s.derivedActions(p)...)
	title := ""
	if len(p.Tasks) > 0 {
		title = words.Say("tui.detail.run_on", "name", p.Name)
	}
	// The forms about the project have no key: they are under More, after the keyed actions.
	more := append(s.projectFormActions(p), s.projectMoreActions()...)
	return detail.Panel{Sections: secs, Actions: actions, ActionsTitle: title, More: more, MoreTitle: words.Say("tui.detail.more"), After: after}
}

// taskHint says in plain words what running a task does, what has to run
// before it, and whether smith remembers its result. The command behind it is
// in the run panel once it ran.
func taskHint(t projectlist.Task) string {
	var parts []string
	switch {
	case t.Description != "":
		parts = append(parts, t.Description)
	case taskMeanings[t.Name] != "":
		parts = append(parts, words.Say(taskMeanings[t.Name]))
	default:
		parts = append(parts, words.Say("tui.detail.a_task_of_this_project"))
	}
	if len(t.DependsOn) > 0 {
		parts = append(parts, needsText(t.DependsOn))
	}
	if t.Cache {
		parts = append(parts, words.Say("tui.detail.result_remembered_skipped_when_nothing_c"))
	}
	return strings.Join(parts, " · ")
}

// taskMeanings are the catalog key of what each common task does.
var taskMeanings = map[string]string{
	"build":      "tui.detail.task.build",
	"lint":       "tui.detail.task.lint",
	"test":       "tui.detail.task.test",
	"typecheck":  "tui.detail.task.typecheck",
	"package":    "tui.detail.task.package",
	"format":     "tui.detail.task.format",
	"deps-check": "tui.detail.task.deps_check",
	"prune":      "tui.detail.task.prune",
	"audit":      "tui.detail.task.audit",
}

// stackMeanings are the catalog key of what a stack gives a project, for the stack row.
var stackMeanings = map[string]string{
	"node-jest":   "tui.detail.stack.node_jest",
	"node-vitest": "tui.detail.stack.node_vitest",
}

// stackText is the stack's name, with what it gives when smith knows.
func stackText(stack string) string {
	if m := stackMeanings[stack]; m != "" {
		return stack + " (" + words.Say(m) + ")"
	}
	return stack
}

// needsText says in plain words what has to run before a task: "^build" is
// the build of the projects this one needs, "build" this project's own.
func needsText(deps []string) string {
	var parts []string
	for _, d := range deps {
		switch {
		case strings.HasPrefix(d, "^"):
			parts = append(parts, words.Say("tui.detail.the_of_the_dependencies", "d", d[1:]))
		case strings.Contains(d, ":"):
			parts = append(parts, d)
		default:
			parts = append(parts, words.Say("tui.detail.its_own", "d", d))
		}
	}
	return words.Say("tui.detail.runs_after", "tasks", strings.Join(parts, words.Say("tui.detail.and")))
}

// taskVerbs are the catalog key of the verb of each well-known task, and
// kindNouns the catalog key of what a project of each kind is called.
var (
	taskVerbs = map[string]string{
		"build":     "tui.detail.verb.build",
		"test":      "tui.detail.verb.test",
		"lint":      "tui.detail.verb.lint",
		"typecheck": "tui.detail.verb.typecheck",
		"package":   "tui.detail.verb.package",
	}
	kindNouns = map[string]string{
		"app":     "tui.detail.noun.app",
		"lib":     "tui.detail.noun.lib",
		"package": "tui.detail.noun.package",
	}
)

// taskLabel is the label of the button that runs a task on a project of a kind:
// the well-known tasks say what they do to that kind ("Build app", "Package
// library"), because "package" is also a kind of project. Any other task, or
// any other kind, reads "Run <task>".
func taskLabel(task, kind string) string {
	verb, noun := taskVerbs[task], kindNouns[kind]
	switch {
	case verb == "" || noun == "":
		return words.Say("tui.detail.run", "task", task)
	case task == "package" && kind == "package":
		return words.Say("tui.detail.pack_package")
	}
	return words.Say("tui.detail.task_label", "verb", words.Say(verb), "kind", words.Say(noun))
}

// taskAction is the action that runs a task on the selected project, whose kind
// names the button; the common four are the commands with their keys.
func (s Shell) taskAction(task, key, kind string) detail.Action {
	label := taskLabel(task, kind)
	for _, id := range []string{"run.build", "run.test", "run.lint", "run.typecheck"} {
		if c, _ := commands.ByID(id); c.Task == task {
			return s.action(id, label)
		}
	}
	return detail.Action{ID: "x:run:" + task, Label: label, Shortcut: key, Disabled: s.running()}
}

// doctorPanel tells the selected check in full: the list beside it has one
// line per check, so the problem, the remedy and the fix are said here, once.
func (s Shell) doctorPanel() detail.Panel {
	f, ok := s.doctor.Selected()
	if !ok {
		if e, ok := s.doctor.SelectedExtra(); ok {
			return s.extraPanel(e)
		}
		return s.workspacePanel()
	}
	planned, hasFix := s.doctor.FixFor(f)
	// Three contexts, in this order: the actions, what is wrong, how to fix it.
	// The state is the icon of the highlighted row. Each fact is said once. The problem folds, so it cannot
	// push the remedy off the screen.
	secs := []detail.Section{{Heading: f.Check}}
	var actions []detail.Action
	if hasFix {
		actions = append(actions, detail.Action{ID: "x:doctor-fix", Label: doctorview.FixLabel(planned), Hint: planned.Description, Disabled: s.running()})
	}
	if id := doctorview.RunNowAction(f); id != "" {
		actions = append(actions, detail.Action{ID: "x:form:" + id, Label: words.Say("tui.detail.run_check_now"), Hint: words.Say("tui.detail.run_check_now_hint", "action", id), Disabled: s.running()})
	}
	actions = append(actions, detail.Action{ID: "x:reload", Label: words.Say("tui.detail.check_again"), Shortcut: commands.ShortcutOf("view.reload"), Hint: words.Say("tui.detail.run_every_check_again_nothing_changes")})
	if file := doctorview.FileOf(f); file != "" {
		actions = append(actions, detail.Action{ID: "x:why-file", Label: words.Say("tui.detail.why_file", "file", file), Hint: words.Say("tui.detail.why_file_hint"), Disabled: s.running()})
	}
	var after []detail.Section
	how := detail.Section{Heading: words.Say("tui.detail.how_to_fix")}
	if f.Remedy != "" {
		how.Items = append(how.Items, f.Remedy)
	}
	switch {
	case hasFix:
		// The fix title is on its button; the plan adds what the button does not say.
		how.Items = append(how.Items, fixItems(planned)...)
		how.Tail = []string{words.Say("tui.detail.or_in_a_terminal_smith_doctor", "check", f.Check)}
	case commandIn(f.Remedy) != "":
		how.Tail = []string{words.Say("tui.detail.in_a_terminal", "commandIn", commandIn(f.Remedy))}
	}
	// What is wrong is the cause, one line each. What the tool printed about its
	// own build goes behind a fold of its own, and no text or path shows twice.
	problems, notes, items := doctorview.TellOnce(f.Check, f.Detail, how.Items)
	how.Items = items
	if len(problems) > 0 {
		after = append(after, detail.Section{Heading: words.Say("tui.detail.what_is_wrong"), Items: problems, Bullets: len(problems) > 1, Collapse: 3})
	}
	if len(notes) > 0 {
		after = append(after, detail.Section{Heading: words.Say("tui.detail.build_details"), Items: notes, Fold: true})
	}
	if len(how.Items)+len(how.Tail) > 0 {
		how.Collapse = 8
		after = append(after, how)
	}
	return detail.Panel{Sections: secs, After: after, Actions: actions}
}

// extraPanel tells a check doctor does not run itself, with the button that
// runs it now.
func (s Shell) extraPanel(e doctorview.Extra) detail.Panel {
	return detail.Panel{
		Sections: []detail.Section{{Heading: e.Title, Items: []string{words.Say("tui.detail.extra_not_run")}}},
		Actions:  []detail.Action{{ID: "x:extra", Label: words.Say("tui.detail.run_check_now"), Hint: words.Say("tui.detail.run_extra_hint", "command", "smith "+strings.Join(e.Run, " ")), Disabled: s.running()}},
	}
}

// fixItems tell a fix in the detail beyond its title, which is on its button:
// the projects it covers, the files it writes, the commands it runs and the
// questions it asks first.
func fixItems(p doctorfix.Planned) []string {
	var items []string
	if scope := doctorfix.Scope(p.AppliesTo); scope != "" {
		items = append(items, scope)
	}
	for _, e := range p.Files {
		verb := words.Say("tui.detail.writes")
		if e.Delete {
			verb = words.Say("tui.detail.removes")
		}
		items = append(items, "  "+verb+e.Path)
	}
	for _, st := range p.Steps {
		items = append(items, words.Say("tui.detail.runs", "list", strings.Join(st.Argv, " ")))
	}
	for _, c := range p.Choices {
		items = append(items, words.Say("tui.detail.asks", "question", c.Question))
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
	if n := derivedNeedingAction(projects); n > 0 {
		rows = append(rows, detail.Row{Label: words.Say("tui.detail.derived_files"), Value: words.Say("tui.detail.need_action_pick_the_project_then", "n", n)})
	}
	return detail.Panel{
		Sections: []detail.Section{{Heading: words.Say("tui.detail.workspace"), Rows: rows}},
		Actions:  []detail.Action{s.action("ws.doctor", words.Say("tui.detail.doctor")), s.action("rel.preview", words.Say("tui.detail.preview_the_release"))},
	}
}

// action is a detail action for a command, with its shortcut and availability.
func (s Shell) action(id, label string) detail.Action {
	c, _ := commands.ByID(id)
	return detail.Action{ID: id, Label: label, Shortcut: s.shortcutHere(id), Disabled: c.Unavailable(s.state()) != ""}
}

// detailActions are the buttons of the detail, in the order the keyboard cursor visits them.
func (s Shell) detailActions() []detail.Action {
	p := s.detailPanel()
	return append(append([]detail.Action(nil), p.Actions...), p.More...)
}
