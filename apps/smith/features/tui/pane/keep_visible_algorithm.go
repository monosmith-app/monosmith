package pane

// KeepVisible is the first line of a window of h rows over total lines that
// shows line row, moving the window from top as little as it can. The rows
// that WindowMarked gives to its "more" markers (the first when lines are cut
// above, the last when they are cut below) never hold row, so a selection is
// never hidden under a marker.
//
// Parameters: top is where the window starts now; row is the line to show;
// total is how many lines there are; h is the rows of the window.
// Returns: the new first line, from 0 to total-h; 0 when everything fits.
func KeepVisible(top, row, total, h int) int {
	if h <= 0 || total <= h {
		return 0
	}
	top = ClampTop(top, total, h)
	row = Clamp(row, 0, total-1)
	if row < top+marker(top > 0) {
		top = max(row-1, 0)
	}
	if row > top+h-1-marker(top+h < total) {
		top = row - h + 2
		if top+h >= total {
			top = total - h
		}
	}
	return ClampTop(top, total, h)
}

func marker(on bool) int {
	if on {
		return 1
	}
	return 0
}
