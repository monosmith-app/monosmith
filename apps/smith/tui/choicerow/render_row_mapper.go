package choicerow

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
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

// Render is a row of a choice list: the cursor mark, the tick box and the
// label. The cursor mark is the theme's pointer when selected, and blank
// cells of the same width when not, so a selection never moves the box.
func Render(th theme.Theme, label string, checked, selected, tri bool) string {
	return Cursor(th, selected) + Box(checked, tri) + " " + label
}
