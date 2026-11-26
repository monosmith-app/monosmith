package statusbar

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Render draws the bar in w cells: the command and its result at the left,
// the branch, the version and the doctor state at the right. When the row is
// short the counts lose their words first ("✗ 1  ✓ 17"), then the right side
// drops the branch, then the doctor state, then everything, and only then is
// the left cut short.
//
// Parameters: th is the theme; in is what the bar shows; w is its width.
// Returns: the bar, exactly w cells wide, on the bar background.
func Render(th theme.Theme, in Info, w int) string {
	left, right := fit(th, in, w)
	if right != "" {
		gap := w - pane.Width(left) - pane.Width(right) - 1
		return th.Bar(left + strings.Repeat(" ", gap) + right + " ")
	}
	// A hint longer than the bar ends in an ellipsis instead of being cut.
	if pane.Width(left) > w {
		left = ansi.Truncate(left, max(w, 1), "…")
	}
	return th.Bar(pane.Fit(left, w))
}

// fit picks the widest left and right sides that share w cells; right is ""
// when no right side fits.
func fit(th theme.Theme, in Info, w int) (left, right string) {
	fits := func(l, r string) bool { return pane.Width(l)+2+pane.Width(r)+1 <= w }
	full, short := leftSide(th, in, false), leftSide(th, in, true)
	for _, drop := range [][]bool{{false, false}, {true, false}, {true, true}} {
		r := rightSide(th, in, drop[0], drop[1])
		if r == "" {
			break
		}
		for _, l := range []string{full, short} {
			if fits(l, r) {
				return l, r
			}
		}
	}
	if pane.Width(full) <= w {
		return full, ""
	}
	return short, ""
}

// leftSide is the left of the bar: the command or the hint, its state and
// result. short drops the words of the counts.
func leftSide(th theme.Theme, in Info, short bool) string {
	left := th.Muted(words.Say("statusbar.ready"))
	switch {
	case in.Hint != "":
		left = " " + th.Warning(th.Icons.Warning) + " " + in.Hint
	case in.Command != "":
		left = " " + th.Muted("$ "+in.Command)
	}
	middle := ""
	if in.State == Running {
		middle = th.Info(th.Spin(in.Tick) + words.Say("statusbar.running"))
	}
	if counts := countsOf(th, in, short); counts != "" && in.State != Running {
		middle = counts
	} else {
		switch in.State {
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
	}
	if middle != "" {
		left += "  " + middle
	}
	if in.MouseOff {
		left += "   " + th.Warning(words.Say("statusbar.mouseoff"))
	}
	return left
}

// countsOf is the result of a finished run as its counts with their symbols,
// then the time: "✗ 1 failed  ✓ 17 passed  38s", or "✗ 1  ✓ 17  38s" when
// short. It is "" when the run has no counts.
func countsOf(th theme.Theme, in Info, short bool) string {
	if in.Failed == 0 && in.Passed == 0 {
		return ""
	}
	failed := words.Say("statusbar.failed_count", "icon", th.Icons.Failed, "n", in.Failed)
	passed := words.Say("statusbar.passed_count", "icon", th.Icons.Passed, "n", in.Passed)
	if short {
		failed = words.Say("statusbar.count", "icon", th.Icons.Failed, "n", in.Failed)
		passed = words.Say("statusbar.count", "icon", th.Icons.Passed, "n", in.Passed)
	}
	var parts []string
	if in.Failed > 0 {
		parts = append(parts, th.Failure(failed))
	}
	if in.Passed > 0 {
		parts = append(parts, th.Success(passed))
	}
	if in.Took != "" {
		parts = append(parts, th.Muted(in.Took))
	}
	return strings.Join(parts, "  ")
}

// rightSide is the right of the bar: the red "update available" with a dot
// when a newer smith is known, then the branch, the version and the doctor
// state, with a muted dot between them. noBranch and noHealth leave those out.
func rightSide(th theme.Theme, in Info, noBranch, noHealth bool) string {
	var parts []string
	if in.Branch != "" && !noBranch {
		parts = append(parts, th.Muted(in.Branch))
	}
	if in.Version != "" {
		parts = append(parts, th.Muted(in.Version))
	}
	if in.Health != "" && !noHealth {
		parts = append(parts, in.Health)
	}
	right := strings.Join(parts, th.Muted(words.Say("statusbar.joiner")))
	if in.Update != "" {
		right = th.Failure(words.Say("statusbar.update")) + " " + th.Muted(th.Icons.Dot) + " " + right
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
	_, right := fit(th, in, w)
	if right == "" {
		return 0, 0, false
	}
	first = w - pane.Width(right) - 1
	return first, first + pane.Width(words.Say("statusbar.update")), true
}
