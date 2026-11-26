package commands

// NeedsOwnTerminal reports whether a command line runs an app's start task:
// the app may draw on the screen and read keys (a TUI, a REPL, a dev server
// with key shortcuts), so it needs a terminal of its own, which the shell's
// run panel cannot give it.
//
// Parameters: argv is the command line without the program name.
// Returns: true for `run start` with any projects and flags.
func NeedsOwnTerminal(argv []string) bool {
	req, ok := ParseRun(argv)
	return ok && argv[0] == "run" && len(req.Tasks) == 1 && req.Tasks[0] == "start"
}
