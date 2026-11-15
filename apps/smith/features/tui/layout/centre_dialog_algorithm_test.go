package layout

import "testing"

func TestADialogIsCentredBetweenTheMarginsAndNeverLeavesTheScreen(t *testing.T) {
	m := Margins{Top: 3, Bottom: 2}
	cases := []struct {
		name         string
		sw, sh, w, h int
		want         Rect
	}{
		{"fits", 100, 30, 40, 10, Rect{X: 30, Y: 3 + (25-10)/2, W: 40, H: 10}},
		{"fills the room", 100, 30, 40, 25, Rect{X: 30, Y: 3, W: 40, H: 25}},
		{"taller than the room gets the room", 100, 30, 40, 29, Rect{X: 30, Y: 3, W: 40, H: 25}},
		{"wider than the screen starts at 0", 30, 30, 40, 5, Rect{X: 0, Y: 13, W: 40, H: 5}},
	}
	for _, c := range cases {
		if got := m.Centre(c.sw, c.sh, c.w, c.h); got != c.want {
			t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
		}
	}
	if got := m.Centre(100, 4, 40, 6); got != (Rect{X: 30, Y: 0, W: 40, H: 6}) {
		t.Errorf("no room keeps the height and ends on the screen: %+v", got)
	}
	if got := m.Room(4); got != 0 {
		t.Errorf("a screen shorter than the margins has no room: %d", got)
	}
}
