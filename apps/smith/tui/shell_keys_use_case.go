package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/packageprompt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/palette"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseview"
)

// detailStep is how many columns [ and ] move the edge of the detail.
const detailStep = 4

// mouseHelp are the mouse actions listed in the shortcuts overlay.
var mouseHelp = []string{
	"Click: select   Double-click: open the detail (narrow terminals)",
	"Click a tab to open the view, a hint in the footer to press its key",
	"Click an action in the detail to do it",
	"Wheel: scroll lists and logs; over the detail it scrolls the detail",
	"Shift+drag selects text (Option+drag in macOS terminals)",
	"m turns the mouse off, so a plain drag selects text",
}

// editing is true while a text field has the keyboard.
func (s Shell) editing() bool {
	return (s.tab == commands.TabGuide && s.guide.Editing()) || (s.tab == commands.TabSettings && s.settings.Modal())
}

// formOpen is true while a guided form is open: its keys are the form's.
func (s Shell) formOpen() bool { return s.tab == commands.TabGuide && s.guide.Title() != "" }

func (s *Shell) key(msg tea.KeyMsg) tea.Cmd {
	name := msg.String()
	switch {
	case s.adopt != nil:
		return s.adoptKey(name)
	case s.fix != nil:
		return s.fixKey(name)
	case s.pkg != nil:
		p, out := s.pkg.Key(msg)
		s.pkg = &p
		return s.afterPackages(p, out)
	case s.confirm != nil:
		c, out := s.confirm.Key(name)
		s.confirm = &c
		return s.afterConfirm(c, out)
	case s.help != nil:
		h, out := s.help.Key(name, s.help.Total(s.th))
		s.help = &h
		if out.Done {
			s.help = nil
		}
		return nil
	case s.palOpen:
		p, res := s.pal.Key(msg)
		s.pal = p
		return s.afterPalette(res)
	}
	if name == "ctrl+c" {
		if s.running() {
			s.stop()
		} else {
			s.quit = true
		}
		return nil
	}
	if name == "q" && !s.editing() {
		switch {
		case s.running():
			s.say("A run is in progress (" + commands.ShortcutOf("run.stop") + " stops it)")
			return nil
		case s.tab == commands.TabGuide && s.guide.Title() != "":
			// Inside a form q goes back one step, as Esc does.
			return s.paneKey(msg)
		}
	}
	if c, reason, ok := commands.ByKey(name, s.state()); ok {
		if reason != "" {
			s.say(c.Title + " is unavailable: " + reason)
			return nil
		}
		return s.perform(c)
	}
	// j and k move as the arrows do, wherever the arrows would.
	if (name == "j" || name == "k") && !s.editing() && !s.formOpen() {
		msg = tea.KeyMsg{Type: map[string]tea.KeyType{"j": tea.KeyDown, "k": tea.KeyUp}[name]}
	}
	return s.page().Update(msg)
}

// openPackages asks which packages to install on, or remove from, the
// selected project.
func (s *Shell) openPackages(remove bool) {
	project := s.selectedProject()
	if project == "" {
		s.say("Select a project first: packages are installed on one project")
		return
	}
	var names []string
	for _, pr := range s.list.Projects() {
		names = append(names, pr.Name)
	}
	p := packageprompt.NewInstall(project, names)
	if remove {
		p = packageprompt.NewRemove(project, names)
	}
	p = p.Place(s.w, s.h)
	s.pkg = &p
}

// afterPackages acts on the package prompt's outcome: the command it shows runs.
func (s *Shell) afterPackages(p packageprompt.Prompt, out dialog.Outcome) tea.Cmd {
	if !out.Done {
		return nil
	}
	s.pkg = nil
	if out.Confirmed {
		return s.execute(p.Argv(), true)
	}
	return nil
}

// afterConfirm acts on the confirmation dialog's outcome.
func (s *Shell) afterConfirm(c dialog.Confirm, out dialog.Outcome) tea.Cmd {
	if !out.Done {
		return nil
	}
	s.confirm = nil
	if c.Info {
		return nil
	}
	if out.Confirmed {
		return s.execute(c.Final(), true)
	}
	s.say("Cancelled: " + commands.Line(c.Argv) + " was not run")
	return nil
}

func (s *Shell) afterPalette(res palette.Result) tea.Cmd {
	if res.Refused != "" {
		s.say(res.Refused)
	}
	if res.Closed {
		s.palOpen = false
		s.pal = s.pal.Reset()
	}
	if res.CommandID != "" {
		return s.activate(res.CommandID)
	}
	return nil
}

