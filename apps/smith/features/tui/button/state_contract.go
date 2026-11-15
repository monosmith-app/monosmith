package button

// State is how a clickable item looks.
type State int

const (
	// Rest is an item nothing is happening to.
	Rest State = iota
	// Hover is an item under the pointer.
	Hover
	// Active is an item being pressed, or the selected tab.
	Active
	// Disabled is an item that does nothing now; it is drawn as plain muted text.
	Disabled
)
