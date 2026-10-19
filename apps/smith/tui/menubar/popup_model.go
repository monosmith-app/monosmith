package menubar

import "github.com/monosmith-app/monosmith/apps/smith/tui/theme"

// Popup is a context menu opened with the right button at a cell.
type Popup struct {
	box   dropdown
	shown bool
}

// OpenPopup shows entries with the box's corner at (x, y), moved to stay
// inside a screen of w by h cells.
func OpenPopup(entries []Entry, x, y, w, h int) Popup {
	d := dropdown{entries: entries, minWidth: 14}
	d.X = max(min(x, w-d.innerWidth()-2), 0)
	d.Y = max(min(y, h-d.height()), 0)
	if len(entries) > 0 {
		d.cursor = d.step(-1, 1)
	}
	return Popup{box: d, shown: len(entries) > 0}
}

// Shown reports whether the popup is open.
func (p Popup) Shown() bool { return p.shown }

// Tooltip is the reason the highlighted entry is unavailable, or its hint.
func (p Popup) Tooltip() string {
	if !p.shown || p.box.cursor < 0 {
		return ""
	}
	e := p.box.entries[p.box.cursor]
	if e.Disabled != "" {
		return e.Name + " is unavailable: " + e.Disabled
	}
	return e.Hint
}

// Click handles a press: an entry inside the box is chosen, anywhere else closes it.
func (p Popup) Click(x, y int) (Popup, Result) {
	if j := p.box.entryAt(x, y); j >= 0 {
		e := p.box.entries[j]
		if e.Disabled != "" {
			p.box.cursor = j
			return p, Result{Handled: true, Refused: e.Name + " is unavailable: " + e.Disabled}
		}
		return Popup{}, Result{Handled: true, Chosen: e.ID}
	}
	if p.box.contains(x, y) {
		return p, Result{Handled: true}
	}
	return Popup{}, Result{Handled: true}
}

// Hover highlights the entry under the pointer.
func (p Popup) Hover(x, y int) Popup {
	if j := p.box.entryAt(x, y); j >= 0 {
		p.box.cursor = j
	}
	return p
}

// Key handles the arrows, Enter and Esc.
func (p Popup) Key(name string) (Popup, Result) {
	switch name {
	case "down":
		p.box.cursor = p.box.step(p.box.cursor, 1)
	case "up":
		p.box.cursor = p.box.step(p.box.cursor, -1)
	case "enter", " ":
		if p.box.cursor >= 0 {
			e := p.box.entries[p.box.cursor]
			if e.Disabled != "" {
				return p, Result{Handled: true, Refused: e.Name + " is unavailable: " + e.Disabled}
			}
			return Popup{}, Result{Handled: true, Chosen: e.ID}
		}
	case "esc", "ctrl+o":
		return Popup{}, Result{Handled: true}
	}
	return p, Result{Handled: true}
}

// Box draws the popup and says where it goes.
func (p Popup) Box(th theme.Theme) (lines []string, x, y int, ok bool) {
	if !p.shown {
		return nil, 0, 0, false
	}
	return p.box.view(th), p.box.X, p.box.Y, true
}