// activate runs a command by id: a palette entry, an action of the detail.
// Ids starting with "x:" are actions that belong to the selection.
func (s *Shell) activate(id string) tea.Cmd {
	if rest, ok := strings.CutPrefix(id, "x:"); ok {
		return s.activateSelection(rest)
	}
	c, ok := commands.ByID(id)
	if !ok {
		return nil
	}
	if reason := c.Unavailable(s.state()); reason != "" {
		s.say(c.Title + " is unavailable: " + reason)
		return nil
	}
	return s.perform(c)
}

// activateSelection runs the actions of the detail that depend on what is
// selected.
func (s *Shell) activateSelection(action string) tea.Cmd {
	project := s.selectedProject()
	if task, ok := strings.CutPrefix(action, "run:"); ok {
		if project == "" {
			return nil
		}
		return s.execute(commands.RunArgv(task, project), false)
	}
	switch action {
	case "install", "uninstall":
		s.openPackages(action == "uninstall")
	case "explain":
		if project != "" {
			return s.execute([]string{"explain", project}, false)
		}
	case "why":
		if p, ok := s.list.Selected(); ok {
			return s.execute([]string{"why", p.Root}, false)
		}
	case "derive-explain", "derive-apply", "derive-adopt", "derive-eject", "derive-replace":
		p, ok := s.list.Selected()
		if !ok {
			return nil
		}
		if argv := p.DeriveArgv(strings.TrimPrefix(action, "derive-")); argv != nil {
			return s.execute(argv, false)
		}
		s.say("No file of " + p.Name + " needs that")
	case "rerun":
		if argv := s.repeatArgv(); argv != nil {
			return s.execute(argv, false)
		}
	case "rerun-failed":
		if argv := s.failedArgv(); argv != nil {
			return s.execute(argv, false)
		}
		s.say("No task failed or was skipped")
	case "copy-log":
		s.copyLog()
	case "run-older", "run-newer":
		s.stepRun(map[string]int{"run-older": -1, "run-newer": 1}[action])
	case "projects":
		return s.gotoTab(commands.TabProjects)
	case "log":
		s.runSel, s.runHome = "", false
		return s.gotoTab(commands.TabRun)
	case "graph":
		if project != "" {
			return s.execute([]string{"graph", "--focus", project}, false)
		}
	case "release-notes":
		if s.release.CanRelease() {
			return s.execute([]string{"release", "--dry-run", "--notes"}, false)
		}
		s.say("There is no release to show the notes of")
	case "doctor-fix":
		if p, ok := s.doctor.SelectedFix(); ok {
			return s.openFix(p)
		}
		s.say("This check has no fix; its remedy says what to do")
	case "deps-up":
		if argv := s.deps.UpArgv(); argv != nil {
			return s.execute(argv, false)
		}
		s.say("No dependency to update")
	case "deps-align":
		if argv := s.deps.AlignArgv(); argv != nil {
			return s.execute(argv, false)
		}
		s.say("The one-version policy found nothing to align")
	case "reload":
		return s.reload()
	case "detail-zoom":
		s.zoomDetail()
	case "detail-wider":
		s.resizeDetail(detailStep)
	case "detail-narrower":
		s.resizeDetail(-detailStep)
	case "detail-up", "detail-down":
		if !s.detailVisible() {
			s.say("The detail is not on screen (" + commands.ShortcutOf("detail.zoom") + " shows it)")
			return nil
		}
		s.pageDetail(map[string]int{"detail-up": -1, "detail-down": 1}[action])
	case "detail-expand":
		if !s.detailVisible() {
			// Alone, the list unfolds the check under its line.
			s.doctor = s.doctor.Update(tea.KeyMsg{Type: tea.KeyEnter})
			return nil
		}
		s.detailAll = !s.detailAll
	}
	return nil
}

