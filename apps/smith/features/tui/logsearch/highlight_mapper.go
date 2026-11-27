package logsearch

import (
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
)

// Highlight draws a row with its matches lit: the row as plain text, each
// match in the accent and bold, and the current one also on the selected
// background. A row with no match is given back as it was, with its colours.
//
// Parameters: th is the theme; row is the row as drawn; spans are its matches
// (see InRow); current is the match the person is on (a Match of another row
// lights none).
// Returns: the row.
func Highlight(th theme.Theme, row string, spans []Match, current Match) string {
	if len(spans) == 0 {
		return row
	}
	text := []rune(pane.Strip(row))
	out, at := "", 0
	for _, m := range spans {
		from, to := min(m.From, len(text)), min(m.To, len(text))
		if from < at {
			continue
		}
		out += string(text[at:from])
		lit := th.Accent(th.Bold(string(text[from:to])))
		if m == current {
			lit = th.Selected(lit)
		}
		out += lit
		at = to
	}
	return out + string(text[at:])
}
