package projectlist

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

// VersionText is the version of a project as the list and the detail show it:
// "current → next" while a release is pending, else the current one alone. When
// the arrow form is wider than w cells it shortens to "→ next", which still says
// that a release is pending, and when even that does not fit it is cut.
//
// Parameters: current is the version now ("" when none is known); next is the
// version the pending release makes ("" when none is pending); w is the room in cells.
// Returns: the text, at most w cells wide.
func VersionText(current, next string, w int) string {
	if next == "" || next == current {
		return ansi.Truncate(current, w, "…")
	}
	if current == "" {
		return ansi.Truncate("→ "+next, w, "…")
	}
	if full := current + " → " + next; pane.Width(full) <= w {
		return full
	}
	return ansi.Truncate("→ "+next, w, "…")
}
