package resultdialog

// Outcome is what a key did beyond changing the dialog. At most one of the
// fields besides Note is set.
type Outcome struct {
	// Done is true when the dialog should close and nothing else happens.
	Done bool
	// Commit is the smith command line (without the leading "smith") that
	// commits the change; the host runs it after its own confirmation.
	Commit []string
	// Install is the smith command line that installs the missing tool.
	Install []string
	// Check is the smith command line that checks the project again.
	Check []string
	// Open is an address to open in the browser and to copy.
	Open string
	// Note is a sentence for the status line, when the choice only closes the dialog.
	Note string
}

// FactKind says how a fact is drawn.
type FactKind int

// The kinds of fact.
const (
	// Good is something that worked.
	Good FactKind = iota
	// Bad is something that could not be done.
	Bad
	// Info is a plain statement.
	Info
)

// Fact is one line of what happened.
type Fact struct {
	Kind FactKind
	Text string
}

// choice is one line of the list: its words and what it does.
type choice struct {
	label, hint string
	out         Outcome
}
