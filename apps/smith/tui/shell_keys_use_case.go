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
	"github.com/monosmith-app/monosmith/apps/smith/tui/helpdialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/palette"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releasepick"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runhistory"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// mouseHelp are the mouse actions listed in the shortcuts overlay.
func mouseHelp() []string {
	return []string{
		words.Say("tui.keys.mouse_help.click"),
		words.Say("tui.keys.mouse_help.tab"),
		words.Say("tui.keys.mouse_help.action"),
		words.Say("tui.keys.mouse_help.drag"),
		words.Say("tui.keys.mouse_help.select"),
		words.Say("tui.keys.mouse_help.wheel"),
		words.Say("tui.keys.mouse_help.shift"),
		words.Say("tui.keys.mouse_help.mouse"),
	}
}

// editing is true while a text field has the keyboard.
func (s Shell) editing() bool {
	return (s.tab == commands.TabGuide && s.guide.Editing()) || (s.tab == commands.TabSettings && s.settings.Modal())
}

// formOpen is true while a guided form is open: its keys are the form's.
func (s Shell) formOpen() bool { return s.tab == commands.TabGuide && s.guide.Title() != "" }

func (s *Shell) key(msg tea.KeyMsg) tea.Cmd {
	name := msg.String()
	if s.dropSelection(name) {
		return nil
	}
	switch {
	case s.result != nil:
		return s.resultKey(name)
	case s.cache != nil:
		return s.cacheKey(msg)
	case s.tools != nil:
		return s.toolsKey(name)
	case s.compose != nil:
		return s.composeKey(msg)
	case s.adopt != nil:
		return s.adoptKey(name)
	case s.fix != nil:
		return s.fixKey(name)
	case s.pkg != nil:
		p, out := s.pkg.Key(msg)
		s.pkg = &p
		return s.afterPackages(p, out)
	case s.relpick != nil:
		m, out := s.relpick.Key(name)
		s.relpick = &m
		return s.afterReleasePick(m, out)
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
	if s.splitKey(name) {
		return nil
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
			s.say(words.Say("tui.keys.a_run_is_in_progress_stops", "shortcutOf", commands.ShortcutOf("run.stop")))
			return nil
		case s.tab == commands.TabGuide && s.guide.Title() != "":
			// Inside a form q goes back one step, as Esc does.
			return s.paneKey(msg)
		}
	}
	if name == "?" && s.formOpen() && s.guide.HasGlossary() {
		// On a question that has words to explain, ? opens or folds them.
		return s.paneKey(msg)
	}
	if c, reason, ok := commands.ByKey(name, s.state()); ok {
		if reason != "" {
			s.say(words.Say("tui.keys.is_unavailable", "title", c.Title, "reason", reason))
			return nil
		}
		return s.perform(c)
	}
	if cmd, ok := s.taskKey(name); ok {
		return cmd
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
		s.say(words.Say("tui.keys.select_a_project_first_packages_are"))
		return
	}
	var names []string
	languages := map[string]string{}
	for _, pr := range s.list.Projects() {
		names = append(names, pr.Name)
		languages[pr.Name] = pr.Language
	}
	p := packageprompt.NewInstall(project, names)
	if remove {
		p = packageprompt.NewRemove(project, names)
	}
	p = p.WithLanguages(languages).WithLibraries(func(project string) []packageprompt.Library { return linkableLibraries(s.list.Projects(), project) }).Place(s.w, s.h)
	s.pkg = &p
}

// afterPackages acts on the package prompt's outcome: the command it shows runs.
func (s *Shell) afterPackages(p packageprompt.Prompt, out dialog.Outcome) tea.Cmd {
	if !out.Done {
		return nil
	}
	s.pkg = nil
	if out.Confirmed {
		// Removing is not undone by a click: the confirmation shows the command
		// and what it touches. An install was already asked for in the prompt.
		return s.execute(p.Argv(), !p.Removes())
	}
	return nil
}

// afterReleasePick acts on the outcome of the hold and bump question: the
// confirmation of the release, with the line that now holds and bumps, comes next.
func (s *Shell) afterReleasePick(m releasepick.Model, out dialog.Outcome) tea.Cmd {
	if !out.Done {
		return nil
	}
	s.relpick = nil
	if out.Confirmed {
		s.askConfirm(m.Argv())
		return nil
	}
	s.say(words.Say("tui.keys.cancelled_nothing_was_released"))
	return nil
}

