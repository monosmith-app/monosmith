package dialog

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestAboutDrawsTheEmblemWhereThereIsRoomAndStaysSquare(t *testing.T) {
	for _, height := range []int{24, 36, 50} {
		a := NewAbout("1.0.0").Place(100, height)
		lines := a.Lines(theme.Default())
		if len(lines) != a.Rect().H {
			t.Fatalf("height %d: %d lines for a box of %d", height, len(lines), a.Rect().H)
		}
		for i, l := range lines {
			if pane.Width(l) != a.Rect().W {
				t.Errorf("height %d row %d is %d wide, the box is %d", height, i, pane.Width(l), a.Rect().W)
			}
		}
	}
}

func TestAboutLogoGrowsWithTheTerminal(t *testing.T) {
	th := theme.Default()
	anvil := len(NewAbout("").Place(100, 24).logo(th))
	small := len(NewAbout("").Place(100, 36).logo(th))
	large := len(NewAbout("").Place(100, 50).logo(th))
	if anvil != len(aboutLogo) || small != 8 || large != 16 {
		t.Errorf("logo rows: %d, %d, %d; want %d, 8, 16", anvil, small, large, len(aboutLogo))
	}
}
