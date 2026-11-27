package runpage

import (
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
	"github.com/muesli/termenv"
)

// Bar draws the progress of a run in w cells: the bar of bubbles/progress, then
// "12/18  67%". bubbles/progress gives the cells of the bar without colour; the
// theme colours the filled part (the error colour once a task failed) and the
// empty part, so the plain theme stays plain. The label always carries the
// counts, so colour is never the only signal.
//
// Parameters: th is the theme; done and total are the tasks that ended and all
// of them; failed is true once a task failed; w is the width.
// Returns: the bar with its label, at most w cells.
func Bar(th theme.Theme, done, total int, failed bool, w int) string {
	percent := 0.0
	if total > 0 {
		percent = min(float64(done)/float64(total), 1)
	}
	label := "  " + words.Say("runpage.bar.label", "done", done, "total", total, "percent", int(percent*100+0.5))
	barW := w - pane.Width(label)
	if barW < 4 {
		return pane.Fit(strings.TrimSpace(label), w)
	}
	full, empty := []rune(th.Icons.Bar)[0], []rune(th.Icons.BarEmpty)[0]
	m := progress.New(progress.WithWidth(barW), progress.WithoutPercentage(), progress.WithColorProfile(termenv.Ascii), progress.WithFillCharacters(full, empty))
	n := strings.Count(m.ViewAs(percent), string(full))
	fill := th.Success
	if failed {
		fill = th.Failure
	}
	return fill(strings.Repeat(string(full), n)) + th.Faint(strings.Repeat(string(empty), max(barW-n, 0))) + label
}
