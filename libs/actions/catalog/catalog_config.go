package catalog

// Reasons an action is left out of the interactive shell's menus. Each
// action's own reason names what is missing; these are the recurring ones.
const (
	noSubforms   = "its subcommands are not described to the shell, so no form can ask for them yet (#302)"
	ownTerminal  = "takes over the terminal, so the shell hands it back to one"
	referenceOut = "reference output for tools and scripts; the command palette runs it"
	noToolYet    = "no tool yet"
	settingsEdit = "an assistant is not handed a settings edit: an editor that shows the change and asks the person calls it through smith serve"
	byTaskName   = "smith_run runs any task by name"
	servesItself = "it serves a protocol on the terminal itself"
)

// catalog is every action. Cobra's own help, completion and version are not
// here (see builtins); hidden commands are internal and are not actions.
var catalog = []Action{
	// Create
	{Command: []string{"new"}, Safety: Changes, NotInMCP: "creates a workspace, and the server serves one that exists"},
	{Command: []string{"add"}, Safety: Changes},
	{Command: []string{"remove"}, Safety: Changes, Asks: true},
	{Command: []string{"adopt"}, Safety: Reads, Variants: []Variant{{When: []string{"--tags", "--init"}, Safety: Changes}}, NotInMCP: "no tool yet: bringing a repository under smith is what the MonoSmith agent will need (#407)"},
	{Command: []string{"convert"}, Safety: Changes, NotInMCP: "no tool yet: converting an Nx or Turborepo workspace is what the MonoSmith agent will need (#407)"},
	{Command: []string{"upgrade"}, Safety: Changes},
	{Command: []string{"config"}, Safety: Reads, NotInMCP: "smith serve answers settings in process (config.list, config.get); no MCP tool"},
	{Command: []string{"config", "set"}, Safety: Changes, NotInMCP: settingsEdit},
	{Command: []string{"config", "unset"}, Safety: Changes, NotInMCP: settingsEdit},

	// Work
	{Command: []string{"run"}, Safety: Reads},
	{Command: []string{"build"}, Safety: Reads, NotInMCP: byTaskName},
	{Command: []string{"test"}, Safety: Reads, NotInMCP: byTaskName},
	{Command: []string{"lint"}, Safety: Reads, NotInMCP: byTaskName},
	{Command: []string{"typecheck"}, Safety: Reads, NotInMCP: byTaskName},
	{Command: []string{"dev"}, Safety: NeedsTerminal, NotInTUI: ownTerminal, NotInMCP: servesItself},
	{Command: []string{"install"}, Safety: Changes},
	{Command: []string{"uninstall"}, Safety: Changes},
	{Command: []string{"sync"}, Safety: Changes, Variants: []Variant{{When: []string{"--check"}, Safety: Reads}}},
	{Command: []string{"deps"}, Safety: Changes, NotInTUI: noSubforms, NotInMCP: "no tool yet: smith_updates lists the candidates"},
	{Command: []string{"up"}, Safety: NeedsTerminal, Variants: []Variant{
		{When: []string{"--yes"}, Safety: Changes},
		{When: []string{"--check", "--json"}, Safety: Reads},
	}},

	// Ship
	{Command: []string{"release"}, Safety: Changes, Asks: true},
	{Command: []string{"ci"}, Safety: Changes, NotInMCP: noToolYet},
	{Command: []string{"pipeline"}, Safety: Changes},
	{Command: []string{"docker"}, Safety: Changes, NotInTUI: noSubforms, NotInMCP: noToolYet},

	// Inspect
	{Command: []string{"doctor"}, Safety: Reads, Asks: true, Variants: []Variant{{When: []string{"--fix"}, Safety: Changes}}},
	{Command: []string{"projects"}, Safety: Reads},
	{Command: []string{"affected"}, Safety: Reads},
	{Command: []string{"graph"}, Safety: Reads, Variants: []Variant{{When: []string{"tui", "--format=tui"}, Safety: NeedsTerminal}}},
	{Command: []string{"info"}, Safety: Reads},
	{Command: []string{"explain"}, Safety: Reads},
	{Command: []string{"why-miss"}, Safety: Reads, NotInMCP: noToolYet},
	{Command: []string{"slices"}, Safety: Reads, NotInTUI: noSubforms, NotInMCP: noToolYet},
	{Command: []string{"kinds"}, Safety: Reads, NotInTUI: referenceOut},
	{Command: []string{"stacks"}, Safety: Reads, NotInTUI: referenceOut},
	{Command: []string{"commands"}, Safety: Reads, NotInTUI: referenceOut},
	{Command: []string{"schema"}, Safety: Reads, NotInTUI: referenceOut, NotInMCP: noToolYet},
	{Command: []string{"plugin"}, Safety: Reads},
	{Command: []string{"plugin", "set"}, Safety: Changes, NotInMCP: settingsEdit},
	{Command: []string{"plugin", "enable"}, Safety: Changes, NotInMCP: settingsEdit},
	{Command: []string{"plugin", "disable"}, Safety: Changes, NotInMCP: settingsEdit},
	{Command: []string{"recipe"}, Safety: Reads},
	{Command: []string{"recipe", "new"}, Safety: NeedsTerminal, Variants: []Variant{{When: []string{"--from-answers"}, Safety: Changes}}, NotInTUI: "its wizard asks on the terminal, and the Recipes view shows what a recipe decides"},
	{Command: []string{"mcp"}, Safety: NeedsTerminal, NotInTUI: ownTerminal, NotInMCP: servesItself},
	{Command: []string{"serve"}, Safety: NeedsTerminal, NotInTUI: ownTerminal, NotInMCP: servesItself},
}
