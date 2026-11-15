package dialog

// Outcome is how a dialog ended.
type Outcome struct {
	// Done is true once the dialog should close.
	Done bool
	// Confirmed is true when the user chose to go ahead.
	Confirmed bool
}
