package catalog

import "slices"

// builtins are the commands cobra adds itself; they only print.
var builtins = []string{"help", "completion", "version"}

// Classify decides how a smith command line (without the program name) is
// treated. Anything no action claims counts as a change.
func Classify(argv []string) Safety {
	if len(argv) == 0 {
		return NeedsTerminal
	}
	has := func(words ...string) bool {
		return slices.ContainsFunc(words, func(w string) bool { return slices.Contains(argv, w) })
	}
	if slices.Contains(builtins, argv[0]) {
		return Reads
	}
	action, ok := Find(argv)
	if !ok {
		return Changes
	}
	safety := action.Safety
	for _, v := range action.Variants {
		if has(v.When...) {
			safety = v.Safety
			break
		}
	}
	if safety == NeedsTerminal {
		return NeedsTerminal
	}
	if has("--dry-run", "--help", "-h") {
		return Reads
	}
	return safety
}

// PlanFlag is the flag that makes a command line print its plan and change
// nothing, for a command that changes the workspace and has such a flag.
//
// Parameters: argv is the command line without the program name.
// Returns: the flag, and false when the command has none, already carries it,
// or does not change the workspace.
func PlanFlag(argv []string) (string, bool) {
	action, ok := Find(argv)
	if !ok || action.Plan == "" || Classify(argv) != Changes || slices.Contains(argv, action.Plan) {
		return "", false
	}
	return action.Plan, true
}

// WithConfirmation returns the command line that runs after the user has
// confirmed: an action that asks on a terminal it does not have gets --yes.
// The confirmation shows this line, so what runs is what was confirmed.
func WithConfirmation(argv []string) []string {
	out := slices.Clone(argv)
	action, ok := Find(out)
	if ok && action.Asks && Classify(out) == Changes && !slices.Contains(out, "--yes") && !slices.Contains(out, "-y") {
		out = append(out, "--yes")
	}
	return out
}
