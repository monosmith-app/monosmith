package fixwizard

// Phase is where the walk is.
type Phase int

// The phases, in the order a walk goes through them.
const (
	// Asking shows the next question of the fix.
	Asking Phase = iota
	// Reviewing shows what the resolved fix will change, to be confirmed.
	Reviewing
	// Applying runs the fix; the shell reports each step.
	Applying
	// Failed shows the step that stopped the fix and what it printed.
	Failed
)

// Outcome is what a key or a click asked of the shell.
type Outcome struct {
	// Done is true once the dialog should close without applying.
	Done bool
	// Apply is true when the shell should start applying Wizard.Resolved().
	Apply bool
	// Stop is true when the user asked to cancel the step that is running.
	Stop bool
	// Decline and Skip answer a review of a Style that offers them: this one
	// is declined, or this one and all that follow.
	Decline, Skip bool
	// Back asks for the stage before the dialog's first screen (Style.Back).
	Back bool
	// Copy is text for the clipboard ("" for none).
	Copy string
}
