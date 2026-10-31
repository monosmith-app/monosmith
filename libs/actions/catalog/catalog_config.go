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
	{Command: []string{"new"}, Safety: Changes, NotInMCP: "creates a workspace, and the server serves one that exists", NotInServe: "creates a workspace, and the engine serves one that exists"},
	{Command: []string{"add"}, Safety: Changes},
	{Command: []string{"remove"}, Safety: Changes, Asks: true},
	{Command: []string{"adopt"}, Safety: Reads, Variants: []Variant{{When: []string{"--tags", "--init"}, Safety: Changes}}, NotInMCP: "no tool yet: bringing a repository under smith is what the MonoSmith agent will need (#407)", NotInServe: noMethodYet},
	{Command: []string{"convert"}, Safety: Changes, NotInMCP: "no tool yet: converting an Nx or Turborepo workspace is what the MonoSmith agent will need (#407)", NotInServe: noMethodYet},
	{Command: []string{"upgrade"}, Safety: Changes},
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
	{Command: []string{"dev"}, Safety: NeedsTerminal, NotInTUI: ownTerminal, NotInMCP: servesItself, NotInServe: servesItself},
	{Command: []string{"kind"}, Safety: Changes, Variants: []Variant{{When: []string{"--dry-run"}, Safety: Reads}}},
	{Command: []string{"eject"}, Safety: Changes, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"tauri"}, Safety: Changes, NotInTUI: groupOnly, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"tauri", "add"}, Safety: Changes, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"tauri", "icon"}, Safety: Changes, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"install"}, Safety: Changes},
	{Command: []string{"uninstall"}, Safety: Changes},
	{Command: []string{"sync"}, Safety: Changes, Variants: []Variant{{When: []string{"--check"}, Safety: Reads}}},
	{Command: []string{"deps"}, Safety: Changes, NotInTUI: groupOnly, NotInMCP: "no tool yet: smith_updates lists the candidates", NotInServe: noMethodYet},
	{Command: []string{"deps", "up"}, Safety: Changes, NotInMCP: "no tool yet: smith_updates lists the candidates", NotInServe: noMethodYet},
	{Command: []string{"up"}, Safety: NeedsTerminal, Variants: []Variant{
		{When: []string{"--yes"}, Safety: Changes},
		{When: []string{"--check", "--json"}, Safety: Reads},
	}},

	// Ship
	{Command: []string{"release"}, Safety: Changes, Asks: true},
	{Command: []string{"ci"}, Safety: Changes, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"pipeline"}, Safety: Changes, NotInServe: noMethodYet},
	{Command: []string{"docker"}, Safety: Changes, NotInTUI: groupOnly, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"docker", "description"}, Safety: Changes, NotInMCP: noToolYet, NotInServe: noMethodYet},

	// Inspect
	{Command: []string{"doctor"}, Safety: Reads, Asks: true, Variants: []Variant{{When: []string{"--fix"}, Safety: Changes}}},
	{Command: []string{"projects"}, Safety: Reads},
	{Command: []string{"affected"}, Safety: Reads},
	{Command: []string{"graph"}, Safety: Reads, Variants: []Variant{{When: []string{"tui", "--format=tui"}, Safety: NeedsTerminal}}},
	{Command: []string{"info"}, Safety: Reads},
	{Command: []string{"explain"}, Safety: Reads},
	{Command: []string{"why-miss"}, Safety: Reads, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"cache"}, Safety: Reads, NotInTUI: groupOnly, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"cache", "status"}, Safety: Reads, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"cache", "prune"}, Safety: Changes, Variants: []Variant{{When: []string{"--dry-run"}, Safety: Reads}}, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"cache", "clear"}, Safety: Changes, Asks: true, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"cache", "verify"}, Safety: Reads, Variants: []Variant{{When: []string{"--fix"}, Safety: Changes}}, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"boundaries"}, Safety: Reads, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"slices"}, Safety: Reads, NotInTUI: groupOnly, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"slices", "check"}, Safety: Reads, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"kinds"}, Safety: Reads, NotInTUI: referenceOut},
	{Command: []string{"stacks"}, Safety: Reads, NotInTUI: referenceOut, NotInServe: noMethodYet},
	{Command: []string{"commands"}, Safety: Reads, NotInTUI: referenceOut, NotInServe: noMethodYet},
	{Command: []string{"schema"}, Safety: Reads, NotInTUI: referenceOut, NotInMCP: noToolYet, NotInServe: noMethodYet},
	{Command: []string{"plugin"}, Safety: Reads},
	{Command: []string{"plugin", "install"}, Safety: Changes, NotInMCP: settingsEdit, NotInServe: noMethodYet},
	{Command: []string{"plugin", "set"}, Safety: Changes, NotInMCP: settingsEdit, NotInServe: noMethodYet},
	{Command: []string{"plugin", "enable"}, Safety: Changes, NotInMCP: settingsEdit, NotInServe: noMethodYet},
	{Command: []string{"plugin", "disable"}, Safety: Changes, NotInMCP: settingsEdit, NotInServe: noMethodYet},
	{Command: []string{"recipe"}, Safety: Reads},
	{Command: []string{"recipe", "new"}, Safety: NeedsTerminal, Variants: []Variant{{When: []string{"--from-answers"}, Safety: Changes}}, NotInTUI: "its wizard asks on the terminal, so the shell hands it one (the palette lists it)", NotInServe: noMethodYet},
	{Command: []string{"mcp"}, Safety: NeedsTerminal, NotInTUI: ownTerminal, NotInMCP: servesItself, NotInServe: servesItself},
	{Command: []string{"mcp", "install"}, Safety: Changes, Asks: true, NotInMCP: agentInstall, NotInServe: noMethodYet},
	{Command: []string{"agent"}, Safety: NeedsTerminal, NotInTUI: ownTerminal, NotInMCP: "starts Claude Code on the terminal", NotInServe: "starts Claude Code on the terminal"},
	{Command: []string{"agent", "install"}, Safety: Changes, Asks: true, NotInMCP: agentInstall, NotInServe: noMethodYet},
	{Command: []string{"serve"}, Safety: NeedsTerminal, NotInTUI: ownTerminal, NotInMCP: servesItself, NotInServe: servesItself},
}
