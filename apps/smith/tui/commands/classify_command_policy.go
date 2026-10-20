package commands

import "slices"

// Class says how the shell treats a smith command line.
type Class int

// The classes.
const (
	// Free commands only read or run tasks: no confirmation.
	Free Class = iota
	// Changes commands alter the workspace, install or publish: the shell
	// shows the command line in a dialog first.
	Changes
	// NeedsTerminal commands ask questions or take over the screen, so they
	// cannot run in a pane; the shell says to run them in a shell.
	NeedsTerminal
)

// readOnly are the top-level commands that only read or run tasks.
var readOnly = []string{
	"run", "build", "test", "lint", "typecheck", "affected", "graph", "projects", "info", "kinds", "stacks",
	"explain", "why-miss", "commands", "schema", "slices", "help", "completion", "version",
}

// Classify decides how a smith command line (without the program name) is treated.
// Anything not known to be harmless counts as a change.
func Classify(argv []string) Class {
	if len(argv) == 0 {
		return NeedsTerminal
	}
	verb := argv[0]
	has := func(flag string) bool { return slices.Contains(argv, flag) }
	switch {
	case verb == "mcp", verb == "up" && !has("--yes"), verb == "dev":
		return NeedsTerminal
	case verb == "graph" && has("--format") && (slices.Contains(argv, "tui") || slices.Contains(argv, "--format=tui")):
		return NeedsTerminal
	case has("--dry-run"), has("--help"), has("-h"):
		return Free
	case verb == "doctor":
		if has("--fix") {
			return Changes
		}
		return Free
	case verb == "plugin":
		if len(argv) > 1 && (argv[1] == "set" || argv[1] == "add" || argv[1] == "remove" || argv[1] == "enable" || argv[1] == "disable") {
			return Changes
		}
		return Free
	case verb == "deps":
		return Changes
	case verb == "recipe" && len(argv) > 1 && argv[1] == "new":
		// The wizard asks its questions on the terminal; with the answers in
		// a file it only writes the recipe.
		if has("--from-answers") {
			return Changes
		}
		return NeedsTerminal
	case verb == "recipe", verb == "config":
		if len(argv) > 1 && (argv[1] == "set" || argv[1] == "unset") {
			return Changes
		}
		return Free
	case slices.Contains(readOnly, verb):
		return Free
	}
	return Changes
}

// WithConfirmation returns the command line that runs after the user has
// confirmed in the dialog: commands that would otherwise ask on a terminal
// they do not have get --yes. The dialog shows this line, so what runs is
// what was confirmed.
func WithConfirmation(argv []string) []string {
	out := slices.Clone(argv)
	asks := (out[0] == "release" || out[0] == "remove" || (out[0] == "doctor" && slices.Contains(out, "--fix"))) &&
		!slices.Contains(out, "--yes") && !slices.Contains(out, "-y") && !slices.Contains(out, "--dry-run")
	if asks {
		out = append(out, "--yes")
	}
	return out
}