// afterConfirm acts on the confirmation dialog's outcome.
func (s *Shell) afterConfirm(c dialog.Confirm, out dialog.Outcome) tea.Cmd {
	if !out.Done {
		return nil
	}
	s.confirm = nil
	if c.Info {
		return s.leaveForm()
	}
	if c.Working() {
		// Esc on a dialog that waits for its plan: stop the work, keep nothing.
		note := s.waitNote()
		s.cancelWaiting()
		if note != "" {
			s.say(note)
			return s.leaveForm()
		}
	}
	if isFirstCommit(c.Argv) {
		if out.Confirmed {
			return s.firstCommit()
		}
		s.say(words.Say("tui.release_first.cancelled"))
		return nil
	}
	if out.Confirmed {
		s.addForm = nil
		return s.execute(c.Final(), true)
	}
	if s.addForm != nil && len(c.Argv) > 0 && c.Argv[0] == "add" {
		// Esc on the plan of an add goes back one question, not out of the form.
		back := s.addForm.Reopen()
		s.addForm = nil
		s.guide = back
		return s.gotoTab(commands.TabGuide)
	}
	s.say(words.Say("tui.keys.cancelled_was_not_run", "line", commands.Line(c.Argv)))
	return s.leaveForm()
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
		s.say(words.Say("tui.keys.is_unavailable", "title", c.Title, "reason", reason))
		return nil
	}
	if c.Kind == commands.PaneKey && len(c.Keys) > 0 {
		// A key the view handles itself (a filter button of Doctor): a click is that key.
		return s.paneKey(pressedKey(c.Keys[0]))
	}
	return s.perform(c)
}

