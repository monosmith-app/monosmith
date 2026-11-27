package tui

import (
	"math"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

// barDrag is a press on a scrollbar that the mouse still holds: the column of
// the bar, its first row, and the row of the thumb the pointer holds.
type barDrag struct {
	x, top, grab int
}

// barAt is the scrollbar the screen shows at (x, y), the dialog over it
// included.
func (s Shell) barAt(x, y int) (pane.Bar, bool) {
	rows := s.selectableRows()
	cell := func(cx, cy int) string {
		if cy < 0 || cy >= len(rows) || cx < 0 {
			return ""
		}
		return pane.Strip(ansi.Cut(rows[cy], cx, cx+1))
	}
	return pane.BarAt(cell, x, y, s.th.Icons.Track, s.th.Icons.Thumb)
}

// scrollbarMouse handles the mouse on a scrollbar as a Mac does: a press on
// the track moves the thumb there, centred on the pointer; a press on the
// thumb holds it; a drag moves it with the pointer until the release. It
// reports whether it took the event.
func (s *Shell) scrollbarMouse(e tea.MouseEvent) (tea.Cmd, bool) {
	switch {
	case s.drag != nil && e.Action == tea.MouseActionMotion && e.Button == tea.MouseButtonLeft:
		d := *s.drag
		return s.moveThumb(d.x, d.top, e.Y-d.top-d.grab), true
	case s.drag != nil && (e.Action == tea.MouseActionMotion || e.Action == tea.MouseActionRelease):
		// A motion with no button held means the terminal lost the release.
		s.drag = nil
		return nil, true
	case e.Action != tea.MouseActionPress || e.Button != tea.MouseButtonLeft:
		return nil, false
	}
	b, ok := s.barAt(e.X, e.Y)
	if !ok {
		return nil, false
	}
	s.selection = s.selection.Clear()
	grab := e.Y - b.Top - b.At
	if grab < 0 || grab >= b.Size {
		grab = b.Size / 2
	}
	s.drag = &barDrag{x: e.X, top: b.Top, grab: grab}
	return s.moveThumb(e.X, b.Top, e.Y-b.Top-grab), true
}

// moveThumb scrolls the pane whose bar is in column x, starting at row top,
// until its thumb starts at cell want (kept inside the bar). It turns the
// wheel of the pane, so each pane keeps its own rules, and measures the bar
// again after each turn.
func (s *Shell) moveThumb(x, top, want int) tea.Cmd {
	var cmds []tea.Cmd
	perCell := 0.0
	for range 12 {
		b, ok := s.barAt(x, top)
		if !ok || b.Top != top {
			break
		}
		want = max(0, min(want, b.H-b.Size))
		if b.At == want {
			break
		}
		if perCell == 0 {
			// The thumb is h*h/total cells high: guess the rows behind one cell.
			total := float64(b.H*b.H) / float64(b.Size)
			perCell = max((total-float64(b.H))/float64(b.H-b.Size)/wheelStep, 1)
		}
		dir := 1
		if want < b.At {
			dir = -1
		}
		turns := max(int(math.Round(math.Abs(float64(want-b.At))*perCell)), 1)
		for range turns {
			cmds = append(cmds, s.wheel(x, top, dir))
		}
		after, ok := s.barAt(x, top)
		if !ok || after.At == b.At {
			break
		}
		perCell = max(float64(turns)/math.Abs(float64(after.At-b.At)), 0.1)
	}
	return tea.Batch(cmds...)
}
