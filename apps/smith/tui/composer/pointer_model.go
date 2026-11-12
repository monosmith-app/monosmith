package composer

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// browseListTop is the first body line of the list while browsing: the search
// box, the summary and a blank line come first.
const browseListTop = 3

// Click handles a left press at the screen cell (x, y). A press on a row of the
// list moves the cursor there and ticks or unticks it, as Enter does. A press
// on an answer of the open question picks it, as Enter does. A press anywhere
// else changes nothing.
//
// Parameters: x and y are screen cells.
// Returns: the screen after the press and what its host must do.
func (m Model) Click(x, y int) (Model, Outcome) {
	switch m.stage {
	case stBrowse:
		if i := m.rowAt(x, y); i >= 0 {
			m.cursor = i
			return m.focusList().enter(), Outcome{}
		}
	case stQuestion:
		if i := m.answerAt(x, y); i >= 0 {
			m.qcursor = i
			return m.questionKey("enter")
		}
	}
	return m, Outcome{}
}

// Hover notes the row under the pointer at (x, y), so it is drawn in the row
// hover look; a cell on no row clears it.
//
// Parameters: x and y are screen cells.
// Returns: the screen with the hover set.
func (m Model) Hover(x, y int) Model {
	m.hover = -1
	switch m.stage {
	case stBrowse:
		m.hover = m.rowAt(x, y)
	case stQuestion:
		m.hover = m.answerAt(x, y)
	}
	return m
}

// bodyLine is the line of the framed body at screen row y, or -1 outside the
// frame.
func (m Model) bodyLine(x, y int) int {
	if !m.rect.Contains(x, y) || y <= m.rect.Y || y >= m.rect.Y+m.rect.H-1 {
		return -1
	}
	return y - m.rect.Y - 1
}

// rowAt is the index of the list row drawn at (x, y) while browsing, or -1. It
// counts the lines as browseBody draws them, through the same window.
func (m Model) rowAt(x, y int) int {
	line := m.bodyLine(x, y) - browseListTop
	if line < 0 || m.rect.W < 20 || m.rect.H < 8 {
		return -1
	}
	th := theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII})
	inner, height := m.rect.W-4, m.rect.H-2
	room := max(height-browseListTop-len(m.statusLines(th, inner))-len(m.keyLines(th, inner)), 3)
	probe := m
	lines, items := probe.listLines(th, inner)
	shown, top := len(lines), 0
	if len(lines) > room {
		shown = room - 1 // the line under the window says what is hidden
		// A copy of the scroll: finding a row must not move the window.
		sc := pane.Scroll{}
		if m.scroll != nil {
			sc = *m.scroll
		}
		top = sc.Show(probe.cursorLine, shown, len(lines))
	}
	if line >= shown || top+line >= len(items) {
		return -1
	}
	return items[top+line]
}

// answerAt is the index of the answer of the open question drawn at (x, y), or
// -1. It counts the lines as questionBody draws them.
func (m Model) answerAt(x, y int) int {
	if m.chk.Question == nil {
		return -1
	}
	th := theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII})
	body := m.questionBody(th, m.rect.W-4)
	first := len(body) - 2 - len(m.chk.Question.Options) // the keys line and a blank close the body
	i := m.bodyLine(x, y) - first
	if m.bodyLine(x, y) < 0 || i < 0 || i >= len(m.chk.Question.Options) {
		return -1
	}
	return i
}
