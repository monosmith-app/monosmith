package splitter

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
)

func TestRunListRowsKeepsTheLogItsRows(t *testing.T) {
	for _, c := range []struct{ given, avail, want int }{
		{0, 10, MinRunListRows},
		{5, 10, 5},
		{99, 10, 10 - MinLogRows},
		{5, 1, MinRunListRows},
	} {
		if got := RunListRows(c.given, c.avail); got != c.want {
			t.Errorf("RunListRows(%d, %d) = %d, want %d", c.given, c.avail, got, c.want)
		}
	}
}

func TestSpotsAreBuiltFromPlainValues(t *testing.T) {
	in := layout.Rect{X: 2, Y: 3, W: 60, H: 20}
	run := RunListSpot(in, 3, 8, 5)
	if run.ID != RunListID || run.Axis != Row || run.Hit.Y != 11 || run.Origin != 6 || run.Sign != 1 || run.Size != 5 {
		t.Fatalf("run list spot %+v", run)
	}
	list := SettingsListSpot(in, 20, 10, 30)
	if list.Axis != Column || list.Hit.X != 22 || list.Hit.H != 20 || list.Origin != 2 {
		t.Fatalf("settings list spot %+v", list)
	}
	detail := SettingsDetailSpot(in, 20, 4, 6, 3, 9)
	if detail.Axis != Row || detail.Hit.X != 23 || detail.Hit.W != 39 || detail.Origin != 13 || detail.Sign != -1 {
		t.Fatalf("settings detail spot %+v", detail)
	}
	l := layout.DivideWith(140, 30, true, 0)
	d := DetailSpot("doctor", l, 140)
	if d.ID != "doctor.detail" || d.Hit.X != l.Detail.X-1 || d.Size != l.Detail.W || d.Origin != 139 || d.Sign != -1 {
		t.Fatalf("detail spot %+v", d)
	}
}
