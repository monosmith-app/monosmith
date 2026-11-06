package guide

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// wordsHint is the line a folded glossary leaves.
var wordsHint = words.Say("guide.list_window.words_hint")

const (
	// glossaryListRows is how many lines of the list a glossary must leave. A row of
	// the list wraps to two or three lines, so 16 lines are about 7 rows; when the
	// words would leave fewer, they fold to the one line above.
	glossaryListRows = 16
	// minListRows is how many rows of the list the pane always keeps.
	minListRows = 8
)

// Window draws the screen in w cells and h lines, scrolling the list of
// options when it does not fit (see pane.Window): the question stays at the
// top and the keys at the bottom, the wheel scrolls the list, and the
// selection keeps in view when a key moves it. A glossary folds to one line
// whenever the pane would leave the list fewer than 16 rows (the height of the
// pane decides, not that of the terminal); "?" opens it.
//
// Parameters: th is the theme; w and h are the cells of the pane.
// Returns: the lines, and optionAt that maps each line of an option to the
// option, for a click.
func (m Model) Window(th theme.Theme, w, h int) (lines []string, optionAt map[int]int) {
	fold := false
	for {
		lines, optionAt = m.body(th, w-1, fold)
		first, last, ok := optionSpan(optionAt)
		if !ok || len(lines) <= h {
			return lines, optionAt
		}
		head, rows, foot := lines[:first], lines[first:last+1], lines[last+1:]
		room := h - len(head) - len(foot)
		if room < glossaryListRows && m.HasGlossary() && !m.words && !fold {
			fold = true // the words give way to the list
			continue
		}
		room = max(room, min(minListRows+1, len(rows))) // the rows and the "more" note
		view := m.view
		if view == nil {
			view = &listView{}
		}
		sel := -1
		for line, i := range optionAt {
			if i == m.cursor && (sel < 0 || line-first < sel) {
				sel = line - first
			}
		}
		shown, top := pane.Window(th, rows, w, room, max(sel, 0), &view.scroll)
		view.page = len(shown)
		out := append(append(append([]string{}, head...), shown...), foot...)
		at := map[int]int{}
		count := len(shown)
		if len(rows) > room {
			count-- // the last line is the "more" note
		}
		for i := 0; i < count; i++ {
			if o, ok := optionAt[first+top+i]; ok {
				at[len(head)+i] = o
			}
		}
		return out, at
	}
}

// optionSpan is the first and the last line that an option takes.
func optionSpan(optionAt map[int]int) (first, last int, ok bool) {
	for line := range optionAt {
		if !ok || line < first {
			first = line
		}
		if !ok || line > last {
			last = line
		}
		ok = true
	}
	return first, last, ok
}
