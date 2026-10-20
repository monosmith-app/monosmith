package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/navigator"
	"github.com/monosmith-app/monosmith/apps/smith/tui/palette"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseview"
)

// globalWhileEditing are the shortcuts that still work while a text field has
// the keyboard; the others (Ctrl+A, Ctrl+F…) are the field's own.
var globalWhileEditing = map[string]bool{
	"ctrl+q": true, "ctrl+p": true, "f1": true, "alt+1": true, "alt+2": true, "alt+3": true, "alt+4": true, "alt+5": true, "alt+6": true, "alt+7": true, "alt+8": true, "alt+9": true,
}

// mouseHelp are the mouse actions listed in the shortcuts overlay.
var mouseHelp = []string{
	"Click: select or focus   Double-click: run the task",
	"Right-click: context menu   Wheel: scroll lists and logs",
	"Click a tab, a menu, a button or a menu item",
	"Hover a menu to switch menus; hover a disabled item for why",
}

// otherKeys are keys that are not commands of their own.
var otherKeys = []dialog.HelpRow{
	{Group: "Moving", Title: "Move in a list or tree", Shortcut: "Up Down"},
	{Group: "Moving", Title: "Open or close a branch", Shortcut: "Right Left"},
	{Group: "Moving", Title: "Run the task, or open the branch", Shortcut: "Enter"},
	{Group: "Moving", Title: "Open the menu bar", Shortcut: "F10 / Alt+letter"},
	{Group: "Moving", Title: "Back, close, clear", Shortcut: "Esc"},
	{Group: "Moving", Title: "Reload the view (Doctor, Release, Plugins, Deps)", Shortcut: "r"},
	{Group: "Moving", Title: "Show or hide this help", Shortcut: "? / F1"},
}

// tabKeys are the keys of one view, shown in the help while it is open.
var tabKeys = map[string][]dialog.HelpRow{
	commands.TabProjects: {
		{Group: "Projects view", Title: "Run the selected task", Shortcut: "Enter / Ctrl+R"},
		{Group: "Projects view", Title: "Build, test, lint, typecheck", Shortcut: "Ctrl+B/T/L/Y"},
		{Group: "Projects view", Title: "Run lint and test on what changed", Shortcut: "Ctrl+A"},
		{Group: "Projects view", Title: "Graph around the project", Shortcut: "Inspector"},
	},
	commands.TabDoctor: {
		{Group: "Doctor view", Title: "Open the detail, remedy and fix of a finding", Shortcut: "Enter"},
		{Group: "Doctor view", Title: "Apply this finding's safe fix (smith doctor --fix --only)", Shortcut: "f"},
		{Group: "Doctor view", Title: "Group the findings by project, or list them flat", Shortcut: "g"},
		{Group: "Doctor view", Title: "Check again", Shortcut: "r / Ctrl+D"},
	},
	commands.TabRelease: {
		{Group: "Release view", Title: "Show the commits behind the selected bump", Shortcut: "Enter"},
		{Group: "Release view", Title: "Print the release notes (smith release --dry-run --notes)", Shortcut: "n"},
		{Group: "Release view", Title: "Preview, or release after confirming", Shortcut: "F2 / F3"},
	},
	commands.TabRecipes: {
		{Group: "Recipes view", Title: "Change the field: mode, kind, a choice, the recipe", Shortcut: "Left Right / Enter"},
		{Group: "Recipes view", Title: "Type a name or a path", Shortcut: "Enter, then Enter / Esc"},
		{Group: "Recipes view", Title: "Create (smith add, smith new, smith recipe new)", Shortcut: "c"},
		{Group: "Recipes view", Title: "Read the recipes again", Shortcut: "r"},
	},
	commands.TabPlugins: {
		{Group: "Plugins view", Title: "Turn the plugin on or off (smith plugin enable|disable)", Shortcut: "e / Enter"},
		{Group: "Plugins view", Title: "List the plugins again", Shortcut: "r"},
	},
	commands.TabDeps: {
		{Group: "Deps view", Title: "Mark or unmark a dependency", Shortcut: "Space"},
		{Group: "Deps view", Title: "Mark all, or none", Shortcut: "a"},
		{Group: "Deps view", Title: "Update the marked ones (smith deps up)", Shortcut: "u / Enter"},
		{Group: "Deps view", Title: "Bring drifting versions to one (--align)", Shortcut: "v"},
		{Group: "Deps view", Title: "Look for releases again", Shortcut: "r"},
	},
}

