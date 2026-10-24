package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/palette"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseview"
)

// mouseHelp are the mouse actions listed in the shortcuts overlay.
var mouseHelp = []string{
	"Click: select   Double-click: open the detail (narrow terminals)",
	"Click a tab to open the view, a hint in the footer to press its key",
	"Click an action in the detail to do it",
	"Wheel: scroll lists and logs",
	"Shift+drag selects text (Option+drag in macOS terminals)",
	"m turns the mouse off, so a plain drag selects text",
}

// editing is true while a text field has the keyboard.
func (s Shell) editing() bool {
	return (s.tab == commands.TabGuide && s.guide.Editing()) || (s.tab == commands.TabSettings && s.settings.Editing) || (s.tab == commands.TabRecipes && s.recipes.Editing())
}

// formOpen is true while a guided form is open: its keys are the form's.
func (s Shell) formOpen() bool { return s.tab == commands.TabGuide && s.guide.Title() != "" }

func (s *Shell) key(msg tea.KeyMsg) tea.Cmd {
	name := msg.String()
	switch {
	case s.confirm != nil:
		c, out := s.confirm.Key(name)
		s.confirm = &c
		return s.afterConfirm(c, out)
	case s.about != nil:
		a, out := s.about.Key(name)
		s.about = &a
		if out.Done {
			s.about = nil
		}
		return nil
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

// afterConfirm acts on the confirmation dialog's outcome.
func (s *Shell) afterConfirm(c dialog.Confirm, out dialog.Outcome) tea.Cmd {
	if !out.Done {
		return nil
	}
	s.confirm = nil
	if out.Confirmed {
		return s.execute(c.Argv, true)
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
	switch {
	case res.CommandID != "":
		return s.activate(res.CommandID)
	case len(res.Argv) > 0:
		return s.execute(res.Argv, false)
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
	case "explain":
		if project != "" {
			return s.execute([]string{"explain", project}, false)
		}
	case "rerun":
		if s.lastArg != nil {
			return s.execute(s.lastArg, false)
		}
	case "projects":
		return s.gotoTab(commands.TabProjects)
	case "log":
		s.runSel = ""
		return s.gotoTab(commands.TabRun)
	case "graph":
		if project != "" {
			return s.execute([]string{"graph", "--focus", project}, false)
		}
	case "plugin-toggle":
		if argv := s.plugins.ToggleArgv(); argv != nil {
			return s.execute(argv, false)
		}
		s.say("Only the built-in language plugins can be turned on or off here")
	case "recipes-create":
		argv, reason := s.recipes.Argv()
		if argv == nil {
			s.say(reason)
			return nil
		}
		return s.execute(argv, false)
	case "release-notes":
		if s.release.CanRelease() {
			return s.execute([]string{"release", "--dry-run", "--notes"}, false)
		}
		s.say("There is no release to show the notes of")
	case "doctor-fix":
		if argv := s.doctor.FixArgv(); argv != nil {
			return s.execute(argv, false)
		}
		s.say("This finding has no safe fix doctor can apply")
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
		project := s.selectedProject()
		if c.All {
			project = ""
		}
		return s.execute(commands.RunArgv(c.Task, project), false)
	case commands.Guided:
		if g, ok := s.guide.Pick(c.Guide); ok {
			s.guide = g
		}
		return s.gotoTab(commands.TabGuide)
	case commands.Palette:
		s.palOpen = true
		s.pal = s.pal.Reset()
	case commands.Help:
		s.openHelp()
	case commands.About:
		a := dialog.NewAbout(s.version).Place(s.w, s.h)
		s.about = &a
	case commands.Quit:
		if s.proc != nil {
			s.proc.Stop()
		}
		s.quit = true
	case commands.Stop:
		s.stop()
	case commands.Follow:
		s.log = s.log.Toggle(len(s.logLines()), s.logHeight())
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
	case commands.ReloadDoctor:
		s.doctor = doctorview.Model{}
		return s.gotoTab(commands.TabDoctor)
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
	if !contains(s.tabs, id) {
		return nil
	}
	if id != s.tab {
		// A result toast belongs to the screen it was raised on.
		s.toast = s.toast.Dismiss()
	}
	s.tab, s.detailOpen = id, false
	switch id {
	case commands.TabDoctor:
		if s.doctor.Report == nil && s.doctor.Err == nil && s.sources.Doctor != nil {
			load := s.sources.Doctor
			return func() tea.Msg { return load() }
		}
	case commands.TabRelease:
		if !s.release.Loaded && s.sources.Release != nil {
			load := s.sources.Release
			return func() tea.Msg { return load() }
		}
	case commands.TabPlugins:
		if !s.plugins.Loaded && s.sources.Plugins != nil {
			load := s.sources.Plugins
			return func() tea.Msg { return load() }
		}
	case commands.TabDeps:
		if !s.deps.Loaded && s.sources.Deps != nil {
			load := s.sources.Deps
			return func() tea.Msg { return load() }
		}
	case commands.TabRecipes:
		if !s.recipes.Loaded && s.sources.Recipes != nil {
			load := s.sources.Recipes
			return func() tea.Msg { return load() }
		}
	case commands.TabSettings:
		if !s.settings.Loaded && s.sources.Settings != nil {
			load := s.sources.Settings
			return func() tea.Msg { return load() }
		}
	}
	return nil
}

// paneKey sends a key to the open view.
func (s *Shell) paneKey(msg tea.KeyMsg) tea.Cmd {
	name := msg.String()
	switch s.tab {
	case commands.TabProjects:
		if s.detailOpen {
			return s.detailKey(name)
		}
		return s.projectsKey(name)
	case commands.TabPlugins:
		s.plugins = s.plugins.Update(msg, s.listPage())
	case commands.TabDeps:
		s.deps = s.deps.Update(msg, s.listPage())
	case commands.TabRun:
		return s.runKey(name)
	case commands.TabRecipes:
		s.recipes = s.recipes.Update(msg)
	case commands.TabRelease:
		s.release = s.release.Update(msg)
	case commands.TabDoctor:
		s.doctor = s.doctor.Update(msg)
	case commands.TabSettings:
		next, argv := s.settings.Update(msg)
		s.settings = next
		if argv != nil {
			return s.execute(argv, false)
		}
	case commands.TabGuide:
		g, cmd := s.guide.Handle(msg)
		s.guide = g
		if g.Done() {
			argv := g.Argv
			s.guide = g.Reset()
			if len(argv) > 0 {
				return s.execute(argv, false)
			}
			if s.inWorkspace {
				return s.gotoTab(commands.TabProjects)
			}
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
		s.detailOpen = false
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
		s.runSel = ""
	case "pgup":
		s.log = s.log.Page(-1, len(s.logLines()), s.logHeight())
	case "pgdown":
		s.log = s.log.Page(1, len(s.logLines()), s.logHeight())
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

// listPage is how many rows a page key moves in a list view.
func (s *Shell) listPage() int { return max(s.mainInner().H-6, 1) }

// reload forgets what the current view loaded and loads it again; nil for a
// view that has nothing to reload.
func (s *Shell) reload() tea.Cmd {
	switch s.tab {
	case commands.TabDoctor:
		s.doctor = doctorview.Model{}
	case commands.TabRelease:
		s.release = releaseview.Model{}
	case commands.TabPlugins:
		s.plugins = s.plugins.Reset()
	case commands.TabDeps:
		s.deps = depsview.Model{}
	case commands.TabRecipes:
		s.recipes = s.recipes.Reset()
	default:
		return nil
	}
	return s.gotoTab(s.tab)
}
