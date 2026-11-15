package pane

import "testing"

// A tool line the TUI shows in colour is measured, cut and padded on its plain
// text: the colour codes take no cell.
func TestAColouredToolLineIsMeasuredWithoutItsCodes(t *testing.T) {
	line := "web:test │ \x1b[31m--- FAIL: TestX\x1b[0m"
	if got := Width(line); got != Width("web:test │ --- FAIL: TestX") {
		t.Errorf("Width %d", got)
	}
	if got := Strip(Fit(line, 40)); got != "web:test │ --- FAIL: TestX"+"              " {
		t.Errorf("Fit pads on the plain text: %q", got)
	}
	if got := Strip(Cut(line, 15)); got != "web:test │ --- " {
		t.Errorf("Cut counts cells, not bytes: %q", got)
	}
	if got := Fit(line, 15); got[len(got)-3:] != "\x1b[m" {
		t.Errorf("a cut colour must be reset before the padding: %q", got)
	}
}

func TestCutEllipsisEndsACutNameInAnEllipsisAndLeavesAShortOneAlone(t *testing.T) {
	if got := CutEllipsis("abcdefgh", 5); got != "abcd…" {
		t.Fatalf("got %q", got)
	}
	if got := CutEllipsis("abc", 5); got != "abc" {
		t.Fatalf("got %q", got)
	}
	if got := CutEllipsis("abc", 0); got != "" {
		t.Fatalf("got %q", got)
	}
}
