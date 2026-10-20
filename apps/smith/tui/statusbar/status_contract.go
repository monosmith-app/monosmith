package statusbar

// State is what the status bar says about the command.
type State int

// The states.
const (
	Idle State = iota
	Running
	Passed
	Failed
	Cancelled
)

// Info is everything the bar shows.
type Info struct {
	// Command is the exact command line, "smith run build -p kit", or "".
	Command string
	State   State
	// Result is a short summary: "3/8 tasks", "12 passed · 1 failed · 38s".
	Result string
	Branch string
	// Version is smith's version.
	Version string
	// Hint replaces the command while it is set: the reason a menu item is
	// unavailable, a refused key, a toast-sized message.
	Hint string
	// Tick drives the spinner.
	Tick int
}
