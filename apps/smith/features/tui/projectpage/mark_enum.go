package projectpage

import "github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"

// Mark is how a task or a project last ran in this session.
type Mark int

const (
	// NotRun is a task or a project that did not run yet.
	NotRun Mark = iota
	// Passed ran and passed.
	Passed
	// Cached passed from the cache.
	Cached
	// Failed ran and failed.
	Failed
	// Running runs now.
	Running
	// Queued waits for its turn in the run.
	Queued
	// Skipped did not run because a task before it failed.
	Skipped
	// Cancelled was stopped.
	Cancelled
)

// Glyph is the symbol of a mark in its colour: the pass mark in the success
// colour, the failure mark in the error colour, the not-run mark muted, and
// the spinner of tick while it runs. A cached result is a pass.
//
// Parameters: th is the theme; m is the mark; tick is the animation frame.
// Returns: the coloured symbol.
func Glyph(th theme.Theme, m Mark, tick int) string {
	switch m {
	case Passed, Cached:
		return th.Success(th.Icons.Passed)
	case Failed:
		return th.Failure(th.Icons.Failed)
	case Running:
		return th.Info(th.Spin(tick))
	case Queued:
		return th.Muted(th.Icons.Queued)
	case Skipped:
		return th.Warning(th.Icons.Skipped)
	case Cancelled:
		return th.Muted(th.Icons.Cancelled)
	}
	return th.Muted(th.Icons.NotRun)
}
