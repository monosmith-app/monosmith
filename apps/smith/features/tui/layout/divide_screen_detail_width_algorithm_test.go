package layout

import "testing"

func TestAChosenDetailWidthReplacesTheAutomaticOne(t *testing.T) {
	l := DivideWith(140, 40, true, 70)
	if l.Detail.W != 70 || l.Main.W != 70 {
		t.Fatalf("%+v", l)
	}
	if got := DivideWith(140, 40, true, 0).Detail.W; got != AutoDetailWidth(140) {
		t.Fatalf("0 keeps the automatic width, got %d", got)
	}
}

func TestAChosenDetailWidthLeavesTheListItsMinimum(t *testing.T) {
	if got := DivideWith(120, 40, true, 500).Detail.W; got != 120-MinList {
		t.Fatalf("too wide: %d", got)
	}
	if got := DivideWith(120, 40, true, 3).Detail.W; got != MinDetail {
		t.Fatalf("too narrow: %d", got)
	}
}

func TestAChosenWidthDoesNotBringTheDetailBackBelow100Columns(t *testing.T) {
	if l := DivideWith(90, 40, true, 50); l.HasDetail() {
		t.Fatalf("%+v", l)
	}
}
