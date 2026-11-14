package statusbar

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Render draws the bar in w cells. The right side (the version) is
// dropped before the left is cut short.
func Render(th theme.Theme, in Info, w int) string {
	left, right := leftSide(th, in), rightSide(th, in)
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

// leftSide is the left of the bar: the command or the hint, its state and result.
func leftSide(th theme.Theme, in Info) string {
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
		middle = th.Muted(th.Icons.Cancelled + words.Say("statusbar.cancelled"))
	}
	if in.Result != "" {
		if middle != "" {
			middle += " "
		}
		middle += in.Result
	}
	if middle != "" {
		left += "   " + middle
	}
	if in.MouseOff {
		left += "   " + th.Warning(words.Say("statusbar.mouseoff"))
	}
	return left
}

// rightSide is the right of the bar: the red "update available" with a dot
// when a newer smith is known, then the version.
func rightSide(th theme.Theme, in Info) string {
	var right string
	if in.Update != "" {
		right = th.Failure(words.Say("statusbar.update")) + " " + th.Muted(th.Icons.Dot) + " "
	}
	if in.Version != "" {
		right += th.Muted(words.Say("statusbar.version", "version", in.Version))
	}
	return right
}

// UpdateSpan is the columns the words "update available" fill in a bar of w
// cells drawn for in, so a click on them can open the update; ok is false when
// the bar shows no update, or has no room for the right side.
//
// Parameters: th is the theme the bar is drawn in; in is what the bar shows; w is its width.
// Returns: first, the column of the first cell; end, one past the last cell.
func UpdateSpan(th theme.Theme, in Info, w int) (first, end int, ok bool) {
	if in.Update == "" {
		return 0, 0, false
	}
	left, right := leftSide(th, in), rightSide(th, in)
	if pane.Width(left)+2+pane.Width(right)+1 > w {
		return 0, 0, false
	}
	first = w - pane.Width(right) - 1
	return first, first + pane.Width(words.Say("statusbar.update")), true
}