// activateSelection runs the actions of the detail that depend on what is
// selected.
func (s *Shell) activateSelection(action string) tea.Cmd {
	project := s.selectedProject()
	if id, ok := strings.CutPrefix(action, "form:"); ok {
		// A check of the Doctor tab run now: the form the palette opens.
		return s.openFormFor(id, s.extraAnswers(id))
	}
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
	case "why-file":
		return s.whyFile()
	case "why-miss":
		return s.whyMiss()
	case "derive-explain", "derive-apply", "derive-adopt", "derive-eject", "derive-replace":
		p, ok := s.list.Selected()
		if !ok {
			return nil
		}
		if argv := p.DeriveArgv(strings.TrimPrefix(action, "derive-")); argv != nil {
			return s.execute(argv, false)
		}
		s.say(words.Say("tui.keys.no_file_of_needs_that", "name", p.Name))
	case "rerun":
		if argv := s.repeatArgv(); argv != nil {
			return s.execute(argv, false)
		}
	case "rerun-failed":
		if argv := s.failedArgv(); argv != nil {
			return s.execute(argv, false)
		}
		s.say(words.Say("tui.keys.no_task_failed_or_was_skipped"))
	case "copy-log":
		s.copyLog()
	case "run-older", "run-newer":
		s.stepRun(map[string]int{"run-older": -1, "run-newer": 1}[action])
	case "run-back":
		s.closeRun()
	case "why-failed":
		// The log of the task that failed, whatever the run panel had selected.
		if s.run == nil {
			return nil
		}
		s.runSel = s.firstFailed()
		return s.revealRun()
	case "graph":
		if project != "" {
			return s.execute([]string{"graph", "--focus", project}, false)
		}
	case "release-why":
		s.releaseWhy()
	case "release-first":
		s.askFirstCommit()
	case "release-notes":
		if s.release.CanRelease() {
			return s.execute([]string{"release", "--dry-run", "--notes"}, false)
		}
		s.say(words.Say("tui.keys.there_is_no_release_to_show"))
	case "doctor-fix":
		if p, ok := s.doctor.SelectedFix(); ok {
			return s.openFix(p)
		}
		s.say(words.Say("tui.keys.this_check_has_no_fix_its"))
	case "deps-up":
		if argv := s.deps.UpArgv(); argv != nil {
			return s.execute(argv, false)
		}
		s.say(words.Say("tui.keys.no_dependency_to_update"))
	case "deps-install":
		return s.installDependency()
	case "deps-remove":
		return s.removeDependency()
	case "deps-open-page":
		return s.openRegistryPage()
	case "deps-copy-page":
		if r, ok := s.deps.Selected(); ok && r.URL != "" {
			s.copyText(r.URL)
		} else if ok {
			s.say(words.Say("tui.registry.unknown", "name", r.Name))
		} else {
			s.say(words.Say("tui.registry.none_selected"))
		}
	case "deps-align":
		if argv := s.deps.AlignArgv(); argv != nil {
			return s.execute(argv, false)
		}
		s.say(words.Say("tui.keys.the_one_version_policy_found_nothing"))
	case "reload":
		return s.reload()
	case "detail-zoom":
		s.zoomDetail()
	case "split-left":
		s.moveSplit(-1)
	case "split-right":
		s.moveSplit(1)
	case "home-up", "home-down":
		s.pageHome(map[string]int{"home-up": -1, "home-down": 1}[action])
	case "detail-up", "detail-down":
		if !s.detailVisible() {
			s.say(words.Say("tui.keys.the_detail_is_not_on_screen", "shortcutOf", commands.ShortcutOf("detail.zoom")))
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
		if c.Guide == addProjectTitle() {
			return s.openAddProject()
		}
		if c.Guide == swapTitle() {
			return s.openSwap()
		}
		if c.Guide == removeProjectTitle() {
			return s.openRemoveProject()
		}
		if c.Guide == guide.EnvironmentsAddTitle || c.Guide == guide.EnvironmentsRemoveTitle {
			return s.openEnvironments(c.Guide)
		}
		if c.Guide == guide.ForcedReleaseTitle {
			return s.openForcedRelease()
		}
		if c.Guide == recipeTitle() {
			// The recipe form lists the plugins of each level: read them first.
			return s.openRecipeForm()
		}
		if g, ok := s.guide.WithDefault("lang", s.mainLanguage()).Pick(c.Guide); ok {
			if g.Done() && len(g.Argv) > 0 {
				// A form with no question (smith boundaries) runs at once.
				s.guide = g.Reset()
				return s.execute(g.Argv, false)
			}
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
			s.say(words.Say("tui.keys.following_the_log"))
		} else {
			s.say(words.Say("tui.keys.log_paused_scroll_or_press_to", "shortcutOf", commands.ShortcutOf("run.follow")))
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
	case commands.OpenCompletion:
		return s.openCompletion()
	case commands.DeclineCompletion:
		s.declineCompletion()
	case commands.OpenDoctor:
		return s.gotoTab(commands.TabDoctor)
	case commands.OpenCompose:
		return s.openCompose()
	case commands.OpenCacheSetup:
		return s.openCacheSetup()
	case commands.OpenAdoptWizard:
		return s.openAdopt()
	case commands.OpenFixWizard:
		if p, ok := s.doctor.FixAll(); ok {
			return s.openFix(p)
		}
		s.say(words.Say("tui.keys.doctor_found_nothing_it_can_fix"))
	}
	return nil
}

func (s *Shell) openHelp() {
	// The keys of the open view come first, then the keys every view shares,
	// all read from the keymap the keys themselves come from.
	var rows []helpdialog.HelpRow
	for _, r := range commands.HelpRows(s.state(), tabTitles()[s.tab]) {
		rows = append(rows, helpdialog.HelpRow(r))
	}
	h := helpdialog.NewHelp(rows, mouseHelp()).Place(s.w, s.h)
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
		s.pluginUndo = nil
	}
	if id == commands.TabGuide && s.tab != commands.TabGuide && s.inWorkspace {
		// A form opened from a tab gives the person back to it when it ends or is cancelled.
		s.formOrigin = s.tab
	}
	s.tab, s.detailOpen, s.zoom = id, false, false
	s.selection = s.selection.Clear()
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
		if !s.release.Loaded {
			return s.loadRelease()
		}
	case commands.TabDeps:
		if !s.deps.Loaded && s.sources.Deps != nil {
			load := s.sources.Deps
			s.depsLoading = true
			return func() tea.Msg { return load() }
		}
	}
	return nil
}

// leaveForm brings the person back to the tab a guided form was opened from,
// after the form or the confirmation it led to was cancelled. It does nothing
// when no form is open from a tab, or when the person is somewhere else.
func (s *Shell) leaveForm() tea.Cmd {
	if s.tab != commands.TabGuide || s.formOrigin == "" || s.guide.Title() != "" {
		return nil
	}
	return s.leaveFormTo(commands.TabHome)
}

