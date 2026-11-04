package pane

import "testing"

func TestBlankKeepsTheFrameAndEmptiesBoxDrawingInTheContent(t *testing.T) {
	rows := []string{
		"╭──┬───╮",
		"│a─ │┃  │", // a rule and a scrollbar thumb in the content
		"│世┃│界 │",   // a wide glyph takes two columns
		"╰──┴───╯",
	}
	got := Blank(rows, 0, 4, 0, 4, 8)
	want := []string{
		"╭──┬───╮",
		"│   │   │",
		"│   │   │",
		"╰──┴───╯",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d = %q, want %q", i, got[i], want[i])
		}
	}
}
