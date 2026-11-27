package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/logsearch"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/logview"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// The search of the log of the Runs view: / types the text, Enter keeps it, n
// and N go from match to match, Esc clears it. The matches are lit in the
// accent and the title of the log counts them.

// logSearchKey types the search of the log: each letter searches at once and
// goes to the first match, Enter keeps the text and gives the keys back to the
// view, Esc clears it.
func (s *Shell) logSearchKey(msg tea.KeyMsg) tea.Cmd {
	text := s.logQuery
	switch msg.Type {
	case tea.KeyEnter:
		s.logTyping = false
		return nil
	case tea.KeyEsc:
		s.logTyping, text = false, ""
	case tea.KeyBackspace:
		if r := []rune(text); len(r) > 0 {
			text = string(r[:len(r)-1])
		}
	case tea.KeyCtrlU:
		text = ""
	case tea.KeyRunes, tea.KeySpace:
		text += string(msg.Runes)
	default:
		return nil
	}
	s.logQuery, s.logAt = text, 0
	s.showMatch()
	return nil
}

// logMatches are the matches of the search in the log shown, at the width of
// its column.
func (s Shell) logMatches() []logsearch.Match {
	if s.logQuery == "" || s.run == nil {
		return nil
	}
	w, h := s.runRect()
	return logsearch.Find(logview.Rows(s.th, s.logLines(), s.runsGeom(w, h).logW-logPad), s.logQuery)
}

// stepMatch goes to the next match (dir 1) or the one before (dir -1), round
// the end of the log.
func (s *Shell) stepMatch(dir int) {
	n := len(s.logMatches())
	if n == 0 {
		s.say(words.Say("runpage.search.none", "text", s.logQuery))
		return
	}
	s.logAt = ((s.logAt+dir)%n + n) % n
	s.showMatch()
}

// showMatch scrolls the log to the match it is on.
func (s *Shell) showMatch() {
	matches := s.logMatches()
	if len(matches) == 0 {
		return
	}
	s.logAt = min(s.logAt, len(matches)-1)
	s.log = s.log.ShowRow(matches[s.logAt].Row, s.logTotal(), s.logHeight())
}

// searchLine is what the title of the log says of the search: the text (with
// the cursor while it is typed) and which match of how many, or that none is
// found; "" with no search.
func (s Shell) searchLine(rows []string) string {
	if s.logQuery == "" && !s.logTyping {
		return ""
	}
	line := s.th.Muted(words.Say("runpage.search.label")) + " " + s.logQuery
	if s.logTyping {
		line += s.th.Accent("▏")
	}
	if s.logQuery == "" {
		return line
	}
	n := len(logsearch.Find(rows, s.logQuery))
	if n == 0 {
		return line + "  " + s.th.Warning(words.Say("runpage.search.nothing"))
	}
	return line + "  " + s.th.Accent(words.Say("runpage.search.count", "at", min(s.logAt, n-1)+1, "total", n))
}

// litRows are the rows of the log with the matches of the search lit.
func (s Shell) litRows(rows []string) []string {
	matches := logsearch.Find(rows, s.logQuery)
	if len(matches) == 0 {
		return rows
	}
	current := matches[min(s.logAt, len(matches)-1)]
	out := make([]string, len(rows))
	copy(out, rows)
	for i := range out {
		if in := logsearch.InRow(matches, i); len(in) > 0 {
			out[i] = logsearch.Highlight(s.th, out[i], in, current)
		}
	}
	return out
}
