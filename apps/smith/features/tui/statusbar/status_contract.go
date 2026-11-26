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
	// The bar shows it when the run has no task counts (Failed and Passed are 0).
	Result string
	// Failed and Passed are the task counts of a finished run, and Took its
	// time ("38s"): the bar shows them as "✗ 1 failed  ✓ 17 passed  38s".
	Failed, Passed int
	Took           string
	// Branch is the git branch of the workspace, or "".
	Branch string
	// Health is the doctor state, already drawn in the theme ("▲2", "✓",
	// "checking…"), or "".
	Health string
	// Version is smith's version.
	Version string
	// Update is the newest smith when a newer one exists and the person did not
	// ask to be left alone about it; the bar then says so, in red, before the version.
	Update string
	// Hint replaces the command while it is set: the reason a key is
	// unavailable, a refused key, a toast-sized message.
	Hint string
	// MouseOff is true while `m` has turned mouse reporting off.
	MouseOff bool
	// Tick drives the spinner.
	Tick int
}