// perform carries out a command.
func (s *Shell) perform(c commands.Command) tea.Cmd {
	switch c.Kind {
	case commands.ShowTab:
		return s.gotoTab(c.Tab)
	case commands.Execute:
		return s.execute(c.Argv, false)
	case commands.RunTask:
		project := s.state().RunProject()
		if c.All {
			project = ""
		}
		return s.execute(commands.RunArgv(c.Task, project), false)
	case commands.Guided:
		if c.Guide == addProjectTitle {
			return s.openAddProject()
		}
		if g, ok := s.guide.WithDefault("lang", s.mainLanguage()).Pick(c.Guide); ok {
			s.guide = g
		}
		return s.gotoTab(commands.TabGuide)
	case commands.Palette:
		s.palOpen = true
		s.pal = s.pal.Reset()
	case commands.Help:
		s.openHelp()
	case commands.Quit:
		if s.proc != nil {
			s.proc.Stop()
		}
		s.quit = true
	case commands.Stop:
		s.stop()
	case commands.Follow:
		s.log = s.log.Toggle(s.logTotal(), s.logHeight())
		if s.log.Following() {
			s.say("Following the log")
		} else {
			s.say("Log paused: scroll or press " + commands.ShortcutOf("run.follow") + " to follow again")
		}
	case commands.CopyFocused:
		s.copyFocused()
	case commands.ToggleMouse:
		return s.toggleMouse()
	case commands.ViewAction:
		return s.activateSelection(c.Action)
	case commands.OpenPackagePrompt:
		s.openPackages(c.Action == "uninstall")
	case commands.OpenUndo:
		return s.openUndo()
	case commands.OpenReport:
		return s.openReport()
	case commands.OpenDoctor:
		return s.gotoTab(commands.TabDoctor)
	case commands.OpenAdoptWizard:
		return s.openAdopt()
	case commands.OpenFixWizard:
		if p, ok := s.doctor.FixAll(); ok {
			return s.openFix(p)
		}
		s.say("Doctor found nothing it can fix")
	}
	return nil
}

func (s *Shell) openHelp() {
	// The keys of the open view come first, then the keys every view shares,
	// all read from the keymap the keys themselves come from.
	var rows []dialog.HelpRow
	for _, r := range commands.HelpRows(s.state(), tabTitles[s.tab]) {
		rows = append(rows, dialog.HelpRow(r))
	}
	h := dialog.NewHelp(rows, mouseHelp).Place(s.w, s.h)
	s.help = &h
}

// gotoTab shows a view and loads it the first time.
func (s *Shell) gotoTab(id string) tea.Cmd {
	if !s.canShow(id) {
		return nil
	}
	if id != s.tab {
		// A result toast belongs to the screen it was raised on.
		s.toast = s.toast.Dismiss()
	}
	s.tab, s.detailOpen, s.zoom = id, false, false
	switch id {
	case commands.TabDoctor:
		if s.doctor.Report == nil && s.doctor.Err == nil && !s.doctor.Loading && s.sources.Doctor != nil {
			load := s.sources.Doctor
			s.doctor.Loading = true
			return func() tea.Msg { return load() }
		}
	case commands.TabSettings:
		var cmd tea.Cmd
		s.settings, cmd = s.settings.Open()
		return cmd
	case commands.TabRelease:
		if !s.release.Loaded && s.sources.Release != nil {
			load := s.sources.Release
			return func() tea.Msg { return load() }
		}
	case commands.TabDeps:
		if !s.deps.Loaded && s.sources.Deps != nil {
			load := s.sources.Deps
			return func() tea.Msg { return load() }
		}
	}
	return nil
}

// paneKey sends a key to the open view.
func (s *Shell) paneKey(msg tea.KeyMsg) tea.Cmd {
	name := msg.String()
	switch s.tab {
	case commands.TabHome:
		if name == "enter" {
			if id := s.nextAction().CommandID; id != "" {
				return s.activate(id)
			}
		}
	case commands.TabProjects:
		if s.detailFull() {
			return s.detailKey(name)
		}
		return s.projectsKey(name)
	case commands.TabDeps:
		s.deps = s.deps.Update(msg, s.listPage())
	case commands.TabRun:
		return s.runKey(name)
	case commands.TabRelease:
		s.release = s.release.Update(msg)
	case commands.TabSettings:
		return s.settingsKey(msg)
	case commands.TabDoctor:
		switch {
		case s.detailFull():
			return s.detailKey(name)
		case name == "enter" && s.lay().HasDetail():
			// Beside the detail Enter unfolds what the detail folded.
			s.detailAll = !s.detailAll
		default:
			s.doctor = s.doctor.Update(msg)
		}
	case commands.TabGuide:
		g, cmd := s.guide.Handle(msg)
		s.guide = g
		if g.Done() {
			argv, wizard := g.Argv, g.Wizard
			s.guide = g.Reset()
			if wizard == guide.WizardAdopt {
				return s.openAdopt()
			}
			if len(argv) > 0 && argv[0] == "add" && s.inWorkspace {
				// The plan comes first: what would be written, and which sibling it follows.
				return s.addPreview(argv)
			}
			if len(argv) > 0 {
				return s.execute(argv, false)
			}
			if s.inWorkspace {
				return s.gotoTab(commands.TabHome)
			}
		}
		if s.inWorkspace && g.Title() == "" {
			// Inside a workspace the guide is only its forms: backing out of
			// the first question leaves, there is no menu to land on.
			s.guide = g.Reset()
			return s.gotoTab(commands.TabHome)
		}
		return cmd
	}
	return nil
}

