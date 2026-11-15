// Package adoptflow is what the TUI's adopt wizard asks of smith, as data and
// calls: what a repository is and what it uses today (Found), the full
// conversion's plan for a set of choices, the conversion itself, the group
// of removals still to ask about, and the closing install and checks.
//
// It holds no screen and asks no question. The wizard (tui/adoptwizard) decides
// what to show and when; each call here either reads or does one step, so
// the wizard can run it off the UI goroutine and report its progress.
package adoptflow
