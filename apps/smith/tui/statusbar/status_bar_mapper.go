package statusbar

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Render draws the bar in w cells. The right side (branch, version) is
// dropped before the left is cut short.
func Render(th theme.Theme, in Info, w int) string {
	left := th.Muted(" ready")
	switch {
	case in.Hint != "":
		left = " " + th.Warning(th.Icons.Warning) + " " + in.Hint
	case in.Command != "":
		left = " " + th.Faint("$") + " " + in.Command
	}
	middle := ""
	switch in.State {
	case Running:
		middle = th.Info(th.Spin(in.Tick) + " running")
	case Passed:
		middle = th.Success(th.Icons.Passed + " passed")
	case Failed:
		middle = th.Failure(th.Icons.Failed + " failed")
	case Cancelled:
		middle = th.Warning(th.Icons.Skipped + " cancelled")
	}
	if in.Result != "" {
		if middle != "" {
			middle += " "
		}
		middle += in.Result
	}
	var right string
	if in.Workspace != "" {
		right += th.Accent(in.Workspace) + "  "
	}
	if in.Branch != "" {
		right += th.Muted("⎇ ") + in.Branch + "  "
	}
	if in.Version != "" {
		right += th.Muted("smith " + in.Version)
	}
	if middle != "" {
		left += "   " + middle
	}
	if in.MouseOff {
		left += "   " + th.Warning("mouse off · m turns it on")
	}
	if right != "" && pane.Width(left)+2+pane.Width(right)+1 <= w {
		gap := w - pane.Width(left) - pane.Width(right) - 1
		return th.Bar(left + strings.Repeat(" ", gap) + right + " ")
	}
	// A hint longer than the bar ends in an ellipsis instead of being cut.
	if pane.Width(left) > w {
		left = ansi.Truncate(left, max(w, 1), "…")
	}
	return th.Bar(pane.Fit(left, w))
}