func (s *Shell) editing() bool {
	return (s.tab == commands.TabGuide && s.guide.Editing()) || (s.tab == commands.TabSettings && s.settings.Editing) || (s.tab == commands.TabRecipes && s.recipes.Editing())
}

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
	case s.popup.Shown():
		p, res := s.popup.Key(name)
		s.popup = p
		return s.afterMenu(res.Chosen, res.Refused)
	}
	if (name == "ctrl+q" || name == "ctrl+c") && s.bar.Active() {
		s.bar = s.bar.Close()
	}
	if b, res := s.bar.Key(name); res.Handled {
		s.bar = b
		return s.afterMenu(res.Chosen, res.Refused)
	}
	if name == "ctrl+c" {
		if s.running() {
			s.stop()
		} else {
			s.quit = true
		}
		return nil
	}
	if c, reason, ok := commands.ByKey(name, s.state()); ok && (!s.editing() || globalWhileEditing[name]) {
		if reason != "" {
			s.say(c.Title + " is unavailable: " + reason)
			return nil
		}
		return s.perform(c)
	}
	return s.paneKey(msg)
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

// afterMenu runs the command a menu chose, or says why it could not.
func (s *Shell) afterMenu(chosen, refused string) tea.Cmd {
	if refused != "" {
		s.say(refused)
	}
	if chosen != "" {
		return s.activate(chosen)
	}
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

// activate runs a command by id: a menu entry, a button, an inspector action.
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

// activateSelection runs the actions of the inspector and context menus that
// depend on what is selected.
func (s *Shell) activateSelection(action string) tea.Cmd {
	sel := s.tree.Selected()
	switch action {
	case "explain":
		if sel.Task != "" {
			return s.execute([]string{"explain", sel.Project + ":" + sel.Task}, false)
		}
		if sel.Project != "" {
			return s.execute([]string{"explain", sel.Project}, false)
		}
	case "why-miss":
		if sel.Task != "" {
			return s.execute([]string{"why-miss", sel.Project + ":" + sel.Task}, false)
		}
	case "rerun":
		if s.lastArg != nil {
			return s.execute(s.lastArg, false)
		}
	case "toggle":
		s.tree = s.tree.Toggle()
	case "projects":
		return s.gotoTab(commands.TabProjects)
	case "log":
		s.runSel = ""
		return s.gotoTab(commands.TabRun)
	case "graph":
		if sel.Project != "" {
			return s.execute([]string{"graph", "--focus", sel.Project}, false)
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
		return s.execute(commands.RunArgv(c.Task, s.selectedProject()), false)
	case commands.RunSelection:
		if sel := s.tree.Selected(); sel.Kind == navigator.TaskRow {
			return s.execute(commands.RunArgv(sel.Task, sel.Project), false)
		}
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
			s.say("Log paused: scroll or press Ctrl+F to follow again")
		}
	case commands.NextPane:
		s.cyclePane(1)
	case commands.PrevPane:
		s.cyclePane(-1)
	case commands.ContextMenu:
		s.bar = s.bar.Close()
		if s.focus == layout.Navigator {
			in := s.navigatorInner()
			s.openContextMenu(in.X+2, in.Y+s.tree.CursorLine())
		} else {
			in := s.mainInner()
			s.openContextMenu(in.X+2, in.Y+1)
		}
	case commands.ReloadDoctor:
		s.doctor = doctorview.Model{}
		return s.gotoTab(commands.TabDoctor)
	}
	return nil
}

func (s *Shell) openHelp() {
	// The keys of the open view come first, then the menus' shortcuts, then
	// the keys every view shares.
	rows := append([]dialog.HelpRow(nil), tabKeys[s.tab]...)
	for _, c := range commands.Shortcuts(s.state()) {
		rows = append(rows, dialog.HelpRow{Group: c.Menu, Title: c.Title, Shortcut: c.Shortcut})
	}
	rows = append(rows, otherKeys...)
	h := dialog.NewHelp(rows, mouseHelp).Place(s.w, s.h)
	s.help = &h
}

// cyclePane moves the keyboard focus to the next visible pane.
func (s *Shell) cyclePane(dir int) {
	panes := []layout.Area{layout.Main}
	if s.lay.HasNavigator() && s.inWorkspace {
		panes = append([]layout.Area{layout.Navigator}, panes...)
	}
	if s.lay.HasInspector() && s.inWorkspace {
		panes = append(panes, layout.Inspector)
	}
	for i, a := range panes {
		if a == s.focus {
			s.focus = panes[((i+dir)%len(panes)+len(panes))%len(panes)]
			return
		}
	}
	s.focus = layout.Main
}

// gotoTab shows a view and loads it the first time.
func (s *Shell) gotoTab(id string) tea.Cmd {
	if !contains(s.tabs, id) {
		return nil
	}
	s.tab, s.focus = id, layout.Main
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

// paneKey sends a key to the focused pane.
func (s *Shell) paneKey(msg tea.KeyMsg) tea.Cmd {
	name := msg.String()
	switch {
	case s.focus == layout.Navigator || s.tab == tabNavigator:
		return s.navigatorKey(name)
	case s.focus == layout.Inspector:
		return s.inspectorKey(name)
	}
	if name == "r" && !s.editing() {
		if cmd := s.reload(); cmd != nil {
			return cmd
		}
	}
	switch s.tab {
	case commands.TabProjects:
		return s.navigatorKey(name)
	case commands.TabPlugins:
		if name == "e" || name == "enter" {
			return s.activateSelection("plugin-toggle")
		}
		s.plugins = s.plugins.Update(msg, s.listPage())
	case commands.TabDeps:
		switch name {
		case "u", "enter":
			return s.activateSelection("deps-up")
		case "v":
			return s.activateSelection("deps-align")
		}
		s.deps = s.deps.Update(msg, s.listPage())
	case commands.TabRun:
		return s.runKey(name)
	case commands.TabRecipes:
		if name == "c" && !s.recipes.Editing() {
			return s.activateSelection("recipes-create")
		}
		s.recipes = s.recipes.Update(msg)
	case commands.TabRelease:
		if name == "n" {
			return s.activateSelection("release-notes")
		}
		s.release = s.release.Update(msg)
	case commands.TabDoctor:
		if name == "f" {
			return s.activateSelection("doctor-fix")
		}
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

func (s *Shell) navigatorKey(name string) tea.Cmd {
	before := s.tree.Selected()
	switch name {
	case "up":
		s.tree = s.tree.Move(-1)
	case "down":
		s.tree = s.tree.Move(1)
	case "pgup":
		s.tree = s.tree.Page(-1)
	case "pgdown":
		s.tree = s.tree.Page(1)
	case "home":
		s.tree = s.tree.Home()
	case "end":
		s.tree = s.tree.End()
	case "left":
		s.tree = s.tree.Collapse()
	case "right":
		s.tree = s.tree.Expand()
	case " ":
		s.tree = s.tree.Toggle()
	case "enter":
		if sel := s.tree.Selected(); sel.Kind == navigator.TaskRow {
			return s.perform(mustCommand("run.selection"))
		}
		s.tree = s.tree.Toggle()
	}
	if s.tree.Selected() != before {
		s.inspCursor = 0
	}
	return nil
}

func mustCommand(id string) commands.Command {
	c, _ := commands.ByID(id)
	return c
}

func (s *Shell) inspectorKey(name string) tea.Cmd {
	actions := s.inspectorActions()
	switch name {
	case "up":
		s.inspCursor = max(s.inspCursor-1, 0)
	case "down":
		s.inspCursor = min(s.inspCursor+1, max(len(actions)-1, 0))
	case "enter":
		if s.inspCursor < len(actions) && !actions[s.inspCursor].Disabled {
			return s.activate(actions[s.inspCursor].ID)
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
