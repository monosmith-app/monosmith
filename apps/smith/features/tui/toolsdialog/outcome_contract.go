package toolsdialog

// Outcome is what a key did beyond changing the dialog.
type Outcome struct {
	// Done is true when the dialog should close.
	Done bool
	// Run is the smith command line to run now (without the leading "smith")
	// and stream into the dialog: an install that needs no terminal.
	Run []string
	// Handover is the smith command line to run in the terminal after the shell
	// quits: an install that asks for a password (sudo) needs a terminal of its
	// own. This is the fallback the dialog keeps.
	Handover []string
	// Stop asks the host to stop the run that is streaming.
	Stop bool
}

// Result is how one tool ended.
type Result struct {
	// Name and Title identify the tool.
	Name, Title string
	// Line says in one sentence what happened: installed, skipped, left to the person.
	Line string
	// Failed is true when an install ran and the tool is still missing.
	Failed bool
	// Output is the end of the install's output, kept for a failure.
	Output []string
}
