package layout

import "testing"

func TestAt140ColumnsTheRailTheListAndTheDetailShareTheWidth(t *testing.T) {
	l := Divide(140, 40, true)
	if l.Mode != Wide || !l.HasDetail() {
		t.Fatalf("mode %v detail %v", l.Mode, l.HasDetail())
	}
	if l.Rail.X != 0 || l.Rail.W != RailWidth || l.Main.X != RailWidth+1 {
		t.Fatalf("the rail is %d columns at the left, then a blank column: %+v", RailWidth, l)
	}
	if l.Main.W+l.Detail.W != ContentWidth(140) || l.Main.X+l.Main.W != l.Detail.X || l.Detail.X+l.Detail.W != 140 || l.Detail.W != 46 {
		t.Fatalf("the frames must fill the content width: %+v", l)
	}
	// The title bar is the first row, a blank row parts it from the frames, and
	// another parts the frames from the footer.
	if l.Header.Y != 0 || l.Header.H != 1 || l.Tabs.W != 0 || l.Main.Y != 2 || l.Main.H != 35 || l.Footer.Y != 38 || l.StatusBar.Y != 39 {
		t.Fatalf("rows: %+v", l)
	}
	if l.Rail.Y != l.Main.Y || l.Rail.H != l.Main.H || l.Detail.Y != l.Main.Y || l.Detail.H != l.Main.H {
		t.Fatalf("the rail and the frames share their rows: %+v", l)
	}
}

func TestAShortWideScreenHasNoBlankRows(t *testing.T) {
	l := Divide(120, gapFrom-1, true)
	if l.Header.Y != 0 || l.Main.Y != 1 || l.Main.Y+l.Main.H != l.Footer.Y {
		t.Fatalf("%+v", l)
	}
}

func TestANarrowScreenHasTheTitleBarThenTheStripOfViews(t *testing.T) {
	l := Divide(80, 30, true)
	if l.Header.Y != 0 || l.Tabs.Y != 1 || l.Tabs.H != 1 || l.Tabs.W != 80 || l.Rail.W != 0 {
		t.Fatalf("%+v", l)
	}
	if l.Main.X != 0 || l.Main.Y != 2 || l.Main.W != 80 || l.Main.Y+l.Main.H != l.Footer.Y {
		t.Fatalf("one frame from the strip to the footer: %+v", l)
	}
}

func TestBelow100ColumnsThereIsOnlyOnePane(t *testing.T) {
	for _, w := range []int{99, 90, 80, 72} {
		l := Divide(w, 30, true)
		if l.Mode != Narrow || l.HasDetail() || l.Main.W != w || l.Rail.W != 0 {
			t.Fatalf("width %d: %+v", w, l)
		}
	}
	if l := Divide(100, 30, true); l.Mode != Wide || l.Detail.W != 38 || l.Main.W != 37 || l.Rail.W != RailWidth {
		t.Fatalf("100 columns is the wide layout: %+v", l)
	}
}

func TestAViewWithoutADetailKeepsTheWholeContentWidth(t *testing.T) {
	if l := Divide(140, 40, false); l.HasDetail() || l.Main.W != ContentWidth(140) {
		t.Fatalf("%+v", l)
	}
}

func TestTheDetailWidthIsClamped(t *testing.T) {
	if got := Divide(300, 40, true).Detail.W; got != 68 {
		t.Fatalf("wide terminal: %d", got)
	}
}

func TestAreaAtFindsTheOwnerOfACell(t *testing.T) {
	l := Divide(140, 40, true)
	cases := []struct {
		x, y int
		want Area
	}{
		{5, 0, Header}, {5, 1, Nowhere}, {5, 2, Rail}, {24, 5, Nowhere}, {30, 2, Main}, {5, 38, Footer}, {5, 39, StatusBar}, {40, 5, Main}, {139, 5, Detail}, {200, 5, Nowhere},
	}
	for _, c := range cases {
		if got := l.AreaAt(c.x, c.y); got != c.want {
			t.Errorf("(%d,%d) = %v, want %v", c.x, c.y, got, c.want)
		}
	}
	if got := Divide(80, 30, true).AreaAt(5, 1); got != Tabs {
		t.Errorf("row 1 of a narrow screen = %v, want the tabs", got)
	}
}

func TestAnAbsurdlySmallScreenDoesNotPanic(t *testing.T) {
	for _, s := range [][2]int{{0, 0}, {1, 1}, {10, 2}, {80, 3}, {80, 4}, {120, 3}} {
		l := Divide(s[0], s[1], true)
		_ = l.AreaAt(0, 0)
	}
}

func TestEveryPaneIsABoxWithOneCellOfPadding(t *testing.T) {
	l := Divide(140, 40, true)
	in, din := l.Main.Inner(), l.DetailInner()
	if in.X != l.Main.X+2 || in.Y != l.Main.Y+1 || in.W != l.Main.W-4 || in.H != l.Main.H-2 {
		t.Fatalf("the list's inside is %+v of %+v", in, l.Main)
	}
	if din.X != l.Detail.X+2 || din.W != l.Detail.W-4 || din.Y != in.Y || din.H != in.H {
		t.Fatalf("the detail's inside is %+v of %+v", din, l.Detail)
	}
	if bottom := in.Y + in.H; bottom != l.Main.Y+l.Main.H-1 {
		t.Fatalf("the content ends above the bottom border: %d", bottom)
	}
}
