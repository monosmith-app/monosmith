package layout

// Layout is where each area is drawn. An area that is not on screen has an
// empty rectangle.
type Layout struct {
	Mode          Mode
	Width, Height int
	// Tall is true when the tabs are three rows high.
	Tall bool
	// Header is the workspace, the branch and the health; Tabs the row of views.
	Header, Tabs Rect
	// Main is the list (or the detail, when it has the whole width); Detail is
	// the pane beside it, present only in the wide layout.
	Main, Detail Rect
	Footer       Rect
	StatusBar    Rect
}

// Divide lays the screen out from the top: a header row, the row of tabs, the
// content, then a one-row footer of key hints and a one-row status bar. From
// TallFrom rows the tabs are drawn three rows high and the last of them is the
// top edge of the content; below it they are one row. Below WideFrom columns
// the content is one pane; from WideFrom it is the list and, when withDetail
// is true, the detail of the selection beside it (two fifths of the width, 38
// to 64 columns).
func Divide(width, height int, withDetail bool) Layout {
	l := Layout{Width: width, Height: height, Mode: Narrow, Tall: height >= TallFrom}
	if width >= WideFrom {
		l.Mode = Wide
	}
	if height > 1 {
		l.StatusBar = Rect{0, height - 1, width, 1}
	}
	if height > 2 {
		l.Footer = Rect{0, height - 2, width, 1}
	}
	if height > 4 {
		l.Header = Rect{0, 0, width, 1}
	}
	detail := 0
	if l.Mode == Wide && withDetail {
		detail = clamp(width*2/5, 38, 64)
	}
	// The content starts under the tabs; the top edge of a tall strip is the
	// content's own.
	tabsH, top := 1, 2
	if l.Tall {
		tabsH, top = 3, 3
	}
	l.Tabs = Rect{0, 1, width, tabsH}
	l.Main = Rect{0, top, width - detail, max(height-2-top, 0)}
	l.Detail = Rect{width - detail, top, detail, max(height-2-top, 0)}
	return l
}

// AreaAt says which area the cell (x, y) belongs to.
func (l Layout) AreaAt(x, y int) Area {
	switch {
	case l.Footer.Contains(x, y):
		return Footer
	case l.StatusBar.Contains(x, y):
		return StatusBar
	case l.Header.Contains(x, y):
		return Header
	case l.Tabs.Contains(x, y):
		return Tabs
	case l.Detail.Contains(x, y):
		return Detail
	case l.Main.Contains(x, y):
		return Main
	}
	return Nowhere
}

// DetailInner is where the detail is drawn: the pane has no left border of its
// own, the list's right border divides them.
func (l Layout) DetailInner() Rect {
	return Rect{X: l.Detail.X, Y: l.Detail.Y + 1, W: max(l.Detail.W-1, 0), H: max(l.Detail.H-2, 0)}
}

// HasDetail says whether the detail is on screen beside the list.
func (l Layout) HasDetail() bool { return l.Detail.W > 0 }

func clamp(v, lo, hi int) int { return min(max(v, lo), hi) }
