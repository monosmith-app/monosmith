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
	// The tabs are one row and, on a wide screen, the first: the header shares
	// it. The content starts under them.
	if l.Header.W != 0 || l.Tabs.Y != 0 || l.Tabs.H != 1 || l.Main.Y != 1 || l.Footer.Y != 38 || l.StatusBar.Y != 39 || l.Main.H != 37 {
		t.Fatalf("rows: %+v", l)
	}
}

func TestTheTabsAreOneRowOnAnyHeight(t *testing.T) {
	l := Divide(120, 23, true)
	if l.Tabs.H != 1 || l.Main.Y != 1 || l.Main.H != 20 {
		t.Fatalf("%+v", l)
	}
	if tall := Divide(120, 60, true); tall.Tabs.H != 1 || tall.Main.Y != 1 {
		t.Fatalf("a tall terminal keeps one row of tabs: %+v", tall)
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
		{5, 0, Tabs}, {5, 1, Main}, {5, 2, Main}, {5, 38, Footer}, {5, 39, StatusBar}, {20, 5, Main}, {139, 5, Detail}, {200, 5, Nowhere},
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

func TestTheMainFrameIsARuleAndTwoMarginsAndNoBottom(t *testing.T) {
	l := Divide(140, 40, true)
	in, din := l.Main.Inner(), l.DetailInner()
	if in.X != 1 || in.Y != l.Main.Y+1 || in.W != l.Main.W-2 || in.H != l.Main.H-1 {
		t.Fatalf("the list's inside is %+v of %+v", in, l.Main)
	}
	// The detail starts at the divider's neighbour, keeps a right margin and
	// ends on the same last row as the list.
	if din.X != l.Detail.X || din.W != l.Detail.W-1 || din.Y != in.Y || din.H != in.H {
		t.Fatalf("the detail's inside is %+v of %+v", din, l.Detail)
	}
	if bottom := in.Y + in.H; bottom != l.Footer.Y {
		t.Fatalf("the content ends above the footer: %d, footer at %d", bottom, l.Footer.Y)
	}
}

func TestBelowTheInlineWidthTheHeaderHasARowOfItsOwn(t *testing.T) {
	l := Divide(InlineHeaderFrom-1, 40, false)
	if l.Header.Y != 0 || l.Header.H != 1 || l.Tabs.Y != 1 || l.Main.Y != 2 {
		t.Fatalf("a narrow screen keeps the header row: %+v", l)
	}
	if wide := Divide(InlineHeaderFrom, 40, false); wide.Header.W != 0 || wide.Tabs.Y != 0 || wide.Main.Y != 1 {
		t.Fatalf("from the inline width the header joins the tabs: %+v", wide)
	}
}