// projectsKey moves in the list of projects. Enter opens the detail over the
// list when it is not beside it.
func (s *Shell) projectsKey(name string) tea.Cmd {
	switch name {
	case "up":
		s.list = s.list.Move(-1)
	case "down":
		s.list = s.list.Move(1)
	case "pgup":
		s.list = s.list.Page(-1, s.listPage())
	case "pgdown":
		s.list = s.list.Page(1, s.listPage())
	case "home":
		s.list = s.list.Home()
	case "end":
		s.list = s.list.End()
	case "enter":
		if s.lay().Mode == layout.Narrow && s.list.Len() > 0 {
			s.detailOpen, s.detailCursor = true, 0
		}
	}
	return nil
}

// detailKey moves between the actions of the detail that fills the pane, runs
// the highlighted one, and goes back to the list.
func (s *Shell) detailKey(name string) tea.Cmd {
	actions := s.detailActions()
	switch name {
	case "esc", "left", "backspace":
		s.detailOpen, s.zoom = false, false
	case "pgup":
		s.pageDetail(-1)
	case "pgdown":
		s.pageDetail(1)
	case "up":
		s.detailCursor = max(s.detailCursor-1, 0)
	case "down":
		s.detailCursor = min(s.detailCursor+1, max(len(actions)-1, 0))
	case "enter":
		if s.detailCursor < len(actions) && !actions[s.detailCursor].Disabled {
			return s.activate(actions[s.detailCursor].ID)
		}
	}
	return nil
}

// runKey moves in the Run tab: the task list and the log.
func (s *Shell) runKey(name string) tea.Cmd {
	if s.run == nil {
		return nil
	}
	if s.runHome {
		if name == "enter" {
			s.runHome = false
		}
		return nil
	}
	ids := s.runIDs()
	at := 0
	for i, id := range ids {
		if id == s.runSel {
			at = i
		}
	}
	switch name {
	case "up":
		s.runSel = ids[max(at-1, 0)]
	case "down":
		s.runSel = ids[min(at+1, len(ids)-1)]
	case "esc":
		// The first Esc shows all output; the next, once nothing runs, the start screen.
		if s.runSel == "" && !s.running() {
			s.runHome = true
		}
		s.runSel = ""
	case "pgup":
		s.log = s.log.Page(-1, s.logTotal(), s.logHeight())
	case "pgdown":
		s.log = s.log.Page(1, s.logTotal(), s.logHeight())
	case "home":
		s.log = s.log.Home()
	case "end":
		s.log = s.log.End()
	}
	s.keepRunVisible()
	return nil
}

// runIDs are the choices in the Run tab's list: "" (all output), then the
// tasks in the order they are drawn.
func (s *Shell) runIDs() []string {
	ids := []string{""}
	if s.run != nil {
		for _, t := range s.run.Ordered() {
			ids = append(ids, t.ID)
		}
	}
	return ids
}

// logLines are the lines of the log being shown: one task's, or everything.
func (s *Shell) logLines() []string {
	if s.run == nil {
		return nil
	}
	if t := s.run.Task(s.runSel); t != nil && s.runSel != "" {
		return t.Log
	}
	return s.run.Lines
}

// loadProjects reads the project list again; nil without a source.
func (s *Shell) loadProjects() tea.Cmd {
	load := s.sources.Projects
	if load == nil {
		return nil
	}
	return func() tea.Msg { return load() }
}

// listPage is how many rows a page key moves in a list view.
func (s *Shell) listPage() int { return max(s.mainInner().H-6, 1) }

// reload forgets what the current view loaded and loads it again; nil for a
// view that has nothing to reload.
func (s *Shell) reload() tea.Cmd {
	switch s.tab {
	case commands.TabHome:
		s.doctor = doctorview.Model{}
		cmds := []tea.Cmd{s.loadProjects()}
		if load := s.sources.Doctor; load != nil {
			s.doctor.Loading = true
			cmds = append(cmds, func() tea.Msg { return load() })
		}
		return tea.Batch(cmds...)
	case commands.TabDoctor:
		s.doctor = doctorview.Model{}
	case commands.TabRelease:
		s.release = releaseview.Model{}
	case commands.TabDeps:
		s.deps = depsview.Model{}
	case commands.TabSettings:
		var cmd tea.Cmd
		s.settings, cmd = s.settings.Reload()
		return cmd
	default:
		return nil
	}
	return s.gotoTab(s.tab)
}
