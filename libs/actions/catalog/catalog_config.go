package catalog

// Reasons an action is left out of the interactive shell's menus. Each
// action's own reason names what is missing; these are the recurring ones.
const (
	groupOnly    = "a group of commands: each subcommand has its own form"
	ownTerminal  = "takes over the terminal, so the shell hands it back to one"
	referenceOut = "reference output for tools and scripts; the command palette runs it"
	noToolYet    = "no tool yet"
	settingsEdit = "an assistant is not handed a settings edit: an editor that shows the change and asks the person calls it through smith serve"
	byTaskName   = "smith_run runs any task by name"
	servesItself = "it serves a protocol on the terminal itself"
	noMethodYet  = "no method yet"
	runByName    = "the run method runs any task by name"
	agentInstall = "installing the agent is the person's step: the assistant that would need it is not running yet"
)

// catalog is every action. Cobra's own help, completion and version are not
// here (see builtins); hidden commands are internal and are not actions.
var catalog = []Action{
	// Create
	{Command: []string{"new"}, Asks: true, Safety: Changes, NotInMCP: "creates a workspace, and the server serves one that exists", NotInServe: "creates a workspace, and the engine serves one that exists"},
	{Command: []string{"add"}, Asks: true, Safety: Changes, Variants: []Variant{{When: []string{"--dry-run"}, Safety: Reads}}},
	{Command: []string{"remove"}, Plan: "--dry-run", Safety: Changes, Asks: true},
	{Command: []string{"adopt"}, Asks: true, Safety: Reads, Variants: []Variant{{When: []string{"--tags", "--init"}, Safety: Changes}}, NotInMCP: "no tool yet: bringing a repository under smith is what the MonoSmith agent will need (#407)", NotInServe: noMethodYet},
	{Command: []string{"convert"}, Plan: "--dry-run", Asks: true, Safety: Changes, NotInMCP: "no tool yet: converting an Nx or Turborepo workspace is what the MonoSmith agent will need (#407)", NotInServe: noMethodYet},
	{Command: []string{"upgrade"}, Plan: "--dry-run", Asks: true, Safety: Changes},
	{Command: []string{"swap"}, Plan: "--dry-run", Asks: true, Safety: Changes, Variants: []Variant{{When: []string{"--dry-run"}, Safety: Reads}}, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"environments"}, Safety: Reads, NotInTUI: groupOnly, NotInMCP: groupOnly, NotInServe: groupOnly},
	{Command: []string{"environments", "list"}, Safety: Reads, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"environments", "add"}, Plan: "--dry-run", Asks: true, Safety: Changes, Variants: []Variant{{When: []string{"--dry-run"}, Safety: Reads}}, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"environments", "remove"}, Plan: "--dry-run", Asks: true, Safety: Changes, Variants: []Variant{{When: []string{"--dry-run"}, Safety: Reads}}, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"bootstrap"}, Plan: "--dry-run", Asks: true, Safety: Changes, NotInMCP: "the person's step: the launcher is what starts smith, so an assistant already has it", NotInServe: noMethodYet},
	{Command: []string{"divergences"}, Safety: Reads, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"copies"}, Safety: Reads, NotInTUI: "a measure for the command line and CI logs", NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"try"}, Safety: Reads, NotInMCP: "an assistant calls the command itself and reads its plan", NotInServe: noMethodYet},
	{Command: []string{"config"}, Safety: Reads, NotInMCP: "smith serve answers settings in process (config.list, config.get); no MCP tool"},
	{Command: []string{"config", "set"}, Safety: Changes, NotInMCP: settingsEdit},
	{Command: []string{"config", "unset"}, Safety: Changes, NotInMCP: settingsEdit},
	{Command: []string{"architecture"}, Safety: Changes, NotInTUI: "the Settings tab has an Architecture section that chooses it", NotInMCP: settingsEdit, NotInServe: "an editor changes smith.toml through config.set; the architecture is a table of it"},

	// Work
	{Command: []string{"run"}, Safety: Reads},
	{Command: []string{"build"}, Safety: Reads, NotInMCP: byTaskName, NotInServe: runByName},
	{Command: []string{"test"}, Safety: Reads, NotInMCP: byTaskName, NotInServe: runByName},
	{Command: []string{"lint"}, Safety: Reads, NotInMCP: byTaskName, NotInServe: runByName},
	{Command: []string{"typecheck"}, Safety: Reads, NotInMCP: byTaskName, NotInServe: runByName},
	{Command: []string{"dev"}, Safety: NeedsTerminal, Terminal: "it runs dev servers until stopped and draws their output on the whole screen", NotInTUI: ownTerminal, NotInMCP: servesItself, NotInServe: servesItself},
	{Command: []string{"kind"}, Plan: "--dry-run", Asks: true, Safety: Changes, Variants: []Variant{{When: []string{"--dry-run"}, Safety: Reads}}},
	{Command: []string{"derive"}, Plan: "--dry-run", Asks: true, Safety: Changes, Variants: []Variant{{When: []string{"--check"}, Safety: Reads}, {When: []string{"--explain"}, Safety: Reads}}, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"undo"}, Asks: true, Safety: Changes, Variants: []Variant{{When: []string{"--dry-run"}, Safety: Reads}}, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"commit"}, Plan: "--dry-run", Asks: true, Safety: Changes, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"eject"}, Plan: "--dry-run", Asks: true, Safety: Changes, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"tauri"}, Safety: Changes, NotInTUI: groupOnly, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"tauri", "add"}, Plan: "--dry-run", Asks: true, Safety: Changes, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"tauri", "icon"}, Plan: "--dry-run", Asks: true, Safety: Changes, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"install"}, Asks: true, Safety: Changes},
	{Command: []string{"uninstall"}, Asks: true, Safety: Changes},
	{Command: []string{"tools"}, Safety: Reads, NotInMCP: noToolYet, NotInServe: "an editor reads the list through smith tools --json"},
	{Command: []string{"tools", "install"}, Safety: NeedsTerminal, Terminal: "a package manager may ask for an administrator password; the shell shows the choices and runs the install in its own dialog, and leaves only for that password", Variants: []Variant{{When: []string{"--dry-run"}, Safety: Reads}}, NotInMCP: "installing on the machine is the person's step: the assistant reads smith tools and says what is missing", NotInServe: "an editor shows the choices from smith tools install --dry-run --json and runs the answer with --manager"},
	{Command: []string{"sync"}, Plan: "--dry-run", Asks: true, Safety: Changes, Variants: []Variant{{When: []string{"--check"}, Safety: Reads}}},
	{Command: []string{"deps"}, Safety: Changes, NotInTUI: groupOnly, NotInMCP: "no tool yet: smith_updates lists the candidates", NotInServe: noMethodYet},
	{Command: []string{"deps", "up"}, Plan: "--check", Asks: true, Safety: Changes, NotInMCP: "no tool yet: smith_updates lists the candidates", NotInServe: noMethodYet},
	{Command: []string{"up"}, Safety: NeedsTerminal, Terminal: "bare `smith up` is an interactive picker; the shell shows the same candidates in its Dependencies tab and never hands this over", Variants: []Variant{
		{When: []string{"--yes"}, Safety: Changes},
		{When: []string{"--check", "--json"}, Safety: Reads},
	}},

	// Ship
	{Command: []string{"release"}, Safety: Changes, Asks: true},
	{Command: []string{"ci", "generate"}, Plan: "--dry-run", Asks: true, Safety: Changes, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"ci"}, Safety: Changes, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"pipeline"}, Plan: "--dry-run", Asks: true, Safety: Changes, NotInServe: noMethodYet},
	{Command: []string{"docker"}, Safety: Changes, NotInTUI: groupOnly, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"docker", "description"}, Safety: Changes, NotInMCP: noToolYet, NotInServe: noMethodYet},

	// Inspect
	{Command: []string{"doctor"}, Safety: Reads, Asks: true, Variants: []Variant{{When: []string{"--fix"}, Safety: Changes}}},
	{Command: []string{"projects"}, Safety: Reads},
	{Command: []string{"affected"}, Safety: Reads},
	{Command: []string{"graph"}, Safety: Reads, Variants: []Variant{{When: []string{"tui", "--format=tui"}, Safety: NeedsTerminal, Terminal: "a full-screen graph browser of its own"}}},
	{Command: []string{"info"}, Safety: Reads},
	{Command: []string{"explain"}, Safety: Reads},
	{Command: []string{"why"}, Safety: Reads, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"why-miss"}, Safety: Reads, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"report"}, Safety: Reads, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"cache"}, Safety: Reads, NotInTUI: groupOnly, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"cache", "status"}, Safety: Reads, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"cache", "prune"}, Plan: "--dry-run", Safety: Changes, Variants: []Variant{{When: []string{"--dry-run"}, Safety: Reads}}, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"cache", "clear"}, Safety: Changes, Asks: true, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"cache", "verify"}, Safety: Reads, Variants: []Variant{{When: []string{"--fix"}, Safety: Changes}}, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"boundaries"}, Safety: Reads, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"slices"}, Safety: Reads, NotInTUI: groupOnly, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"slices", "check"}, Safety: Reads, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"kinds"}, Safety: Reads, NotInTUI: referenceOut},
	{Command: []string{"stacks"}, Safety: Reads, NotInTUI: referenceOut, NotInServe: noMethodYet},
	{Command: []string{"compose"}, Safety: NeedsTerminal, Terminal: "a questionnaire; the shell has its own Compose screen that asks the same questions", Variants: []Variant{{When: []string{"--json"}, Safety: Reads}}, NotInTUI: "the shell has its own Compose a stack screen (palette, or A on the project list), not a guided form", NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"choices"}, Safety: Reads, NotInTUI: referenceOut, NotInServe: noMethodYet},
	{Command: []string{"commands"}, Safety: Reads, NotInTUI: referenceOut, NotInServe: noMethodYet},
	{Command: []string{"schema"}, Safety: Reads, NotInTUI: referenceOut, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"plugin"}, Safety: Reads},
	{Command: []string{"plugin", "draft"}, Safety: Changes, NotInTUI: "for a plugin author, who works at a terminal on a sample project: the command line only", NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"plugin", "install"}, Safety: Changes, NotInMCP: settingsEdit, NotInServe: noMethodYet},
	{Command: []string{"plugin", "set"}, Safety: Changes, NotInMCP: settingsEdit, NotInServe: noMethodYet},
	{Command: []string{"plugin", "enable"}, Safety: Changes, NotInMCP: settingsEdit, NotInServe: noMethodYet},
	{Command: []string{"plugin", "disable"}, Safety: Changes, NotInMCP: settingsEdit, NotInServe: noMethodYet},
	{Command: []string{"recipe"}, Safety: Reads},
	{Command: []string{"recipe", "new"}, Safety: NeedsTerminal, Terminal: "a questionnaire; the shell has a guided form that collects the answers and runs it with --from-answers", Variants: []Variant{{When: []string{"--list-levels"}, Safety: Reads}, {When: []string{"--from-answers", "--fixed", "--name"}, Safety: Changes}}, NotInServe: noMethodYet},
	{Command: []string{"mcp"}, Safety: NeedsTerminal, Terminal: "it serves the MCP protocol on the terminal streams", NotInTUI: ownTerminal, NotInMCP: servesItself, NotInServe: servesItself},
	{Command: []string{"mcp", "install"}, Safety: Changes, Asks: true, NotInMCP: agentInstall, NotInServe: noMethodYet},
	{Command: []string{"agent"}, Safety: NeedsTerminal, Terminal: "it starts Claude Code, an interactive program that owns the terminal", NotInTUI: ownTerminal, NotInMCP: "starts Claude Code on the terminal", NotInServe: "starts Claude Code on the terminal"},
	{Command: []string{"agent", "file"}, Plan: "--dry-run", Asks: true, Safety: Changes, Variants: []Variant{{When: []string{"--check"}, Safety: Reads}}, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"agent", "install"}, Safety: Changes, Asks: true, NotInMCP: agentInstall, NotInServe: noMethodYet},
	{Command: []string{"serve"}, Safety: NeedsTerminal, Terminal: "it serves the editor protocol on the terminal streams", NotInTUI: ownTerminal, NotInMCP: servesItself, NotInServe: servesItself},
	{Command: []string{"lsp"}, Safety: NeedsTerminal, Terminal: "it serves the language protocol on the terminal streams", NotInTUI: ownTerminal, NotInMCP: servesItself, NotInServe: servesItself},
}