// leaveFormTo shows the tab the open form came from, or fallback when it came
// from none.
func (s *Shell) leaveFormTo(fallback string) tea.Cmd {
	to := fallback
	if s.formOrigin != "" && s.canShow(s.formOrigin) && s.formOrigin != commands.TabGuide {
		to = s.formOrigin
	}
	s.formOrigin = ""
	return s.gotoTab(to)
}

// paneKey sends a key to the open view.
func (s *Shell) paneKey(msg tea.KeyMsg) tea.Cmd {
	name := msg.String()
	switch s.tab {
	case commands.TabHome:
		return s.homeKey(name)
	case commands.TabProjects:
		switch {
		case s.runShown() && s.sideKeys():
			return s.runKey(name)
		case s.recentFocus && !s.detailFull():
			return s.recentKey(name)
		case s.sideKeys():
			return s.detailKey(name)
		}
		return s.projectsKey(name)
	case commands.TabDeps:
		s.deps = s.deps.Update(msg, s.listPage())
	case commands.TabRelease:
		s.release = s.release.Update(msg)
	case commands.TabSettings:
		return s.settingsKey(msg)
	case commands.TabDoctor:
		switch {
		case s.sideKeys():
			return s.detailKey(name)
		case (name == "tab" || (name == "right" && !s.doctor.OnHeader())) && s.beside():
			s.focusSide()
		case commands.NavNamed(name) == commands.NavConfirm && s.onDoctorExtra():
			// Enter on a check doctor does not run opens the form that runs it.
			cmd, _ := s.runSelectedExtra()
			return cmd
		case commands.NavNamed(name) == commands.NavConfirm && s.lay().HasDetail() && !s.doctor.OnPassedLine() && !s.doctor.OnHeader():
			// Beside the detail Enter unfolds what the detail folded. On the line
			// that hides the passed checks it shows them, as that line says.
			s.detailAll = !s.detailAll
		default:
			s.doctor = s.doctor.Update(msg)
		}
	case commands.TabGuide:
		g, cmd := s.guide.Handle(msg)
		if s.guide.Title() == "" && g.Title() == recipeTitle() && s.recipeLevels == nil {
			// The recipe form lists the plugins of each level: read them first.
			return s.openRecipeForm()
		}
		s.guide = g
		if g.Done() {
			argv, wizard := g.Argv, g.Wizard
			s.Form, s.Answers = g.Title(), g.Answers()
			if len(argv) > 0 && argv[0] == "add" && s.inWorkspace {
				kept := g
				s.addForm = &kept
			}
			s.guide = g.Reset()
			if wizard == guide.WizardAdopt {
				return s.openAdopt()
			}
			if wizard == guide.WizardFixConfig && s.repair != nil {
				return s.openFix(s.repair.Fix)
			}
			if wizard == guide.WizardRecipe {
				// "Make my own recipe" in Add a project: the Write a recipe form.
				return s.openRecipeForm()
			}
			if len(argv) >= 2 && argv[0] == "plugin" && argv[1] == "search" {
				// The hits come back as a form that installs the one picked.
				return s.openPluginSearch(strings.Join(argv[2:], " "))
			}
			if len(argv) > 0 && argv[0] == "swap" && s.inWorkspace {
				// The plan comes first: what leaves, what joins, what is left by hand.
				return s.swapPreview(argv)
			}
			if len(argv) > 0 && argv[0] == "add" && s.inWorkspace {
				// The plan comes first: what would be written, and which sibling it follows.
				return s.addPreview(argv)
			}
			if len(argv) > 0 {
				return s.execute(argv, false)
			}
			if s.inWorkspace {
				return s.leaveFormTo(commands.TabHome)
			}
		}
		if s.inWorkspace && g.Title() == "" {
			// Inside a workspace the guide is only its forms: backing out of
			// the first question leaves, there is no menu to land on.
			s.guide = g.Reset()
			return s.leaveFormTo(commands.TabHome)
		}
		return cmd
	}
	return nil
}

// projectsKey moves in the list of projects. Enter opens the detail over the
// list when it is not beside it.
func (s *Shell) projectsKey(name string) tea.Cmd {
	switch commands.NavNamed(name) {
	case commands.NavUp:
		s.list = s.list.Move(-1)
	case commands.NavDown:
		s.list = s.list.Move(1)
	case commands.NavPageUp:
		s.list = s.list.Page(-1, s.listPage())
	case commands.NavPageDown:
		s.list = s.list.Page(1, s.listPage())
	case commands.NavHome:
		s.list = s.list.Home()
	case commands.NavEnd:
		s.list = s.list.End()
	case commands.NavConfirm:
		if s.lay().Mode == layout.Narrow && s.list.Len() > 0 {
			s.runOpen, s.detailOpen, s.detailCursor = false, true, 0
		}
	case commands.NavNext:
		if len(s.runs) > 0 {
			s.recentFocus = true
		} else {
			s.focusSide()
		}
	default:
		if name == "right" {
			s.focusSide()
		}
	}
	return nil
}

