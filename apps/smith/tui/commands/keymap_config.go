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
	vHome     = []string{TabHome}
	vProjects = []string{TabProjects}
	vRun      = []string{TabRun}
	vDoctor   = []string{TabDoctor}
	vRelease  = []string{TabRelease}
	vDeps     = []string{TabDeps}
	vGuide    = []string{TabGuide}
	vSettings = []string{TabSettings}
	// vReloadable are the views that read something and can read it again.
	vReloadable = []string{TabHome, TabDoctor, TabRelease, TabDeps, TabSettings}
	// vWorkspace are the views of a workspace where the run keys are shown.
	vWorkspace = []string{TabHome, TabProjects, TabRun, TabDoctor, TabDeps, TabRelease}
	vCopy      = []string{TabProjects, TabRun, TabDoctor, TabDeps}
	// vSplit are the views whose selection has a detail pane beside the list.
	vSplit = []string{TabProjects, TabRun, TabDoctor}
	// vZoom are the views whose detail can fill the pane.
	vZoom = []string{TabProjects, TabDoctor}
)

// all is the keymap; the order is the order within each group and in the footer.
var all = []Command{
	// The keys of each view's own pane come first: they win over a global key
	// that is the same letter (Deps marks all with a, which adds a project elsewhere).
	{ID: "home.next", Group: GroupRun, Title: "Do the suggested next action", Keys: []string{"enter"}, Kind: PaneKey, Views: vHome, Footer: "next", Glyph: "Enter"},

	{ID: "projects.move", Group: GroupRun, Title: "Move in the list", Keys: []string{"down", "up", "j", "k"}, Shortcut: "Up Down / j k", Kind: PaneKey, Views: vProjects},
	{ID: "projects.open", Group: GroupRun, Title: "Open the detail of the project (narrow terminals)", Keys: []string{"enter"}, Kind: PaneKey, Views: vProjects, NeedsNarrowList: true, Footer: "detail", Glyph: "Enter"},
	{ID: "projects.install-package", Group: GroupProjects, Title: "Install a package…", Hint: "smith install -w <project> <package>: asks which project and package, then adds it with the project's own toolchain", Keys: []string{"+"}, Kind: OpenPackagePrompt, Action: "install", NeedsWorkspace: true, NeedsIdle: true},
	{ID: "projects.remove-package", Group: GroupProjects, Title: "Remove a package…", Hint: "smith uninstall -w <project> <package>: asks which project and package", Keys: []string{"-"}, Kind: OpenPackagePrompt, Action: "uninstall", NeedsWorkspace: true, NeedsIdle: true},
	{ID: "projects.back", Group: GroupRun, Title: "Back to the list", Keys: []string{"esc"}, Kind: PaneKey, Views: vProjects, NeedsDetailOpen: true, Footer: "back"},

	{ID: "run.again", Group: GroupRun, Title: "Run the last command again", Hint: "the last command line, as it was", Keys: []string{"r"}, Kind: ViewAction, Action: "rerun", Views: vRun, NeedsWorkspace: true, NeedsIdle: true, Footer: "again"},
	{ID: "run.again-failed", Group: GroupRun, Title: "Run the failed and skipped tasks again", Hint: "smith run on just the tasks that did not pass", Keys: []string{"R"}, Kind: ViewAction, Action: "rerun-failed", Views: vRun, NeedsWorkspace: true, NeedsIdle: true, Footer: "failed"},
	{ID: "run.copy-log", Group: GroupRun, Title: "Copy the log shown", Hint: "the output of the selected task, or all of it, to the clipboard", Keys: []string{"Y"}, Kind: ViewAction, Action: "copy-log", Views: vRun, Footer: "copy log"},
	{ID: "run.older", Group: GroupRun, Title: "Show the previous run", Hint: "the runs of this session are kept", Keys: []string{","}, Kind: ViewAction, Action: "run-older", Views: vRun, NeedsIdle: true, Footer: "runs", Glyph: ", ."},
	{ID: "run.newer", Group: GroupRun, Title: "Show the next run", Keys: []string{"."}, Kind: ViewAction, Action: "run-newer", Views: vRun, NeedsIdle: true},
	{ID: "run.follow", Group: GroupRun, Title: "Follow the log, or pause it", Hint: "stay at the end of the log, or pause to read", Keys: []string{"f"}, Kind: Follow, Views: vRun, Footer: "follow"},
	{ID: "run.all-output", Group: GroupRun, Title: "Show all output, then the start screen", Keys: []string{"esc"}, Kind: PaneKey, Views: vRun, Footer: "start screen"},
	{ID: "run.scroll", Group: GroupRun, Title: "Pick a task, scroll its log", Keys: []string{"down", "up", "pgdown", "pgup"}, Shortcut: "Up Down PgUp PgDn", Glyph: "↑↓", Kind: PaneKey, Views: vRun, Footer: "task"},

	{ID: "doctor.open", Group: GroupWorkspace, Title: "Expand the check: its problem, remedy and fix", Keys: []string{"enter"}, Kind: PaneKey, Views: vDoctor, NeedsNarrowList: true, Footer: "expand", Glyph: "Enter"},
	{ID: "doctor.fix", Group: GroupWorkspace, Title: "Fix this check: the wizard asks, shows what changes, then applies", Keys: []string{"f"}, Kind: ViewAction, Action: "doctor-fix", Views: vDoctor, NeedsWorkspace: true, NeedsIdle: true, Footer: "fix"},
	{ID: "ws.fix", Group: GroupWorkspace, Title: "Fix with doctor…", Hint: "the fix wizard: review the fixes doctor can make, then apply them", Keys: []string{"F"}, Kind: OpenFixWizard, Views: vDoctor, NeedsWorkspace: true, NeedsIdle: true, Footer: "fix all"},
	{ID: "doctor.back", Group: GroupWorkspace, Title: "Back to the list", Keys: []string{"esc"}, Kind: PaneKey, Views: vDoctor, NeedsDetailOpen: true, Footer: "back"},
	{ID: "doctor.page-up", Group: GroupWorkspace, Title: "Scroll the detail up", Keys: []string{"pgup"}, Kind: ViewAction, Action: "detail-up", Views: vDoctor},
	{ID: "doctor.page-down", Group: GroupWorkspace, Title: "Scroll the detail down", Keys: []string{"pgdown"}, Kind: ViewAction, Action: "detail-down", Views: vDoctor},
	{ID: "doctor.group", Group: GroupWorkspace, Title: "Group the checks by project, or list them flat", Keys: []string{"g"}, Kind: PaneKey, Views: vDoctor, Footer: "group"},
	{ID: "doctor.passed", Group: GroupWorkspace, Title: "Show or hide the passed checks", Keys: []string{"p"}, Kind: PaneKey, Views: vDoctor, Footer: "passed"},

	{ID: "release.open", Group: GroupRelease, Title: "Show the commits behind the selected bump", Keys: []string{"enter"}, Kind: PaneKey, Views: vRelease, Footer: "commits", Glyph: "Enter"},
	{ID: "rel.preview", Group: GroupRelease, Title: "Preview next release", Hint: "smith release --dry-run: changes nothing", Keys: []string{"p"}, Kind: Execute, Argv: []string{"release", "--dry-run"}, Views: vRelease, NeedsWorkspace: true, NeedsIdle: true, Footer: "preview"},
	{ID: "rel.notes", Group: GroupRelease, Title: "Print the release notes", Hint: "smith release --dry-run --notes", Keys: []string{"n"}, Kind: ViewAction, Action: "release-notes", Views: vRelease, NeedsWorkspace: true, NeedsIdle: true, Footer: "notes"},
	{ID: "rel.release", Group: GroupRelease, Title: "Release…", Hint: "smith release: version, publish and tag", Keys: []string{"x"}, Kind: Execute, Argv: []string{"release"}, Confirm: true, Views: vRelease, NeedsWorkspace: true, NeedsIdle: true, Footer: "release"},

	{ID: "deps.mark", Group: GroupProjects, Title: "Mark or unmark a dependency", Keys: []string{" "}, Shortcut: "Space", Kind: PaneKey, Views: vDeps, Footer: "mark"},
	{ID: "deps.mark-safe", Group: GroupProjects, Title: "Mark the patch and minor ones", Keys: []string{"s"}, Kind: PaneKey, Views: vDeps, Footer: "safe"},
	{ID: "deps.mark-all", Group: GroupProjects, Title: "Mark all, or none", Keys: []string{"a"}, Kind: PaneKey, Views: vDeps, Footer: "all"},
	{ID: "deps.update", Group: GroupProjects, Title: "Update the marked ones", Hint: "smith deps up", Keys: []string{"u", "enter"}, Kind: ViewAction, Action: "deps-up", Views: vDeps, NeedsWorkspace: true, NeedsIdle: true, Footer: "update"},
	{ID: "deps.align", Group: GroupProjects, Title: "Bring drifting versions to one", Hint: "smith deps up --align", Keys: []string{"v"}, Kind: ViewAction, Action: "deps-align", Views: vDeps, NeedsWorkspace: true, NeedsIdle: true, Footer: "align"},
	{ID: "deps.one-version", Group: GroupProjects, Title: "Require one version of each dependency", Hint: "smith config set dependencies.one_version true", Keys: []string{"o"}, Kind: Execute, Argv: []string{"config", "set", "dependencies.one_version", "true"}, Confirm: true, Views: vDeps, NeedsWorkspace: true, NeedsIdle: true, Footer: "one version"},

	{ID: "settings.move", Group: GroupWorkspace, Title: "Move in the list of sections or its rows", Keys: []string{"down", "up", "j", "k"}, Shortcut: "Up Down / j k", Glyph: "↑↓", Kind: PaneKey, Views: vSettings, Footer: "move"},
	{ID: "settings.change", Group: GroupWorkspace, Title: "Change the selected plugin or setting: a switch flips, a list is ticked, text is typed", Keys: []string{"enter", " "}, Kind: PaneKey, Views: vSettings, Footer: "change"},
	{ID: "settings.sections", Group: GroupWorkspace, Title: "Switch between the list of sections and its rows", Keys: []string{"tab", "left", "right"}, Shortcut: "Tab / Left Right", Glyph: "Tab", Kind: PaneKey, Views: vSettings, Footer: "list/sections"},
	{ID: "settings.back", Group: GroupWorkspace, Title: "Back to the list of sections, or cancel the editor", Keys: []string{"esc"}, Kind: PaneKey, Views: vSettings, Footer: "back"},
	{ID: "settings.trust", Group: GroupWorkspace, Title: "Review an external plugin and trust it", Hint: "smith plugin trust: what it is, what it runs, what it sees", Keys: []string{"t"}, Kind: PaneKey, Views: vSettings, Footer: "trust"},
	{ID: "settings.forget", Group: GroupWorkspace, Title: "Forget that an external plugin is trusted", Hint: "smith plugin untrust", Keys: []string{"f"}, Kind: PaneKey, Views: vSettings, Footer: "forget"},
	{ID: "settings.unset", Group: GroupWorkspace, Title: "Reset a setting to its default", Hint: "smith config unset", Keys: []string{"u"}, Kind: PaneKey, Views: vSettings, Footer: "reset"},

	{ID: "guide.choose", Group: GroupStart, Title: "Choose and go on", Keys: []string{"enter"}, Kind: PaneKey, Views: vGuide, Footer: "choose", Glyph: "Enter"},
	{ID: "guide.back", Group: GroupStart, Title: "Back one step", Keys: []string{"esc"}, Kind: PaneKey, Views: vGuide, Footer: "back"},

	{ID: "view.reload", Group: GroupViews, Title: "Read the view again", Keys: []string{"r"}, Kind: ViewAction, Action: "reload", Views: vReloadable, NeedsWorkspace: true, NeedsIdle: true, Footer: "reload"},

	// Moving: the keys every pane understands, handled by the pane itself.
	{ID: "move.arrows", Group: GroupMoving, Title: "Move in a list", Keys: []string{"down", "up", "j", "k"}, Shortcut: "Up Down / j k", Kind: PaneKey},
	{ID: "move.enter", Group: GroupMoving, Title: "Open or run the selection", Keys: []string{"enter"}, Kind: PaneKey},
	{ID: "move.esc", Group: GroupMoving, Title: "Back, close, clear (never quits)", Keys: []string{"esc"}, Kind: PaneKey},

	// Run
	{ID: "run.build", Group: GroupRun, Title: "Build", Hint: "smith run build, on the project selected on Projects, else on every project", Keys: []string{"b"}, Kind: RunTask, Task: "build", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"build"}, Footer: "build", ScopeInFooter: true, ShowIn: vWorkspace},
	{ID: "run.build-all", Group: GroupRun, Title: "Build all", Hint: "smith run build, on every project", Keys: []string{"B"}, Kind: RunTask, Task: "build", All: true, NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"build"}, Footer: "build all", ShowIn: vWorkspace},
	{ID: "run.test", Group: GroupRun, Title: "Test", Hint: "smith run test, on the project selected on Projects, else on every project", Keys: []string{"t"}, Kind: RunTask, Task: "test", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"test"}, Footer: "test", ScopeInFooter: true, ShowIn: vWorkspace},
	{ID: "run.lint", Group: GroupRun, Title: "Lint", Hint: "smith run lint, on the project selected on Projects, else on every project", Keys: []string{"l"}, Kind: RunTask, Task: "lint", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"lint"}, Footer: "lint", ShowIn: vProjects},
	{ID: "run.typecheck", Group: GroupRun, Title: "Typecheck", Hint: "smith run typecheck, on the project selected on Projects, else on every project", Keys: []string{"c"}, Kind: RunTask, Task: "typecheck", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"typecheck"}, Footer: "typecheck", ShowIn: vProjects},
	{ID: "run.affected", Group: GroupRun, Title: "Lint and test what changed", Hint: "smith run lint test --affected", Keys: []string{"w"}, Kind: Execute, Argv: []string{"run", "lint", "test", "--affected"}, NeedsWorkspace: true, NeedsIdle: true, Footer: "changed", ShowIn: []string{TabHome, TabProjects}},
	{ID: "run.test-all", Group: GroupRun, Title: "Test all", Hint: "smith run test, on every project", Keys: []string{"T"}, Kind: RunTask, Task: "test", All: true, NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"test"}},
	{ID: "run.lint-all", Group: GroupRun, Title: "Lint all", Hint: "smith run lint, on every project", Keys: []string{"L"}, Kind: RunTask, Task: "lint", All: true, NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"lint"}},
	{ID: "run.typecheck-all", Group: GroupRun, Title: "Typecheck all", Hint: "smith run typecheck, on every project", Keys: []string{"C"}, Kind: RunTask, Task: "typecheck", All: true, NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"typecheck"}},
	{ID: "run.verify", Group: GroupRun, Title: "Verify all", Hint: "smith run lint typecheck test, on every project: what CI checks before it builds", Keys: []string{"V"}, Kind: Execute, Argv: []string{"run", "lint", "typecheck", "test"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "run.stop", Group: GroupRun, Title: "Stop", Hint: "stops the run; pressing it again quits", Keys: []string{"ctrl+c"}, Kind: Stop, NeedsRunning: true, Footer: "stop", ShowIn: vWorkspace},

	// Workspace
	{ID: "ws.doctor", Group: GroupWorkspace, Title: "Doctor", Hint: "check the workspace; every problem names its remedy", Keys: []string{"d"}, Kind: OpenDoctor, NeedsWorkspace: true, Presents: []string{"doctor"}, Footer: "doctor", ShowIn: []string{TabHome, TabProjects}},
	{ID: "ws.ci", Group: GroupWorkspace, Title: "Regenerate CI pipelines…", Hint: "smith ci generate", Kind: Execute, Argv: []string{"ci", "generate"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.info", Group: GroupWorkspace, Title: "Info", Hint: "smith info: the installed smith and the workspace's settings", Kind: Execute, Argv: []string{"info"}, NeedsIdle: true},
	{ID: "ws.undo", Group: GroupWorkspace, Title: "Undo the last change…", Hint: "smith undo: shows what it would revert, then asks", Kind: OpenUndo, NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"undo"}},
	{ID: "ws.report", Group: GroupWorkspace, Title: "Open the report", Hint: "smith report: one web page about the workspace, opened in your browser", Kind: OpenReport, NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"report"}},
	{ID: "ws.agent-file", Group: GroupWorkspace, Title: "Write the agent file…", Hint: "smith agent file: AGENTS.md, what a coding agent should know about this workspace", Kind: Execute, Argv: []string{"agent", "file"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.derive-check", Group: GroupWorkspace, Title: "Check the derived files", Hint: "smith derive --check: which files follow from the project graph, and which were edited", Kind: Execute, Argv: []string{"derive", "--check"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.derive", Group: GroupWorkspace, Title: "Update the derived files…", Hint: "smith derive: rewrite the files that follow from the project graph", Kind: Execute, Argv: []string{"derive"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.derive-preview", Group: GroupWorkspace, Title: "Preview the derived files", Hint: "smith derive --dry-run: the plan of the update, nothing is written", Kind: Execute, Argv: []string{"derive", "--dry-run"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.why", Group: GroupWorkspace, Title: "Why is this here?…", Hint: "smith why <file or setting>: who owns a file, or why a setting has its value, and how to change it", Kind: Guided, Guide: "Why is this here?", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"why"}},
	{ID: "ws.divergences", Group: GroupWorkspace, Title: "Divergences from the defaults", Hint: "smith divergences: every override, ejected file and drift, with the reason and how to go back", Kind: Execute, Argv: []string{"divergences"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.try", Group: GroupWorkspace, Title: "Try a command in a copy…", Hint: "smith try <command>: runs it in a temporary copy, then doctor, and says what changed. Your files stay as they are", Kind: Guided, Guide: "Try a command in a copy", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"try"}},
	{ID: "ws.bootstrap", Group: GroupWorkspace, Title: "Write the launcher…", Hint: "smith bootstrap: tools/smith/smith, so a laptop and CI run the same smith", Kind: Execute, Argv: []string{"bootstrap"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},

	{ID: "cli.guide", Group: GroupHelp, Title: "Commands and docs", Hint: "smith help: every command, with its examples", Kind: Execute, Argv: []string{"help"}, NeedsIdle: true},

	// Start
	{ID: "file.new", Group: GroupStart, Title: "New workspace…", Hint: "smith new: create a monorepo", Keys: []string{"n"}, Kind: Guided, Guide: "Create a workspace", OnlyOutside: true, Footer: "new workspace", ShowIn: vGuide},
	{ID: "file.add", Group: GroupStart, Title: "Add project…", Hint: "smith add: an app, lib or package", Keys: []string{"a"}, Kind: Guided, Guide: "Add a project", NeedsWorkspace: true, NeedsIdle: true, OnlyInside: true, Footer: "add project", ShowIn: vProjects},
	{ID: "file.adopt", Group: GroupStart, Title: "Bring into MonoSmith…", Hint: "bring this folder under smith: what it found, how far to go, every file that changes, then the checks", Keys: []string{"a"}, Kind: OpenAdoptWizard, Views: vGuide, NeedsIdle: true, OnlyOutside: true, NeedsContent: true, Footer: "adopt"},

	// Projects
	{ID: "proj.install", Group: GroupProjects, Title: "Install everything…", Hint: "smith install: installs what every project already declares, as after a fresh clone (to add one package, use Install a package)", Keys: []string{"i"}, Kind: Execute, Argv: []string{"install"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.affected", Group: GroupProjects, Title: "Show what changed", Hint: "smith affected: projects touched since the base branch", Kind: Execute, Argv: []string{"affected"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.graph", Group: GroupProjects, Title: "Project graph", Hint: "smith graph", Kind: Execute, Argv: []string{"graph"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.plugins", Group: GroupProjects, Title: "Manage plugins…", Hint: "choose which built-in plugins are on", Kind: Guided, Guide: "Manage plugins", NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.upgrade", Group: GroupProjects, Title: "Upgrade workspace…", Hint: "smith upgrade: migrate smith.toml, regenerate what smith owns", Kind: Execute, Argv: []string{"upgrade"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.swap", Group: GroupProjects, Title: "Swap a choice…", Hint: "smith swap <project> <axis> <choice>: change a test runner, a linter or a style; the plan comes first", Kind: Guided, Guide: "Swap a choice", NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.remove", Group: GroupProjects, Title: "Remove project…", Hint: "smith remove: its directory, its traces in the workspace, its CI", Kind: Guided, Guide: "Remove a project", NeedsWorkspace: true, NeedsIdle: true},

	// Release

	// Views: the number is the position in the tab row.
	{ID: "view.home", Group: GroupViews, Title: "Home", Hint: "the health of the workspace and what to do next", Keys: []string{"1"}, Glyph: "1-7", Kind: ShowTab, Tab: TabHome, NeedsWorkspace: true, Footer: "views", Tail: true},
	{ID: "view.projects", Group: GroupViews, Title: "Projects", Keys: []string{"2"}, Kind: ShowTab, Tab: TabProjects, NeedsWorkspace: true, Presents: []string{"projects"}},
	{ID: "view.run", Group: GroupViews, Title: "Run", Keys: []string{"3"}, Kind: ShowTab, Tab: TabRun, NeedsWorkspace: true},
	{ID: "view.doctor", Group: GroupViews, Title: "Doctor", Keys: []string{"4"}, Kind: ShowTab, Tab: TabDoctor, NeedsWorkspace: true},
	{ID: "view.deps", Group: GroupViews, Title: "Dependencies", Hint: "newer releases and the one-version policy", Keys: []string{"5"}, Kind: ShowTab, Tab: TabDeps, NeedsWorkspace: true, Presents: []string{"up"}},
	{ID: "view.release", Group: GroupViews, Title: "Release", Hint: "what the next release would do", Keys: []string{"6"}, Kind: ShowTab, Tab: TabRelease, NeedsWorkspace: true, Presents: []string{"release"}},
	// The Settings tab is where plugins, recipes and smith.toml settings are
	// listed and changed (smith plugin, recipe and config in the terminal).
	{ID: "view.settings", Group: GroupViews, Title: "Settings", Hint: "plugins, style and the settings of smith.toml", Keys: []string{"7"}, Kind: ShowTab, Tab: TabSettings, NeedsWorkspace: true, Presents: []string{"plugin", "plugin enable", "plugin disable", "recipe", "config", "config set", "config unset"}},
	{ID: "detail.zoom", Group: GroupViews, Title: "Zoom the detail to the full width, and back", Hint: "the detail alone fills the pane; on a narrow screen it replaces the list", Keys: []string{"z"}, Kind: ViewAction, Action: "detail-zoom", Views: vZoom, Footer: "zoom"},
	{ID: "detail.expand", Group: GroupViews, Title: "Show every line of the problem, or fewer", Keys: []string{"e"}, Kind: ViewAction, Action: "detail-expand", Views: vDoctor},
	{ID: "detail.wider", Group: GroupViews, Title: "Make the detail wider", Hint: "kept until smith closes", Keys: []string{"]"}, Kind: ViewAction, Action: "detail-wider", Views: vSplit},
	{ID: "detail.narrower", Group: GroupViews, Title: "Make the detail narrower", Hint: "kept until smith closes", Keys: []string{"["}, Kind: ViewAction, Action: "detail-narrower", Views: vSplit},
	{ID: "detail.up", Group: GroupViews, Title: "Scroll the detail up", Keys: []string{"ctrl+u"}, Kind: ViewAction, Action: "detail-up", Views: vSplit},
	{ID: "detail.down", Group: GroupViews, Title: "Scroll the detail down", Keys: []string{"ctrl+d"}, Kind: ViewAction, Action: "detail-down", Views: vSplit},
	{ID: "view.copy", Group: GroupViews, Title: "Copy the focused item", Hint: "remedy, command, project path or package to the clipboard", Keys: []string{"y"}, Kind: CopyFocused, Footer: "copy", ShowIn: vCopy},
	{ID: "view.mouse", Group: GroupViews, Title: "Mouse on or off", Hint: "off lets the terminal select text; Shift+drag selects with it on", Keys: []string{"m"}, Kind: ToggleMouse, Footer: "mouse", Tail: true},

	// Help
	{ID: "file.palette", Group: GroupHelp, Title: "Command palette…", Hint: "find and run any smith command", Keys: []string{":", "ctrl+p"}, Kind: Palette, WhileEditing: true, Footer: "palette", Tail: true},
	{ID: "file.find", Group: GroupHelp, Title: "Find a command…", Hint: "the same palette, opened with the filter key", Keys: []string{"/"}, Kind: Palette},
	{ID: "help.keys", Group: GroupHelp, Title: "Keyboard shortcuts", Hint: "every shortcut", Keys: []string{"?", "f1"}, Kind: Help, WhileEditing: true, Footer: "help", Tail: true},
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
