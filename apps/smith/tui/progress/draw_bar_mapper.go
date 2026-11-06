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
