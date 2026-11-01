package sources

// Captured is what a smith command printed and whether it succeeded. The
// shell asks for it when it needs the text itself, to show a plan or an error
// in a dialog, instead of showing the command in the Run tab.
type Captured struct {
	// Text is the output of the command, the error included, without colour.
	Text string
	// OK is true when the command ended with code 0.
	OK bool
	// Err is why the command could not start or was stopped; empty Text and a
	// non-nil Err mean smith never ran it.
	Err error
}
