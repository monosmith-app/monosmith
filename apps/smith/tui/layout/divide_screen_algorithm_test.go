package layout

import "testing"

func TestTheLayoutAt140ColumnsHasThreePanes(t *testing.T) {
	l := Divide(140, 40)
	if l.Mode != Wide || !l.HasNavigator() || !l.HasInspector() {
		t.Fatalf("mode %v nav %v insp %v", l.Mode, l.HasNavigator(), l.HasInspector())
	}
	if l.Navigator.W+l.Main.W+l.Inspector.W != 140 {
		t.Fatalf("panes must fill the width: %+v", l)
	}
	if l.Footer.Y != 38 || l.StatusBar.Y != 39 || l.Main.H != 38 {
		t.Fatalf("rows: %+v", l)
	}
}

func TestTheInspectorFoldsIntoTheMainViewBelow100Columns(t *testing.T) {
	for _, w := range []int{99, 90, 80} {
		l := Divide(w, 30)
		if l.Mode != Medium || l.HasInspector() || !l.HasNavigator() {
			t.Fatalf("width %d: mode %v nav %v insp %v", w, l.Mode, l.HasNavigator(), l.HasInspector())
		}
		if l.Navigator.W+l.Main.W != w {
			t.Fatalf("width %d: %+v", w, l)
		}
	}
	if l := Divide(100, 30); l.Mode != Wide || !l.HasInspector() {
		t.Fatalf("100 columns is still the wide layout: %+v", l)
	}
}

func TestTheNavigatorBecomesATabBelow80Columns(t *testing.T) {
	l := Divide(79, 24)
	if l.Mode != Narrow || l.HasNavigator() || l.HasInspector() || l.Main.W != 79 {
		t.Fatalf("narrow: %+v", l)
	}
}

func TestAtTheSupportedFloorTheMainViewIsStillUsable(t *testing.T) {
	l := Divide(80, 24)
	if l.Main.W < 50 {
		t.Fatalf("main is %d columns at 80", l.Main.W)
	}
}

func TestAreaAtFindsTheOwnerOfACell(t *testing.T) {
	l := Divide(140, 40)
	cases := []struct {
		x, y int
		want Area
	}{
		{5, 38, Footer}, {5, 39, StatusBar}, {2, 5, Navigator}, {70, 5, Main}, {139, 5, Inspector}, {200, 5, Nowhere},
	}
	for _, c := range cases {
		if got := l.AreaAt(c.x, c.y); got != c.want {
			t.Errorf("(%d,%d) = %v, want %v", c.x, c.y, got, c.want)
		}
	}
}

func TestAnAbsurdlySmallScreenDoesNotPanic(t *testing.T) {
	for _, s := range [][2]int{{0, 0}, {1, 1}, {10, 2}, {80, 3}} {
		l := Divide(s[0], s[1])
		_ = l.AreaAt(0, 0)
	}
}
