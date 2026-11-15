package shellcompletion

import (
	"strings"
	"testing"
)

func TestABlockRoundTripsExactly(t *testing.T) {
	for _, original := range []string{"", "alias a=b\n", "alias a=b", "one\ntwo\n\nthree\n", "one\r\ntwo\r\n"} {
		with := WithBlock(original, "line one\nline two")
		if !HasBlock(with) {
			t.Fatalf("%q: no block after WithBlock", original)
		}
		got, found := WithoutBlock(with)
		want := original
		if want != "" && !strings.HasSuffix(want, "\n") {
			want += "\n" // the block adds the missing line end of the last line
		}
		if !found || got != want {
			t.Errorf("%q: got %q, want %q", original, got, want)
		}
	}
}

func TestAnExistingBlockIsReplacedInPlace(t *testing.T) {
	first := WithBlock("before\n", "old")
	text := first + "after\n"
	next := WithBlock(text, "new")
	if strings.Count(next, beginMarker) != 1 || !strings.Contains(next, "new") || strings.Contains(next, "old") {
		t.Fatalf("block not replaced once: %q", next)
	}
	if !strings.HasPrefix(next, "before\n") || !strings.HasSuffix(next, "after\n") {
		t.Fatalf("the lines around the block changed: %q", next)
	}
	if body, _ := BlockBody(next); body != "new" {
		t.Fatalf("body %q", body)
	}
}

func TestWindowsLineEndsAreKept(t *testing.T) {
	text := WithBlock("a\r\nb\r\n", "x\ny")
	if strings.Contains(strings.ReplaceAll(text, "\r\n", ""), "\n") {
		t.Fatalf("a bare line feed in a CRLF file: %q", text)
	}
}

func TestABlockThatIsNotClosedIsNotABlock(t *testing.T) {
	if HasBlock(beginMarker + "\nnever closed\n") {
		t.Fatal("an unclosed block counts as installed")
	}
	if _, found := WithoutBlock("plain\n"); found {
		t.Fatal("found a block in plain text")
	}
}
