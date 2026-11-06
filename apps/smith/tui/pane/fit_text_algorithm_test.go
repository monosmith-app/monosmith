package pane

import "testing"

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
