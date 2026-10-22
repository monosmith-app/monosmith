package actions

// Reasons an action is left out of the interactive shell's menus. Each
// action's own reason names what is missing; these are the recurring ones.
const (
	needsForm    = "no menu entry or guided form yet: it takes arguments the shell cannot ask for (#302)"
	ownTerminal  = "takes over the terminal, so the shell hands it back to one"
	referenceOut = "reference output for tools and scripts; the command palette runs it"
)

// catalog is every action. Cobra's own help, completion and version are not
// here (see builtins); hidden commands are internal and are not actions.
var catalog = []Action{
	// Create
	{Command: []string{"new"}, Safety: Changes},
	{Command: []string{"add"}, Safety: Changes},
	{Command: []string{"remove"}, Safety: Changes, Asks: true},
	{Command: []string{"adopt"}, Safety: Reads, Variants: []Variant{{When: []string{"--tags", "--init"}, Safety: Changes}}},
	{Command: []string{"convert"}, Safety: Changes},
	{Command: []string{"upgrade"}, Safety: Changes},
	{Command: []string{"config"}, Safety: Reads},
	{Command: []string{"config", "set"}, Safety: Changes},
	{Command: []string{"config", "unset"}, Safety: Changes},

	// Work
	{Command: []string{"run"}, Safety: Reads},
	{Command: []string{"build"}, Safety: Reads},
	{Command: []string{"test"}, Safety: Reads},
	{Command: []string{"lint"}, Safety: Reads},
	{Command: []string{"typecheck"}, Safety: Reads},
	{Command: []string{"dev"}, Safety: NeedsTerminal, NotInTUI: ownTerminal},
	{Command: []string{"install"}, Safety: Changes},
	{Command: []string{"uninstall"}, Safety: Changes, NotInTUI: needsForm},
	{Command: []string{"sync"}, Safety: Changes, Variants: []Variant{{When: []string{"--check"}, Safety: Reads}}, NotInTUI: needsForm},
	{Command: []string{"deps"}, Safety: Changes, NotInTUI: needsForm},
	{Command: []string{"up"}, Safety: NeedsTerminal, Variants: []Variant{
		{When: []string{"--yes"}, Safety: Changes},
		{When: []string{"--check", "--json"}, Safety: Reads},
	}},

	// Ship
	{Command: []string{"release"}, Safety: Changes, Asks: true},
	{Command: []string{"ci"}, Safety: Changes},
	{Command: []string{"pipeline"}, Safety: Changes, NotInTUI: needsForm},
	{Command: []string{"docker"}, Safety: Changes, NotInTUI: needsForm},

	// Inspect
	{Command: []string{"doctor"}, Safety: Reads, Asks: true, Variants: []Variant{{When: []string{"--fix"}, Safety: Changes}}},
	{Command: []string{"projects"}, Safety: Reads},
	{Command: []string{"affected"}, Safety: Reads},
	{Command: []string{"graph"}, Safety: Reads, Variants: []Variant{{When: []string{"tui", "--format=tui"}, Safety: NeedsTerminal}}},
	{Command: []string{"info"}, Safety: Reads},
	{Command: []string{"explain"}, Safety: Reads, NotInTUI: needsForm},
	{Command: []string{"why-miss"}, Safety: Reads, NotInTUI: needsForm},
	{Command: []string{"slices"}, Safety: Reads, NotInTUI: needsForm},
	{Command: []string{"kinds"}, Safety: Reads, NotInTUI: referenceOut},
	{Command: []string{"stacks"}, Safety: Reads, NotInTUI: referenceOut},
	{Command: []string{"commands"}, Safety: Reads, NotInTUI: referenceOut},
	{Command: []string{"schema"}, Safety: Reads, NotInTUI: referenceOut},
	{Command: []string{"plugin"}, Safety: Reads},
	{Command: []string{"plugin", "set"}, Safety: Changes},
	{Command: []string{"plugin", "enable"}, Safety: Changes},
	{Command: []string{"plugin", "disable"}, Safety: Changes},
	{Command: []string{"recipe"}, Safety: Reads},
	{Command: []string{"recipe", "new"}, Safety: NeedsTerminal, Variants: []Variant{{When: []string{"--from-answers"}, Safety: Changes}}, NotInTUI: "its wizard asks on the terminal, and the Recipes view shows what a recipe decides"},
	{Command: []string{"mcp"}, Safety: NeedsTerminal, NotInTUI: ownTerminal},
	{Command: []string{"serve"}, Safety: NeedsTerminal, NotInTUI: ownTerminal},
}
