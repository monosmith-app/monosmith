package statusbar

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// Render draws the bar in w cells. The right side (the version) is
// dropped before the left is cut short.
func Render(th theme.Theme, in Info, w int) string {
	left := th.Muted(words.Say("statusbar.ready"))
	switch {
	case in.Hint != "":
		left = " " + th.Warning(th.Icons.Warning) + " " + in.Hint
	case in.Command != "":
		left = " " + th.Faint("$") + " " + in.Command
	}
	middle := ""
	switch in.State {
	case Running:
		middle = th.Info(th.Spin(in.Tick) + words.Say("statusbar.running"))
	case Passed:
		middle = th.Success(th.Icons.Passed + words.Say("statusbar.passed"))
	case Failed:
		middle = th.Failure(th.Icons.Failed + words.Say("statusbar.failed"))
	case Cancelled:
		middle = th.Warning(th.Icons.Skipped + words.Say("statusbar.cancelled"))
	}
	if in.Result != "" {
		if middle != "" {
			middle += " "
		}
		middle += in.Result
	}
	var right string
	if in.Version != "" {
		right += th.Muted(words.Say("statusbar.version", "version", in.Version))
	}
	if middle != "" {
		left += "   " + middle
	}
	if in.MouseOff {
		left += "   " + th.Warning(words.Say("statusbar.mouseoff"))
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
