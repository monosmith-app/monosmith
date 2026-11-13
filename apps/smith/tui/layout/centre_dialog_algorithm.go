package layout

// Margins are the rows a dialog leaves free at the top of the screen (the
// header and the tab strip) and at the bottom (the footer and the status bar),
// so every dialog sits in the same place: over the content, never over the tabs
// or the footer.
type Margins struct{ Top, Bottom int }

// Room is the height between the margins on a screen sh rows tall, never
// below 0.
//
// Parameters: sh is the height of the screen.
// Returns: the rows a dialog may take.
func (m Margins) Room(sh int) int { return max(sh-m.Top-m.Bottom, 0) }

// Centre places a dialog of w by h cells on a screen of sw by sh cells: in the
// middle across, and in the middle of the rows between the margins. A dialog
// taller than that room gets the room's height, so it never covers the tab
// strip or the footer: the dialog draws itself in the Rect's height and
// scrolls its body (pane.DialogBody). Only a screen with no room at all keeps
// the height asked for, and then the dialog still ends on the screen.
//
// Parameters: sw and sh are the screen's size; w and h the dialog's.
// Returns: where the dialog sits.
func (m Margins) Centre(sw, sh, w, h int) Rect {
	if room := m.Room(sh); room > 0 {
		h = min(h, room)
	}
	y := m.Top + max((m.Room(sh)-h)/2, 0)
	if y+h > sh {
		y = max(sh-h, 0)
	}
	return Rect{X: max((sw-w)/2, 0), Y: y, W: w, H: h}
}
