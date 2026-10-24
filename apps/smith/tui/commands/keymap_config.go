package commands

// The keymap: the one list the keys, the footer hints, the help (F1 and ?)
// and the command palette are all drawn from, so a label cannot go stale.
// Every key is a single key, nothing needs a modifier, and none works while a
// text field is editing (see ByKey).

// Group names, in the order of the help.
const (
	GroupMoving    = "Moving"
	GroupStart     = "Start"
	GroupRun       = "Run"
	GroupWorkspace = "Workspace"
	GroupProjects  = "Projects"
	GroupRelease   = "Release"
	GroupViews     = "Views"
	GroupHelp      = "Help"
)

// GroupOrder is the order of the groups in the help.
var GroupOrder = []string{GroupMoving, GroupStart, GroupRun, GroupWorkspace, GroupProjects, GroupRelease, GroupViews, GroupHelp}

// The views, as slices to scope a command to (see Command.Views).
var (
	vProjects = []string{TabProjects}
	vRun      = []string{TabRun}
	vDoctor   = []string{TabDoctor}
	vRelease  = []string{TabRelease}
	vPlugins  = []string{TabPlugins}
	vDeps     = []string{TabDeps}
	vRecipes  = []string{TabRecipes}
	vSettings = []string{TabSettings}
	vGuide    = []string{TabGuide}
	// vReloadable are the views that read something and can read it again.
	vReloadable = []string{TabDoctor, TabRelease, TabPlugins, TabDeps, TabRecipes}
	// vWorkspace are the views of a workspace where the run keys are shown.
	vWorkspace = []string{TabProjects, TabRun, TabDoctor, TabRelease, TabPlugins, TabDeps, TabRecipes, TabSettings}
	vCopy      = []string{TabProjects, TabRun, TabDoctor, TabDeps}
)

