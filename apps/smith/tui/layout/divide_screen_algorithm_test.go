package layout

import "testing"

func TestAt140ColumnsTheListAndTheDetailShareTheWidth(t *testing.T) {
	l := Divide(140, 40, true)
	if l.Mode != Wide || !l.HasDetail() {
		t.Fatalf("mode %v detail %v", l.Mode, l.HasDetail())
	}
	if l.Main.W+l.Detail.W != 140 || l.Detail.W != 56 {
		t.Fatalf("panes must fill the width: %+v", l)
	}
	// Tall tabs: three rows, the last one is the content's top edge.
	if !l.Tall || l.Header.Y != 0 || l.Tabs.Y != 1 || l.Tabs.H != 3 || l.Main.Y != 3 || l.Footer.Y != 38 || l.StatusBar.Y != 39 || l.Main.H != 35 {
		t.Fatalf("rows: %+v", l)
	}
}

func TestOnAShortTerminalTheTabsAreOneRow(t *testing.T) {
	l := Divide(120, 23, true)
	if l.Tall || l.Tabs.H != 1 || l.Main.Y != 2 || l.Main.H != 19 {
		t.Fatalf("%+v", l)
	}
	if Divide(120, 24, true).Tall != true {
		t.Fatal("24 rows is tall enough")
	}
}

func TestBelow100ColumnsThereIsOnlyOnePane(t *testing.T) {
	for _, w := range []int{99, 90, 80, 72} {
		l := Divide(w, 30, true)
		if l.Mode != Narrow || l.HasDetail() || l.Main.W != w {
			t.Fatalf("width %d: %+v", w, l)
		}
	}
	if l := Divide(100, 30, true); l.Mode != Wide || l.Detail.W != 40 || l.Main.W != 60 {
		t.Fatalf("100 columns is the wide layout: %+v", l)
	}
}

func TestAViewWithoutADetailKeepsTheWholeWidth(t *testing.T) {
	if l := Divide(140, 40, false); l.HasDetail() || l.Main.W != 140 {
		t.Fatalf("%+v", l)
	}
}

func TestTheDetailWidthIsClamped(t *testing.T) {
	if got := Divide(300, 40, true).Detail.W; got != 64 {
		t.Fatalf("wide terminal: %d", got)
	}
}

func TestAreaAtFindsTheOwnerOfACell(t *testing.T) {
	l := Divide(140, 40, true)
	cases := []struct {
		x, y int
		want Area
	}{
		{5, 0, Header}, {5, 1, Tabs}, {5, 3, Tabs}, {5, 38, Footer}, {5, 39, StatusBar}, {20, 5, Main}, {139, 5, Detail}, {200, 5, Nowhere},
	}
	for _, c := range cases {
		if got := l.AreaAt(c.x, c.y); got != c.want {
			t.Errorf("(%d,%d) = %v, want %v", c.x, c.y, got, c.want)
		}
	}
}

func TestAnAbsurdlySmallScreenDoesNotPanic(t *testing.T) {
	for _, s := range [][2]int{{0, 0}, {1, 1}, {10, 2}, {80, 3}, {80, 4}} {
		l := Divide(s[0], s[1], true)
		_ = l.AreaAt(0, 0)
	}
}
