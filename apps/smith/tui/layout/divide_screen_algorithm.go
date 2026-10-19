package layout

// Layout is where each area is drawn. An area that is folded away has an
// empty rectangle.
type Layout struct {
	Mode          Mode
	Width, Height int
	MenuBar       Rect
	Navigator     Rect
	Main          Rect
	Inspector     Rect
	StatusBar     Rect
}

// Divide lays the screen out: a one-row menu bar on top, a one-row status bar
// at the bottom, and between them the navigator, the main view and the
// inspector. The navigator is a fifth of the width (22 to 30 columns), the
// inspector a quarter (26 to 36).
func Divide(width, height int) Layout {
	l := Layout{Width: width, Height: height}
	switch {
	case width >= WideFrom:
		l.Mode = Wide
	case width >= MediumFrom:
		l.Mode = Medium
	default:
		l.Mode = Narrow
	}
	body := max(height-2, 0)
	l.MenuBar = Rect{0, 0, width, min(1, height)}
	if height > 1 {
		l.StatusBar = Rect{0, height - 1, width, 1}
	}
	nav, insp := 0, 0
	if l.Mode != Narrow {
		nav = clamp(width/5, 22, 30)
	}
	if l.Mode == Wide {
		insp = clamp(width/4, 26, 36)
	}
	l.Navigator = Rect{0, 1, nav, body}
	l.Main = Rect{nav, 1, width - nav - insp, body}
	l.Inspector = Rect{width - insp, 1, insp, body}
	return l
}

// AreaAt says which area the cell (x, y) belongs to.
func (l Layout) AreaAt(x, y int) Area {
	switch {
	case l.MenuBar.Contains(x, y):
		return MenuBar
	case l.StatusBar.Contains(x, y):
		return StatusBar
	case l.Navigator.Contains(x, y):
		return Navigator
	case l.Inspector.Contains(x, y):
		return Inspector
	case l.Main.Contains(x, y):
		return Main
	}
	return Nowhere
}

// Rect is the rectangle of an area.
func (l Layout) Rect(a Area) Rect {
	switch a {
	case MenuBar:
		return l.MenuBar
	case Navigator:
		return l.Navigator
	case Main:
		return l.Main
	case Inspector:
		return l.Inspector
	case StatusBar:
		return l.StatusBar
	}
	return Rect{}
}

// HasNavigator and HasInspector say whether those panes are on screen.
func (l Layout) HasNavigator() bool { return l.Navigator.W > 0 }
func (l Layout) HasInspector() bool { return l.Inspector.W > 0 }

func clamp(v, lo, hi int) int { return min(max(v, lo), hi) }
