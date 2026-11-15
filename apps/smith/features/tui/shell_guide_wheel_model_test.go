package tui

import "testing"

func TestTheWheelScrollsAGuideListAndNeverMovesTheSelection(t *testing.T) {
	for _, size := range [][2]int{{110, 32}, {80, 24}} {
		l := newLab(t, true, size[0], size[1])
		l.keys("a")
		l.has("What are you building?")
		before := l.view()
		for range 4 {
			l.send(wheel(10, 8, 1))
		}
		for range 4 {
			l.send(wheel(10, 8, -1))
		}
		if got := l.view(); got != before {
			t.Fatalf("%dx%d: the wheel changed the screen:\nbefore:\n%s\nafter:\n%s", size[0], size[1], before, got)
		}
	}
}
