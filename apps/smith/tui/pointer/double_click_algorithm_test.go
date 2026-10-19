package pointer

import (
	"testing"
	"time"
)

func TestTwoQuickPressesOnTheSameCellAreADoubleClick(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	first := Click{X: 4, Y: 7, At: t0}
	if !IsDouble(first, Click{X: 4, Y: 7, At: t0.Add(200 * time.Millisecond)}) {
		t.Fatal("200ms apart on the same cell should be a double click")
	}
	if IsDouble(first, Click{X: 4, Y: 7, At: t0.Add(900 * time.Millisecond)}) {
		t.Fatal("900ms apart is two clicks")
	}
	if IsDouble(first, Click{X: 5, Y: 7, At: t0.Add(100 * time.Millisecond)}) {
		t.Fatal("another cell is a new click")
	}
	if IsDouble(Click{}, Click{X: 0, Y: 0, At: t0}) {
		t.Fatal("the first press ever is not a double click")
	}
}
