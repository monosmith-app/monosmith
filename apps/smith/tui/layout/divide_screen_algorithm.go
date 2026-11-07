package layout

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

// Layout is where each area is drawn. An area that is not on screen has an
// empty rectangle.
type Layout struct {
	Mode          Mode
	Width, Height int
	// Header is the workspace, the branch and the health; Tabs the row of views.
	Header, Tabs Rect
	// Main is the list (or the detail, when it has the whole width); Detail is
	// the pane beside it, present only in the wide layout.
	Main, Detail Rect
	Footer       Rect
	StatusBar    Rect
}

// Divide lays the screen out from the top: a header row, the row of tabs, the
// content, then a one-row footer of key hints and a one-row status bar. Below WideFrom columns
// the content is one pane; from WideFrom it is the list and, when withDetail
// is true, the detail of the selection beside it (two fifths of the width, 38
// to 64 columns).
func Divide(width, height int, withDetail bool) Layout {
	return DivideWith(width, height, withDetail, 0)
}

// The least and the most the user can make the detail, in columns; the list
// keeps at least MinList.
const (
	MinDetail = 28
	MinList   = 36
)

// AutoDetailWidth is the width of the detail the user has not resized.
func AutoDetailWidth(width int) int { return pane.Clamp(width*2/5, 38, 64) }

// ClampDetailWidth keeps a width the user chose between MinDetail and what
// leaves the list its MinList columns.
func ClampDetailWidth(width, detailW int) int {
	return pane.Clamp(detailW, MinDetail, max(width-MinList, MinDetail))
}

// DivideWith is Divide with a detail width the user chose (0 keeps the
// automatic one).
func DivideWith(width, height int, withDetail bool, detailW int) Layout {
	l := Layout{Width: width, Height: height, Mode: Narrow}
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
		detail = AutoDetailWidth(width)
		if detailW > 0 {
			detail = ClampDetailWidth(width, detailW)
		}
	}
	// The content starts under the one row of tabs.
	const top = 2
	l.Tabs = Rect{0, 1, width, 1}
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

// DetailInner is where the detail is drawn: the pane has no left margin of its
// own, the list's divider is its left edge (see DetailFrame).
func (l Layout) DetailInner() Rect { return DetailFrame.Inside(l.Detail) }

// HasDetail says whether the detail is on screen beside the list.
func (l Layout) HasDetail() bool { return l.Detail.W > 0 }
