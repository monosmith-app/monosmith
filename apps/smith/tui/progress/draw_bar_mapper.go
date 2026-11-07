package progress

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Draw renders "████░░░░ 3/8 37%" in width cells at most. The filled part is
// the success colour, or the error colour when failed is true; the label
// always carries the counts, so the colour is never the only signal.
func Draw(th theme.Theme, done, total, width int, failed bool) string {
	label := words.Say("progress.bar.label", "done", done, "total", total, "percent", Percent(done, total))
	barWidth := width - len([]rune(label))
	if barWidth < 4 {
		return strings.TrimSpace(label)
	}
	n := Filled(done, total, barWidth)
	fill := strings.Repeat(th.Icons.Bar, n)
	if failed {
		fill = th.Failure(fill)
	} else {
		fill = th.Success(fill)
	}
	return fill + th.Faint(strings.Repeat(th.Icons.BarEmpty, barWidth-n)) + label
}

// DrawCancelled renders the bar of a run the person stopped: "████░░░░ 3/8 done,
// 5 skipped". It never reads 100%, the fill is neutral (a stop is not a
// failure), and the label says what was left undone.
//
// Parameters: th is the theme; done counts the tasks that ended with an
// outcome of their own; skipped counts those that did not run or were
// stopped; total is every task; width is the most cells the bar takes.
// Returns: the bar with its label.
func DrawCancelled(th theme.Theme, done, skipped, total, width int) string {
	label := words.Say("progress.bar.cancelled", "done", done, "total", total, "skipped", skipped)
	barWidth := width - len([]rune(label))
	if barWidth < 4 {
		return strings.TrimSpace(label)
	}
	n := min(Filled(done, total, barWidth), barWidth)
	return th.Muted(strings.Repeat(th.Icons.Bar, n)) + th.Faint(strings.Repeat(th.Icons.BarEmpty, barWidth-n)) + label
}
