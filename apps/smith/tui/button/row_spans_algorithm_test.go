package button

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestRowSpansCentresTheButtonsOneCellApart(t *testing.T) {
	th := theme.Default()
	labels, keys := []string{"Run", "Stop"}, []string{"enter", "esc"}
	a, b := Width(th, "Run", "enter"), Width(th, "Stop", "esc")
	spans := RowSpans(th, labels, keys, 60)
	start := 2 + (60-(a+1+b))/2
	if spans[0] != [2]int{start, start + a} || spans[1] != [2]int{start + a + 1, start + a + 1 + b} {
		t.Fatalf("spans %v", spans)
	}
}

func TestRowSpansNeverStartsLeftOfThePadding(t *testing.T) {
	spans := RowSpans(theme.Default(), []string{"Run"}, []string{"enter"}, 2)
	if spans[0][0] != 2 {
		t.Fatalf("spans %v", spans)
	}
}

func TestAtFindsTheButtonUnderACellOnTheButtonRowOnly(t *testing.T) {
	rect := layout.Rect{X: 10, Y: 4, W: 40, H: 8}
	spans := [][2]int{{4, 12}, {13, 20}}
	y := rect.Y + rect.H - 2
	if At(spans, rect, 10+5, y) != 0 || At(spans, rect, 10+13, y) != 1 {
		t.Fatal("a cell on a button must find it")
	}
	if At(spans, rect, 10+12, y) != -1 {
		t.Fatal("the gap between two buttons is on neither")
	}
	if At(spans, rect, 10+5, y-1) != -1 {
		t.Fatal("another row is on no button")
	}
}
