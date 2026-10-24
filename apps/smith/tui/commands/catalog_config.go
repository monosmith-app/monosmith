package commands

// Menu names, in the order of the menu bar.
const (
	MenuFile      = "File"
	MenuWorkspace = "Workspace"
	MenuRun       = "Run"
	MenuProjects  = "Projects"
	MenuRelease   = "Release"
	MenuView      = "View"
	MenuHelp      = "Help"
)

// Mnemonics are the letters that open each menu with Alt. Release uses E
// because R is Run's.
var Mnemonics = map[string]rune{
	MenuFile: 'f', MenuWorkspace: 'w', MenuRun: 'r', MenuProjects: 'p', MenuRelease: 'e', MenuView: 'v', MenuHelp: 'h',
}

// MenuOrder is the order of the menus in the bar.
var MenuOrder = []string{MenuFile, MenuWorkspace, MenuRun, MenuProjects, MenuRelease, MenuView, MenuHelp}

// separatorBefore marks where a rule is drawn in a menu: before these IDs.
var separatorBefore = map[string]bool{
	"file.quit": true, "ws.fix": true, "run.build": true, "run.stop": true, "proj.install": true,
	"rel.preview": true, "view.next-pane": true, "help.about": true,
}

// all is the catalog; the order is the order within each menu.
var all = []Command{
	// File
	{ID: "file.new", Menu: MenuFile, Title: "New workspace…", Hint: "smith new: create a monorepo", Keys: []string{"ctrl+n"}, Shortcut: "Ctrl+N", Kind: Guided, Guide: "Create a workspace", OnlyOutside: true},
	{ID: "file.add", Menu: MenuFile, Title: "Add project…", Hint: "smith add: an app, lib or package", Keys: []string{"ctrl+n"}, Shortcut: "Ctrl+N", Kind: Guided, Guide: "Add a project", NeedsWorkspace: true, NeedsIdle: true, OnlyInside: true},
	{ID: "file.adopt", Menu: MenuFile, Title: "Adopt this repository…", Hint: "smith adopt: bring an existing repository under smith", Kind: Guided, Guide: "Adopt this repository", NeedsIdle: true, OnlyOutside: true, NeedsContent: true},
	{ID: "file.convert", Menu: MenuFile, Title: "Convert an Nx or Turborepo workspace…", Hint: "smith convert: turn it into a smith workspace", Kind: Guided, Guide: "Convert an Nx or Turborepo workspace", NeedsIdle: true, OnlyOutside: true, NeedsContent: true},
	{ID: "file.guide", Menu: MenuFile, Title: "Guided actions", Hint: "the step-by-step menu", Keys: []string{"alt+6"}, Shortcut: "Alt+6", Kind: ShowTab, Tab: TabGuide},
	{ID: "file.palette", Menu: MenuFile, Title: "Command palette…", Hint: "find and run any smith command", Keys: []string{"ctrl+p"}, Shortcut: "Ctrl+P", Kind: Palette},
	{ID: "file.quit", Menu: MenuFile, Title: "Quit", Hint: "leave smith", Keys: []string{"ctrl+q"}, Shortcut: "Ctrl+C", Kind: Quit},

	// Workspace
	{ID: "ws.doctor", Menu: MenuWorkspace, Title: "Doctor", Hint: "check the workspace; every problem names its remedy", Keys: []string{"ctrl+d"}, Shortcut: "Ctrl+D", Kind: ReloadDoctor, NeedsWorkspace: true, Presents: []string{"doctor"}},
	{ID: "ws.fix", Menu: MenuWorkspace, Title: "Fix with doctor…", Hint: "smith doctor --fix: apply the safe fixes", Keys: []string{"ctrl+x"}, Shortcut: "Ctrl+X", Kind: Execute, Argv: []string{"doctor", "--fix"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.settings", Menu: MenuWorkspace, Title: "Settings", Hint: "smith.toml, checked against its schema", Keys: []string{"alt+5"}, Shortcut: "Alt+5", Kind: ShowTab, Tab: TabSettings, NeedsWorkspace: true, Presents: []string{"config", "config set", "config unset"}},
	{ID: "ws.ci", Menu: MenuWorkspace, Title: "Regenerate CI pipelines…", Hint: "smith ci generate", Keys: []string{"ctrl+g"}, Shortcut: "Ctrl+G", Kind: Execute, Argv: []string{"ci", "generate"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.info", Menu: MenuWorkspace, Title: "Info", Hint: "smith info: the installed smith and the workspace's settings", Kind: Execute, Argv: []string{"info"}, NeedsIdle: true},

	// Run
	{ID: "run.selection", Menu: MenuRun, Title: "Run selected task", Hint: "the task selected in the navigator", Keys: []string{"ctrl+r", "f5"}, Shortcut: "Ctrl+R", Kind: RunSelection, NeedsWorkspace: true, NeedsIdle: true, NeedsSelection: true, Presents: []string{"run"}},
	{ID: "run.build", Menu: MenuRun, Title: "Build", Hint: "smith build, on the selected project or all", Keys: []string{"ctrl+b"}, Shortcut: "Ctrl+B", Kind: RunTask, Task: "build", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"build"}},
	{ID: "run.test", Menu: MenuRun, Title: "Test", Hint: "smith test, on the selected project or all", Keys: []string{"ctrl+t"}, Shortcut: "Ctrl+T", Kind: RunTask, Task: "test", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"test"}},
	{ID: "run.lint", Menu: MenuRun, Title: "Lint", Hint: "smith lint, on the selected project or all", Keys: []string{"ctrl+l"}, Shortcut: "Ctrl+L", Kind: RunTask, Task: "lint", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"lint"}},
	{ID: "run.typecheck", Menu: MenuRun, Title: "Typecheck", Hint: "smith typecheck, on the selected project or all", Keys: []string{"ctrl+y"}, Shortcut: "Ctrl+Y", Kind: RunTask, Task: "typecheck", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"typecheck"}},
	{ID: "run.affected", Menu: MenuRun, Title: "Lint and test what changed", Hint: "smith run lint test --affected", Keys: []string{"ctrl+a"}, Shortcut: "Ctrl+A", Kind: Execute, Argv: []string{"run", "lint", "test", "--affected"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "run.stop", Menu: MenuRun, Title: "Stop", Hint: "Ctrl+C stops the run; Ctrl+C again quits", Keys: []string{"ctrl+c"}, Shortcut: "Ctrl+C", Kind: Stop, NeedsRunning: true},
	{ID: "run.follow", Menu: MenuRun, Title: "Follow log", Hint: "stay at the end of the log, or pause to read", Keys: []string{"ctrl+f"}, Shortcut: "Ctrl+F", Kind: Follow},

	// Projects
	{ID: "proj.install", Menu: MenuProjects, Title: "Install dependencies…", Hint: "smith install: a fresh clone's npm install, venv, go mod download", Keys: []string{"f4"}, Shortcut: "F4", Kind: Execute, Argv: []string{"install"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.affected", Menu: MenuProjects, Title: "Show what changed", Hint: "smith affected: projects touched since the base branch", Keys: []string{"f7"}, Shortcut: "F7", Kind: Execute, Argv: []string{"affected"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.graph", Menu: MenuProjects, Title: "Project graph", Hint: "smith graph", Kind: Execute, Argv: []string{"graph"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.plugins", Menu: MenuProjects, Title: "Manage plugins…", Hint: "choose which built-in plugins are on", Keys: []string{"f8"}, Shortcut: "F8", Kind: Guided, Guide: "Manage plugins", NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.upgrade", Menu: MenuProjects, Title: "Upgrade workspace…", Hint: "smith upgrade: migrate smith.toml, regenerate what smith owns", Kind: Execute, Argv: []string{"upgrade"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.remove", Menu: MenuProjects, Title: "Remove project…", Hint: "smith remove: its directory, its traces in the workspace, its CI", Kind: Guided, Guide: "Remove a project", NeedsWorkspace: true, NeedsIdle: true},

	// Release
	{ID: "rel.view", Menu: MenuRelease, Title: "Release view", Hint: "what the next release would do", Keys: []string{"alt+4"}, Shortcut: "Alt+4", Kind: ShowTab, Tab: TabRelease, NeedsWorkspace: true, Presents: []string{"release"}},
	{ID: "rel.preview", Menu: MenuRelease, Title: "Preview next release", Hint: "smith release --dry-run: changes nothing", Keys: []string{"f2"}, Shortcut: "F2", Kind: Execute, Argv: []string{"release", "--dry-run"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "rel.release", Menu: MenuRelease, Title: "Release…", Hint: "smith release: version, publish and tag", Keys: []string{"f3"}, Shortcut: "F3", Kind: Execute, Argv: []string{"release"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},

	// View
	{ID: "view.projects", Menu: MenuView, Title: "Projects", Keys: []string{"alt+1"}, Shortcut: "Alt+1", Kind: ShowTab, Tab: TabProjects, NeedsWorkspace: true, Presents: []string{"projects"}},
	{ID: "view.run", Menu: MenuView, Title: "Run", Keys: []string{"alt+2"}, Shortcut: "Alt+2", Kind: ShowTab, Tab: TabRun, NeedsWorkspace: true},
	{ID: "view.doctor", Menu: MenuView, Title: "Doctor", Keys: []string{"alt+3"}, Shortcut: "Alt+3", Kind: ShowTab, Tab: TabDoctor, NeedsWorkspace: true},
	{ID: "view.plugins", Menu: MenuView, Title: "Plugins", Hint: "built-in and installed plugins, their roles and what uses them", Keys: []string{"alt+7"}, Shortcut: "Alt+7", Kind: ShowTab, Tab: TabPlugins, NeedsWorkspace: true, Presents: []string{"plugin", "plugin enable", "plugin disable"}},
	{ID: "view.recipes", Menu: MenuView, Title: "Recipes", Hint: "what a recipe puts in each role, and new projects from its choices", Keys: []string{"alt+9"}, Shortcut: "Alt+9", Kind: ShowTab, Tab: TabRecipes, NeedsWorkspace: true, Presents: []string{"recipe"}},
	{ID: "view.deps", Menu: MenuView, Title: "Dependencies", Hint: "newer releases and the one-version policy", Keys: []string{"alt+8"}, Shortcut: "Alt+8", Kind: ShowTab, Tab: TabDeps, NeedsWorkspace: true, Presents: []string{"up"}},
	{ID: "view.next-pane", Menu: MenuView, Title: "Next pane", Hint: "move the focus: navigator, main, inspector", Keys: []string{"tab", "f6"}, Shortcut: "Tab", Kind: NextPane},
	{ID: "view.prev-pane", Menu: MenuView, Title: "Previous pane", Keys: []string{"shift+tab"}, Shortcut: "Shift+Tab", Kind: PrevPane},
	{ID: "view.context", Menu: MenuView, Title: "Context menu", Hint: "actions for the selection (the right button does the same)", Keys: []string{"ctrl+o"}, Shortcut: "Ctrl+O", Kind: ContextMenu},

	// Help
	{ID: "help.keys", Menu: MenuHelp, Title: "Keyboard shortcuts", Hint: "every shortcut", Keys: []string{"f1", "?"}, Shortcut: "F1", Kind: Help},
	{ID: "help.about", Menu: MenuHelp, Title: "About smith", Kind: About},
}

// All returns every command.
func All() []Command { return append([]Command(nil), all...) }
