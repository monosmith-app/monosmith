package logsearch

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

// Match is one place the text is found: the row, and the runes of the row's
// plain text it covers, From included and To not.
type Match struct {
	Row, From, To int
}

// Find lists every place query is in rows, row by row and left to right,
// ignoring case and the escape codes of the rows. Matches in one row do not
// overlap.
//
// Parameters: rows are the rows of the log as drawn; query is the text.
// Returns: the matches, none for an empty query.
func Find(rows []string, query string) []Match {
	q := []rune(strings.ToLower(query))
	if len(q) == 0 {
		return nil
	}
	var out []Match
	for i, row := range rows {
		text := []rune(strings.ToLower(pane.Strip(row)))
		for at := 0; at+len(q) <= len(text); {
			if string(text[at:at+len(q)]) == string(q) {
				out = append(out, Match{Row: i, From: at, To: at + len(q)})
				at += len(q)
				continue
			}
			at++
		}
	}
	return out
}

// InRow are the matches of one row, in order.
//
// Parameters: matches are the matches of Find; row is the row.
// Returns: the matches on that row.
func InRow(matches []Match, row int) []Match {
	var out []Match
	for _, m := range matches {
		if m.Row == row {
			out = append(out, m)
		}
	}
	return out
}