// recentKey moves in the list of recent runs under the projects: arrows choose,
// Enter opens the run in the right panel, Esc or Left give the keyboard back to
// the projects, and Tab moves it on to the right panel (or back to the projects
// when none is beside the list). The list is open while it has the keyboard.
func (s *Shell) recentKey(name string) tea.Cmd {
	switch act := commands.NavNamed(name); {
	case act == commands.NavUp:
		s.recentAt = runhistory.StepRecent(s.recentAt, len(s.runs), -1)
	case act == commands.NavDown:
		s.recentAt = runhistory.StepRecent(s.recentAt, len(s.runs), 1)
	case act == commands.NavConfirm:
		return s.openRecent()
	case act == commands.NavNext || name == "right":
		s.recentFocus = false
		s.focusSide()
	case act == commands.NavCancel || name == "left":
		s.recentFocus = false
	}
	return nil
}

// detailKey moves between the actions of the detail (it fills the pane, or the
// keyboard moved to it beside the list), runs the highlighted one, and goes back
// to the list.
func (s *Shell) detailKey(name string) tea.Cmd {
	actions := s.detailActions()
	act := commands.NavNamed(name)
	switch {
	case !s.detailFull() && (act == commands.NavCancel || act == commands.NavNext || name == "left" || name == "backspace"):
		s.focusList()
	case act == commands.NavCancel || name == "left" || name == "backspace":
		s.detailOpen, s.zoom = false, false
	case act == commands.NavPageUp:
		s.pageDetail(-1)
	case act == commands.NavPageDown:
		s.pageDetail(1)
	case act == commands.NavUp:
		s.detailCursor = max(s.detailCursor-1, 0)
		s.keepDetailCursorVisible()
	case act == commands.NavDown:
		s.detailCursor = min(s.detailCursor+1, max(len(actions)-1, 0))
		s.keepDetailCursorVisible()
	case act == commands.NavConfirm:
		if s.detailCursor < len(actions) && !actions[s.detailCursor].Disabled {
			return s.activate(actions[s.detailCursor].ID)
		}
	}
	return nil
}

// runKey moves in the run panel: the task list and the log.
func (s *Shell) runKey(name string) tea.Cmd {
	if s.run == nil {
		return nil
	}
	ids := s.runIDs()
	if name == "left" || commands.NavNamed(name) == commands.NavNext {
		// Beside the list, the keyboard goes back to the list; the run stays open.
		if s.beside() {
			s.focusList()
		}
		return nil
	}
	switch commands.NavNamed(name) {
	case commands.NavUp:
		s.runSel = runhistory.StepSelection(ids, s.runSel, -1)
	case commands.NavDown:
		s.runSel = runhistory.StepSelection(ids, s.runSel, 1)
	case commands.NavCancel:
		// The first Esc shows all output; the next goes back to the project.
		if s.runSel == "" {
			s.closeRun()
			return nil
		}
		s.runSel = ""
	case commands.NavPageUp:
		s.log = s.log.Page(-1, s.logTotal(), s.logHeight())
	case commands.NavPageDown:
		s.log = s.log.Page(1, s.logTotal(), s.logHeight())
	case commands.NavHome:
		s.log = s.log.Home()
	case commands.NavEnd:
		s.log = s.log.End()
	}
	s.keepRunVisible()
	return nil
}

// runIDs are the choices in the run panel's list: "" (all output), then the
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
		// The release item of Home is read again too, in the background.
		s.release = releaseview.Model{}
		cmds = append(cmds, s.loadRelease())
		return tea.Batch(cmds...)
	case commands.TabDoctor:
		if s.doctor.Report != nil && s.sources.Doctor != nil {
			// The list stays, with its folds, filters and cursor, until the new one replaces it.
			if !s.doctor.Loading {
				s.doctor.Loading, s.rechecking = true, true
				load := s.sources.Doctor
				return func() tea.Msg { return load() }
			}
			return nil
		}
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
