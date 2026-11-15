package cipipeline

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// maxTableRows is how many rows a log table shows before it counts the rest.
const maxTableRows = 30

// tableRow is one line of a result table; Bad rows (a failure, a warning)
// always show.
type tableRow struct {
	Cells []string
	Bad   bool
	// Tints colours cells by index; a cell without one takes the colour of its
	// leading status glyph.
	Tints map[int]tint
}

// resultTable is the same rows as aligned text for the log and as Markdown for
// the job summary.
type resultTable struct {
	Headers []string
	// Right marks the columns aligned to the right (numbers, durations).
	Right []bool
	Rows  []tableRow
}

// shown is the rows to print: all of them up to maxTableRows, else the bad
// ones and the first of the rest in order, and how many were left out.
func (t resultTable) shown() (rows []tableRow, left int) {
	if len(t.Rows) <= maxTableRows {
		return t.Rows, 0
	}
	room := maxTableRows
	for _, r := range t.Rows {
		if r.Bad {
			room--
		}
	}
	for _, r := range t.Rows {
		if r.Bad {
			rows = append(rows, r)
		} else if room > 0 {
			rows = append(rows, r)
			room--
		}
	}
	return rows, len(t.Rows) - len(rows)
}

// Text is the table as padded columns (spaces, no box drawing), then a line
// counting the rows left out; more names what they were ("passed"). Colour
// follows colourWanted.
func (t resultTable) Text(more string) []string {
	return t.render(more, colourOn())
}

// render pads on the plain text and paints after, so colour never moves a column.
func (t resultTable) render(more string, colour bool) []string {
	rows, left := t.shown()
	widths := make([]int, len(t.Headers))
	for i, h := range t.Headers {
		widths[i] = utf8.RuneCountInString(h)
	}
	for _, r := range rows {
		for i, c := range r.Cells {
			widths[i] = max(widths[i], utf8.RuneCountInString(c))
		}
	}
	line := func(cells []string, tints map[int]tint) string {
		parts := make([]string, len(cells))
		for i, c := range cells {
			pad := strings.Repeat(" ", widths[i]-utf8.RuneCountInString(c))
			k, ok := tints[i]
			if !ok {
				k = glyphTint(c)
			}
			c = paint(c, k, colour)
			if i < len(t.Right) && t.Right[i] {
				parts[i] = pad + c
			} else {
				parts[i] = c + pad
			}
		}
		return strings.TrimRight(strings.Join(parts, "  "), " ")
	}
	lines := []string{line(t.Headers, nil)}
	for _, r := range rows {
		lines = append(lines, line(r.Cells, r.Tints))
	}
	if left > 0 {
		lines = append(lines, fmt.Sprintf("... %d more %s", left, more))
	}
	return lines
}

// Markdown is the table for a job summary.
func (t resultTable) Markdown() []string {
	row := func(cells []string) string {
		out := make([]string, len(cells))
		for i, c := range cells {
			out[i] = strings.ReplaceAll(c, "|", "\\|")
		}
		return "| " + strings.Join(out, " | ") + " |"
	}
	sep := make([]string, len(t.Headers))
	for i := range sep {
		sep[i] = "---"
		if i < len(t.Right) && t.Right[i] {
			sep[i] = "---:"
		}
	}
	lines := []string{row(t.Headers), "| " + strings.Join(sep, " | ") + " |"}
	for _, r := range t.Rows {
		lines = append(lines, row(r.Cells))
	}
	return lines
}
