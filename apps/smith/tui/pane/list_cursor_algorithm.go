package pane

import "github.com/monosmith-app/monosmith/apps/smith/tui/commands"

// MoveCursor is where a list cursor goes for a navigation action in a list of
// n rows; page is how many rows a page action moves. Other actions leave it.
func MoveCursor(cursor, n int, action commands.NavAction, page int) int {
	if n == 0 {
		return 0
	}
	switch action {
	case commands.NavUp:
		cursor--
	case commands.NavDown:
		cursor++
	case commands.NavPageUp:
		cursor -= max(page, 1)
	case commands.NavPageDown:
		cursor += max(page, 1)
	case commands.NavHome:
		cursor = 0
	case commands.NavEnd:
		cursor = n - 1
	}
	return min(max(cursor, 0), n-1)
}

// ScrollTo is the first row of a window of h rows, given the row top it shows
// now, that keeps row sel in view and moves no further than it must: a
// selection already in view leaves the window where it is. n is the row count.
func ScrollTo(top, sel, h, n int) int {
	switch {
	case sel < top:
		top = sel
	case sel >= top+h:
		top = sel - h + 1
	}
	return min(max(top, 0), max(n-h, 0))
}
