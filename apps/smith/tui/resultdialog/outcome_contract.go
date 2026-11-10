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
	// Stop asks the host to stop the command that runs in the dialog.
	Stop bool
	// Enter is the folder to open the shell in, now that the command made a
	// workspace there.
	Enter string
	// CommitChoice is the answer to "commit each change so it can be undone?":
	// Yes, No or Ask; Open asks the host to show that question.
	CommitChoice CommitChoice
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

// CommitChoice is how the person wants smith to treat undo.
type CommitChoice string

// The answers.
const (
	// CommitYes makes smith commit each change, so it can be undone.
	CommitYes CommitChoice = "yes"
	// CommitNo leaves every change uncommitted.
	CommitNo CommitChoice = "no"
	// CommitAsk shows the commit box on each change.
	CommitAsk CommitChoice = "ask"
	// CommitOpen asks the host to show the question.
	CommitOpen CommitChoice = "open"
)
