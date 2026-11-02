package catalog

// Safety says what an action does to the workspace, which decides how an
// interface may run it.
type Safety int

// The safeties.
const (
	// Reads actions only read the workspace or run its tasks: no confirmation.
	Reads Safety = iota
	// Changes actions alter the workspace, install or publish: an interface
	// that cannot ask on a terminal shows the command line and asks first.
	Changes
	// NeedsTerminal actions ask questions or take over the screen, so they
	// can only run in a terminal of their own.
	NeedsTerminal
)

// Variant is a form of an action whose safety differs from the action's own:
// when the command line has any of the words in When, Safety replaces the
// action's.
type Variant struct {
	When   []string
	Safety Safety
}

// Action is one smith command.
type Action struct {
	// Command is the command's path under smith: {"plugin", "set"}.
	Command []string
	Safety  Safety
	// Variants are the forms of it with another safety, the first match wins.
	Variants []Variant
	// Asks means that, when it changes the workspace, it asks for a yes on a
	// terminal, so an interface that has confirmed it itself passes --yes.
	Asks bool
	// Plan is the flag that prints what the action would do and changes
	// nothing (--dry-run, or --check for a report). An interface that confirms
	// the action itself runs it first and shows the text in the dialog, so the
	// person answers yes to a plan and not to a sentence. Empty means the
	// action has no such flag.
	Plan string
	// NotInTUI is why the interactive shell's menus and guided forms leave
	// it out (it stays reachable from the command palette). Empty means a
	// menu entry or a guided form presents it.
	NotInTUI string
	// NotInMCP is why the MCP server has no tool for it. Empty means a tool
	// runs it (the tool's Action names this action).
	NotInMCP string
	// NotInServe is why `smith serve` (the editors' engine) has no method for
	// it. Empty means a method runs it.
	NotInServe string
}

// ID is the action's name: its command path joined by spaces.
func (a Action) ID() string {
	id := ""
	for i, w := range a.Command {
		if i > 0 {
			id += " "
		}
		id += w
	}
	return id
}
