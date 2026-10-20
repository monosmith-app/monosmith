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

// WindowTop is the first row of a window of size rows that keeps the cursor in
// the middle where it can.
func WindowTop(cursor, n, size int) int {
	return min(max(cursor-size/2, 0), max(n-size, 0))
}
