package layout

// DivideFull lays the screen out for one view that takes the whole terminal: a
// single top row (the bar that closes it) and the content under it. The header,
// the footer and the status bar are off screen, so their areas are empty.
//
// Parameters: width and height are the terminal's size.
// Returns: the layout; Tabs is the top row and Main is everything under it.
func DivideFull(width, height int) Layout {
	l := Layout{Width: width, Height: height, Mode: Narrow}
	if width >= WideFrom {
		l.Mode = Wide
	}
	l.Tabs = Rect{0, 0, width, min(height, 1)}
	l.Main = Rect{0, 1, width, max(height-1, 0)}
	return l
}
