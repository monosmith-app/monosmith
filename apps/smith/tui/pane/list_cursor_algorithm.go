package pane

// MoveCursor is where a list cursor goes for a key in a list of n rows; page
// is how many rows a page key moves.
func MoveCursor(cursor, n int, key string, page int) int {
	if n == 0 {
		return 0
	}
	switch key {
	case "up", "k":
		cursor--
	case "down", "j":
		cursor++
	case "pgup":
		cursor -= max(page, 1)
	case "pgdown":
		cursor += max(page, 1)
	case "home":
		cursor = 0
	case "end":
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
