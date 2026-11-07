package pane

import "testing"

func TestBlankKeepsTheRuleAndEmptiesBoxDrawingInTheContent(t *testing.T) {
	rows := []string{
		"────┬───",
		" a─ │┃  ", // a rule and a scrollbar thumb in the content
		" 世┃│界 ",   // a wide glyph takes two columns
		" ─┃ │ ─ ", // the last row is content too: there is no bottom edge
	}
	got := Blank(rows, 0, 4, 4)
	want := []string{
		"────┬───",
		"    │   ",
		"    │   ",
		"    │   ",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d = %q, want %q", i, got[i], want[i])
		}
	}
}