// all is the keymap; the order is the order within each group and in the footer.
var all = []Command{
	// The keys of each view's own pane come first: they win over a global key
	// that is the same letter (Deps marks all with a, which adds a project elsewhere).
	{ID: "projects.move", Group: GroupRun, Title: "Move in the list", Keys: []string{"down", "up", "j", "k"}, Shortcut: "Up Down / j k", Kind: PaneKey, Views: vProjects},
	{ID: "projects.open", Group: GroupRun, Title: "Open the detail of the project (narrow terminals)", Keys: []string{"enter"}, Kind: PaneKey, Views: vProjects, NeedsNarrowList: true, Footer: "detail", Glyph: "Enter"},
	{ID: "projects.back", Group: GroupRun, Title: "Back to the list", Keys: []string{"esc"}, Kind: PaneKey, Views: vProjects, NeedsDetailOpen: true, Footer: "back"},

	{ID: "run.again", Group: GroupRun, Title: "Run the last command again", Hint: "the last command line, as it was", Keys: []string{"r"}, Kind: ViewAction, Action: "rerun", Views: vRun, NeedsWorkspace: true, NeedsIdle: true, Footer: "again"},
	{ID: "run.follow", Group: GroupRun, Title: "Follow the log, or pause it", Hint: "stay at the end of the log, or pause to read", Keys: []string{"f"}, Kind: Follow, Views: vRun, Footer: "follow"},
	{ID: "run.all-output", Group: GroupRun, Title: "Show all output", Keys: []string{"esc"}, Kind: PaneKey, Views: vRun},
	{ID: "run.scroll", Group: GroupRun, Title: "Pick a task, scroll its log", Keys: []string{"down", "up", "pgdown", "pgup"}, Shortcut: "Up Down PgUp PgDn", Glyph: "↑↓", Kind: PaneKey, Views: vRun, Footer: "task"},

	{ID: "doctor.open", Group: GroupWorkspace, Title: "Open the detail, remedy and fix of a finding", Keys: []string{"enter"}, Kind: PaneKey, Views: vDoctor, Footer: "detail", Glyph: "Enter"},
	{ID: "doctor.fix", Group: GroupWorkspace, Title: "Apply this finding's safe fix (smith doctor --fix --only)", Keys: []string{"f"}, Kind: ViewAction, Action: "doctor-fix", Views: vDoctor, NeedsWorkspace: true, NeedsIdle: true, Footer: "fix"},
	{ID: "ws.fix", Group: GroupWorkspace, Title: "Fix with doctor…", Hint: "smith doctor --fix: apply the safe fixes", Keys: []string{"x"}, Kind: Execute, Argv: []string{"doctor", "--fix"}, Confirm: true, Views: vDoctor, NeedsWorkspace: true, NeedsIdle: true, Footer: "fix all"},
	{ID: "doctor.group", Group: GroupWorkspace, Title: "Group the findings by project, or list them flat", Keys: []string{"g"}, Kind: PaneKey, Views: vDoctor, Footer: "group"},
	{ID: "doctor.passed", Group: GroupWorkspace, Title: "Show or hide the passed checks", Keys: []string{"p"}, Kind: PaneKey, Views: vDoctor, Footer: "passed"},

	{ID: "release.open", Group: GroupRelease, Title: "Show the commits behind the selected bump", Keys: []string{"enter"}, Kind: PaneKey, Views: vRelease, Footer: "commits", Glyph: "Enter"},
	{ID: "rel.preview", Group: GroupRelease, Title: "Preview next release", Hint: "smith release --dry-run: changes nothing", Keys: []string{"p"}, Kind: Execute, Argv: []string{"release", "--dry-run"}, Views: vRelease, NeedsWorkspace: true, NeedsIdle: true, Footer: "preview"},
	{ID: "rel.notes", Group: GroupRelease, Title: "Print the release notes", Hint: "smith release --dry-run --notes", Keys: []string{"n"}, Kind: ViewAction, Action: "release-notes", Views: vRelease, NeedsWorkspace: true, NeedsIdle: true, Footer: "notes"},
	{ID: "rel.release", Group: GroupRelease, Title: "Release…", Hint: "smith release: version, publish and tag", Keys: []string{"x"}, Kind: Execute, Argv: []string{"release"}, Confirm: true, Views: vRelease, NeedsWorkspace: true, NeedsIdle: true, Footer: "release"},

	{ID: "plugins.toggle", Group: GroupProjects, Title: "Turn the plugin on or off", Hint: "smith plugin enable|disable", Keys: []string{"e", "enter"}, Kind: ViewAction, Action: "plugin-toggle", Views: vPlugins, NeedsWorkspace: true, NeedsIdle: true, Footer: "on/off"},

	{ID: "deps.mark", Group: GroupProjects, Title: "Mark or unmark a dependency", Keys: []string{" "}, Shortcut: "Space", Kind: PaneKey, Views: vDeps, Footer: "mark"},
	{ID: "deps.mark-all", Group: GroupProjects, Title: "Mark all, or none", Keys: []string{"a"}, Kind: PaneKey, Views: vDeps, Footer: "all"},
	{ID: "deps.update", Group: GroupProjects, Title: "Update the marked ones", Hint: "smith deps up", Keys: []string{"u", "enter"}, Kind: ViewAction, Action: "deps-up", Views: vDeps, NeedsWorkspace: true, NeedsIdle: true, Footer: "update"},
	{ID: "deps.align", Group: GroupProjects, Title: "Bring drifting versions to one", Hint: "smith deps up --align", Keys: []string{"v"}, Kind: ViewAction, Action: "deps-align", Views: vDeps, NeedsWorkspace: true, NeedsIdle: true, Footer: "align"},

	{ID: "recipes.field", Group: GroupProjects, Title: "Change the field: mode, kind, a choice, the recipe", Keys: []string{"left", "right"}, Shortcut: "Left Right / Enter", Glyph: "←→", Kind: PaneKey, Views: vRecipes, Footer: "field"},
	{ID: "recipes.create", Group: GroupProjects, Title: "Create (smith add, smith new, smith recipe new)", Keys: []string{"c"}, Kind: ViewAction, Action: "recipes-create", Views: vRecipes, NeedsWorkspace: true, NeedsIdle: true, Footer: "create"},

	{ID: "settings.edit", Group: GroupWorkspace, Title: "Edit the setting (true/false flips)", Keys: []string{"enter"}, Kind: PaneKey, Views: vSettings, Footer: "edit", Glyph: "Enter"},
	{ID: "settings.default", Group: GroupWorkspace, Title: "Back to the default (smith config unset)", Keys: []string{"ctrl+u", "delete"}, Shortcut: "Ctrl+U", Kind: PaneKey, Views: vSettings, Footer: "default"},

	{ID: "guide.choose", Group: GroupStart, Title: "Choose and go on", Keys: []string{"enter"}, Kind: PaneKey, Views: vGuide, Footer: "choose", Glyph: "Enter"},
	{ID: "guide.back", Group: GroupStart, Title: "Back one step", Keys: []string{"esc"}, Kind: PaneKey, Views: vGuide, Footer: "back"},

	{ID: "view.reload", Group: GroupViews, Title: "Read the view again", Keys: []string{"r"}, Kind: ViewAction, Action: "reload", Views: vReloadable, NeedsWorkspace: true, Footer: "reload"},

	// Moving: the keys every pane understands, handled by the pane itself.
	{ID: "move.arrows", Group: GroupMoving, Title: "Move in a list", Keys: []string{"down", "up", "j", "k"}, Shortcut: "Up Down / j k", Kind: PaneKey},
	{ID: "move.enter", Group: GroupMoving, Title: "Open or run the selection", Keys: []string{"enter"}, Kind: PaneKey},
	{ID: "move.esc", Group: GroupMoving, Title: "Back, close, clear (never quits)", Keys: []string{"esc"}, Kind: PaneKey},

	// Run
	{ID: "run.build", Group: GroupRun, Title: "Build", Hint: "smith build, on the selected project or all", Keys: []string{"b"}, Kind: RunTask, Task: "build", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"build"}, Footer: "build", ShowIn: vWorkspace},
	{ID: "run.test", Group: GroupRun, Title: "Test", Hint: "smith test, on the selected project or all", Keys: []string{"t"}, Kind: RunTask, Task: "test", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"test"}, Footer: "test", ShowIn: vWorkspace},
	{ID: "run.lint", Group: GroupRun, Title: "Lint", Hint: "smith lint, on the selected project or all", Keys: []string{"l"}, Kind: RunTask, Task: "lint", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"lint"}, Footer: "lint", ShowIn: vProjects},
	{ID: "run.typecheck", Group: GroupRun, Title: "Typecheck", Hint: "smith typecheck, on the selected project or all", Keys: []string{"c"}, Kind: RunTask, Task: "typecheck", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"typecheck"}, Footer: "typecheck", ShowIn: vProjects},
	{ID: "run.affected", Group: GroupRun, Title: "Lint and test what changed", Hint: "smith run lint test --affected", Keys: []string{"w"}, Kind: Execute, Argv: []string{"run", "lint", "test", "--affected"}, NeedsWorkspace: true, NeedsIdle: true, Footer: "changed", ShowIn: vProjects},
	{ID: "run.build-all", Group: GroupRun, Title: "Build all", Hint: "smith run build, on every project", Kind: RunTask, Task: "build", All: true, NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"build"}},
	{ID: "run.test-all", Group: GroupRun, Title: "Test all", Hint: "smith run test, on every project", Kind: RunTask, Task: "test", All: true, NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"test"}},
	{ID: "run.lint-all", Group: GroupRun, Title: "Lint all", Hint: "smith run lint, on every project", Kind: RunTask, Task: "lint", All: true, NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"lint"}},
	{ID: "run.stop", Group: GroupRun, Title: "Stop", Hint: "stops the run; pressing it again quits", Keys: []string{"ctrl+c"}, Kind: Stop, NeedsRunning: true, Footer: "stop", ShowIn: vWorkspace},

	// Workspace
	{ID: "ws.doctor", Group: GroupWorkspace, Title: "Doctor", Hint: "check the workspace; every problem names its remedy", Keys: []string{"d"}, Kind: ReloadDoctor, NeedsWorkspace: true, Presents: []string{"doctor"}, Footer: "doctor", ShowIn: vProjects},
	{ID: "ws.ci", Group: GroupWorkspace, Title: "Regenerate CI pipelines…", Hint: "smith ci generate", Kind: Execute, Argv: []string{"ci", "generate"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.info", Group: GroupWorkspace, Title: "Info", Hint: "smith info: the installed smith and the workspace's settings", Kind: Execute, Argv: []string{"info"}, NeedsIdle: true},

	// Start
	{ID: "file.new", Group: GroupStart, Title: "New workspace…", Hint: "smith new: create a monorepo", Keys: []string{"n"}, Kind: Guided, Guide: "Create a workspace", OnlyOutside: true, Footer: "new workspace", ShowIn: vGuide},
	{ID: "file.add", Group: GroupStart, Title: "Add project…", Hint: "smith add: an app, lib or package", Keys: []string{"a"}, Kind: Guided, Guide: "Add a project", NeedsWorkspace: true, NeedsIdle: true, OnlyInside: true, Footer: "add project", ShowIn: vProjects},
	{ID: "file.adopt", Group: GroupStart, Title: "Adopt this repository…", Hint: "smith adopt: bring an existing repository under smith", Kind: Guided, Guide: "Adopt this repository", NeedsIdle: true, OnlyOutside: true, NeedsContent: true},
	{ID: "file.convert", Group: GroupStart, Title: "Convert an Nx or Turborepo workspace…", Hint: "smith convert: turn it into a smith workspace", Kind: Guided, Guide: "Convert an Nx or Turborepo workspace", NeedsIdle: true, OnlyOutside: true, NeedsContent: true},

	// Projects
	{ID: "proj.install", Group: GroupProjects, Title: "Install dependencies…", Hint: "smith install: a fresh clone's npm install, venv, go mod download", Keys: []string{"i"}, Kind: Execute, Argv: []string{"install"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.affected", Group: GroupProjects, Title: "Show what changed", Hint: "smith affected: projects touched since the base branch", Kind: Execute, Argv: []string{"affected"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.graph", Group: GroupProjects, Title: "Project graph", Hint: "smith graph", Kind: Execute, Argv: []string{"graph"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.plugins", Group: GroupProjects, Title: "Manage plugins…", Hint: "choose which built-in plugins are on", Kind: Guided, Guide: "Manage plugins", NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.upgrade", Group: GroupProjects, Title: "Upgrade workspace…", Hint: "smith upgrade: migrate smith.toml, regenerate what smith owns", Kind: Execute, Argv: []string{"upgrade"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.remove", Group: GroupProjects, Title: "Remove project…", Hint: "smith remove: its directory, its traces in the workspace, its CI", Kind: Guided, Guide: "Remove a project", NeedsWorkspace: true, NeedsIdle: true},

	// Release

	// Views: the number is the position in the tab row.
	{ID: "view.projects", Group: GroupViews, Title: "Projects", Keys: []string{"1"}, Glyph: "1-9", Kind: ShowTab, Tab: TabProjects, NeedsWorkspace: true, Presents: []string{"projects"}, Footer: "views", Tail: true},
	{ID: "view.run", Group: GroupViews, Title: "Run", Keys: []string{"2"}, Kind: ShowTab, Tab: TabRun, NeedsWorkspace: true},
	{ID: "view.doctor", Group: GroupViews, Title: "Doctor", Keys: []string{"3"}, Kind: ShowTab, Tab: TabDoctor, NeedsWorkspace: true},
	{ID: "view.release", Group: GroupViews, Title: "Release", Hint: "what the next release would do", Keys: []string{"4"}, Kind: ShowTab, Tab: TabRelease, NeedsWorkspace: true, Presents: []string{"release"}},
	{ID: "view.plugins", Group: GroupViews, Title: "Plugins", Hint: "built-in and installed plugins, their roles and what uses them", Keys: []string{"5"}, Kind: ShowTab, Tab: TabPlugins, NeedsWorkspace: true, Presents: []string{"plugin", "plugin enable", "plugin disable"}},
	{ID: "view.deps", Group: GroupViews, Title: "Dependencies", Hint: "newer releases and the one-version policy", Keys: []string{"6"}, Kind: ShowTab, Tab: TabDeps, NeedsWorkspace: true, Presents: []string{"up"}},
	{ID: "view.recipes", Group: GroupViews, Title: "Recipes", Hint: "what a recipe puts in each role, and new projects from its choices", Keys: []string{"7"}, Kind: ShowTab, Tab: TabRecipes, NeedsWorkspace: true, Presents: []string{"recipe"}},
	{ID: "view.settings", Group: GroupViews, Title: "Settings", Hint: "smith.toml, checked against its schema", Keys: []string{"8"}, Kind: ShowTab, Tab: TabSettings, NeedsWorkspace: true, Presents: []string{"config", "config set", "config unset"}},
	{ID: "view.guide", Group: GroupViews, Title: "Guided actions", Hint: "the step-by-step menu", Keys: []string{"9"}, Kind: ShowTab, Tab: TabGuide},
	{ID: "view.copy", Group: GroupViews, Title: "Copy the focused item", Hint: "remedy, command, project path or package to the clipboard", Keys: []string{"y"}, Kind: CopyFocused, Footer: "copy", ShowIn: vCopy},
	{ID: "view.mouse", Group: GroupViews, Title: "Mouse on or off", Hint: "off lets the terminal select text; Shift+drag selects with it on", Keys: []string{"m"}, Kind: ToggleMouse, Footer: "mouse", Tail: true},

	// Help
	{ID: "file.palette", Group: GroupHelp, Title: "Command palette…", Hint: "find and run any smith command", Keys: []string{":", "ctrl+p"}, Kind: Palette, WhileEditing: true, Footer: "palette", Tail: true},
	{ID: "file.find", Group: GroupHelp, Title: "Find a command…", Hint: "the same palette, opened with the filter key", Keys: []string{"/"}, Kind: Palette},
	{ID: "help.keys", Group: GroupHelp, Title: "Keyboard shortcuts", Hint: "every shortcut", Keys: []string{"?", "f1"}, Kind: Help, WhileEditing: true, Footer: "help", Tail: true},
	{ID: "help.about", Group: GroupHelp, Title: "About smith", Kind: About},
	{ID: "file.quit", Group: GroupHelp, Title: "Quit", Hint: "leave smith (not while a run is going: Ctrl+C stops it first)", Keys: []string{"q"}, Kind: Quit, Footer: "quit", Tail: true},
}

// All returns every command.
func All() []Command { return append([]Command(nil), all...) }

// init gives each command the shortcut its keys spell, so the label shown
// beside a control is always the key that triggers it.
func init() {
	for i := range all {
		if all[i].Shortcut == "" {
			all[i].Shortcut = keysLabel(all[i].Keys)
		}
	}
}
