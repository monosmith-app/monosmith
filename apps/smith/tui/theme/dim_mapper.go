package theme

import "github.com/charmbracelet/x/ansi"

// Dim redraws one screen row in the faint style with every colour of the row
// removed, so a screen behind a dialog shows nothing that looks interactive.
// In a theme without colour the row gets the faint attribute (SGR 2) when the
// terminal prints styles at all; where styles print nothing the row is returned
// as it is.
//
// Parameters: row is one screen row, with or without styles.
// Returns: the dimmed row; it has the same cells as row.
func (t Theme) Dim(row string) string {
	plain := ansi.Strip(row)
	if styled := t.faint.Render(plain); styled != plain {
		return styled
	}
	if t.selected.Render("x") != "x" {
		return "\x1b[2m" + plain + "\x1b[m"
	}
	return row
}
