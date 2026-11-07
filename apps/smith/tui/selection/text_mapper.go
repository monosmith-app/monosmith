package selection

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Text is the plain text of the selected cells of the screen rows (one string
// per row, ANSI codes allowed, in screen coordinates). Every row loses its
// trailing spaces, and empty rows at the start and the end are dropped. A wide
// character is taken whole. It is empty when nothing is selected.
func (s State) Text(lines []string) string {
	first, last := s.bounds()
	var rows []string
	for y := first.Y; y <= last.Y && y < len(lines); y++ {
		if y < 0 {
			continue
		}
		from, to, ok := s.cols(y)
		if !ok {
			continue
		}
		from, to = snap(lines[y], from, to)
		rows = append(rows, strings.TrimRight(ansi.Strip(ansi.Cut(lines[y], from, to)), " "))
	}
	for len(rows) > 0 && rows[0] == "" {
		rows = rows[1:]
	}
	for len(rows) > 0 && rows[len(rows)-1] == "" {
		rows = rows[:len(rows)-1]
	}
	return strings.Join(rows, "\n")
}
