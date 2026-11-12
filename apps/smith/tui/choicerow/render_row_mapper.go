package choicerow

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Box is the tick box of a row: "[x]" when checked, "[-]" when tri (some of
// what the row stands for is ticked, or the row cannot be ticked), and "[ ]"
// otherwise. tri wins over checked.
func Box(checked, tri bool) string {
	switch {
	case tri:
		return "[-]"
	case checked:
		return "[x]"
	}
	return "[ ]"
}

// Radio is the mark of a choice among several: a dot in the success colour
// when on, blank when off.
func Radio(th theme.Theme, on bool) string {
	if on {
		return "(" + th.Success(th.Icons.Dot) + ")"
	}
	return "( )"
}

// Cursor is the mark in front of a row: the theme's pointer and a space when
// selected, and blank cells of the same display width when not. Every list and
// choice row takes its mark from here, so one theme setting changes them all.
func Cursor(th theme.Theme, selected bool) string {
	mark := th.Icons.Pointer + " "
	if selected {
		return mark
	}
	return strings.Repeat(" ", ansi.StringWidth(mark))
}

// Lines draws a row in w cells: the cursor mark, the lead (a tick box, a radio
// mark, an icon), the label and the muted hint. A row too long for w wraps
// under its label. The selected row has the cursor mark and an accent label; a
// hovered row has the theme's row hover background on every line, the full
// width; a disabled row has a muted label. The cursor mark is blank cells when
// the row is not selected, so a selection never moves the box.
//
// Parameters: th is the theme; r is the row; w is the width, 0 for no limit
// (one line, no background past the text).
// Returns: the lines of the row, at least one.
func Lines(th theme.Theme, r Row, w int) []string {
	label := r.Label
	switch {
	case r.Disabled:
		label = th.Muted(pane.Strip(label))
	case r.Selected:
		label = th.Accent(label)
	}
	mark := Cursor(th, r.Selected)
	if r.Selected {
		mark = th.Accent(mark)
	}
	text := mark + r.Lead + label
	if r.Hint != "" {
		text += "  " + th.Muted(r.Hint)
	}
	lines := []string{text}
	if w > 0 {
		lines = pane.Wrap(text, w, pane.Width(Cursor(th, false)+r.Lead))
		for i := range lines {
			lines[i] = pane.Fit(lines[i], w)
		}
	}
	if r.Hovered {
		for i := range lines {
			lines[i] = th.RowHover(lines[i])
		}
	}
	return lines
}

// Line is a row on one line of w cells (0 for no limit): Lines, cut to its
// first line with an ellipsis when it would wrap.
//
// Parameters: th is the theme; r is the row; w is the width.
// Returns: the line.
func Line(th theme.Theme, r Row, w int) string {
	if w <= 0 {
		return Lines(th, r, 0)[0]
	}
	hovered := r.Hovered
	r.Hovered = false
	line := Lines(th, r, 0)[0]
	if pane.Width(line) > w {
		line = ansi.Truncate(line, w-1, "") + "…"
	}
	line = pane.Fit(line, w)
	if hovered {
		line = th.RowHover(line)
	}
	return line
}
