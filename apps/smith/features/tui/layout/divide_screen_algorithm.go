package layout

import (
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

// Layout is where each area is drawn. An area that is not on screen has an
// empty rectangle.
type Layout struct {
	Mode          Mode
	Width, Height int
	// Header is the title bar: the first row at every width.
	Header Rect
	// Tabs is the one-line strip of views under the title bar of a narrow
	// screen; empty on a wide one, where the rail lists the views.
	Tabs Rect
	// Rail is the frame of views at the left of a wide screen; empty on a
	// narrow one.
	Rail Rect
	// Main is the frame of the list (or of the detail, when it has the whole
	// width); Detail is the frame beside it, present only in the wide layout.
	Main, Detail Rect
	Footer       Rect
	StatusBar    Rect
}

// RailWidth is the width of the rail of views, its frame included. The rail
// is not resizable.
const RailWidth = 24

// railGap is the blank column between the rail and the content frames.
const railGap = 1

// gapFrom is the height from which a wide screen keeps a blank row between
// the title bar and the frames, and between the frames and the footer.
const gapFrom = 20

// Divide lays the screen out from the top: the title bar, the strip of views
// on a narrow screen, the content frames, then a one-row footer of key hints
// and a one-row status bar. From WideFrom columns the rail of views sits at
// the left, and the content is the list and, when withDetail is true, the
// detail of the selection beside it (two fifths of the content width, 38 to
// 68 columns). Below WideFrom the content is one frame.
//
// Parameters: width and height are the terminal's size; withDetail asks for
// the detail beside the list.
// Returns: the layout.
func Divide(width, height int, withDetail bool) Layout {
	return DivideWith(width, height, withDetail, 0)
}

// The least and the most the user can make the detail, in columns; the list
// keeps at least MinList.
const (
	MinDetail = 28
	MinList   = 36
)

// AutoDetailWidth is the width of the detail the user has not resized. Its
// most, 68, keeps 64 columns for the text inside the frame and its padding.
//
// Parameters: width is the width of the content (ContentWidth).
// Returns: the detail's width in columns.
func AutoDetailWidth(width int) int { return pane.Clamp(width*2/5, 38, 68) }

// ClampDetailWidth keeps a width the user chose between MinDetail and what
// leaves the list its MinList columns.
//
// Parameters: width is the width of the content (ContentWidth); detailW is
// the width the user chose.
// Returns: the width the detail gets.
func ClampDetailWidth(width, detailW int) int {
	return pane.Clamp(detailW, MinDetail, max(width-MinList, MinDetail))
}

// ContentWidth is the width the list and the detail share on a screen width
// columns wide: all of it on a narrow screen, what the rail leaves on a wide one.
//
// Parameters: width is the terminal's width.
// Returns: the content width in columns.
func ContentWidth(width int) int {
	if width >= WideFrom {
		return max(width-RailWidth-railGap, 0)
	}
	return width
}

// DivideWith is Divide with a detail width the user chose (0 keeps the
// automatic one).
//
// Parameters: width and height are the terminal's size; withDetail asks for
// the detail; detailW is the chosen detail width, 0 for the automatic one.
// Returns: the layout.
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
	top, bottom := 0, max(height-2, 0)
	if height > 4 {
		l.Header = Rect{0, 0, width, 1}
		top = 1
	}
	x, cw := 0, width
	if l.Mode == Wide {
		if height >= gapFrom {
			top, bottom = top+1, bottom-1
		}
		x, cw = RailWidth+railGap, ContentWidth(width)
	} else {
		l.Tabs = Rect{0, top, width, 1}
		top++
	}
	h := max(bottom-top, 0)
	detail := 0
	if l.Mode == Wide && withDetail {
		detail = AutoDetailWidth(cw)
		if detailW > 0 {
			detail = ClampDetailWidth(cw, detailW)
		}
	}
	if l.Mode == Wide {
		l.Rail = Rect{0, top, RailWidth, h}
	}
	l.Main = Rect{x, top, cw - detail, h}
	l.Detail = Rect{x + cw - detail, top, detail, h}
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
	case l.Rail.Contains(x, y):
		return Rail
	case l.Detail.Contains(x, y):
		return Detail
	case l.Main.Contains(x, y):
		return Main
	}
	return Nowhere
}

// DetailInner is where the detail is drawn: inside its frame (see DetailFrame).
func (l Layout) DetailInner() Rect { return DetailFrame.Inside(l.Detail) }

// HasDetail says whether the detail is on screen beside the list.
func (l Layout) HasDetail() bool { return l.Detail.W > 0 }

// InlineHeaderFrom was the width from which the header shared the row of the
// tabs. The title bar now always has its own row, so no width puts it inline;
// the constant stays so the layout package keeps its exported names.
const InlineHeaderFrom = 0

// DivideShared is DivideWith. The header no longer shares the row of the tabs,
// so inline is ignored.
//
// Parameters: width and height are the terminal's size; withDetail asks for
// the detail; detailW is the chosen detail width, 0 for the automatic one;
// inline is ignored.
// Returns: the layout.
func DivideShared(width, height int, withDetail bool, detailW int, inline bool) Layout {
	_ = inline
	return DivideWith(width, height, withDetail, detailW)
}
