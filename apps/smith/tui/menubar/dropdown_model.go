package menubar

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// padX is the space at each side of an entry, inside the box.
const padX = 1

// dropdown is a box of entries drawn with its top-left corner at (X, Y).
type dropdown struct {
	entries []Entry
	cursor  int
	X, Y    int
	// minWidth keeps the box at least this wide (inner cells).
	minWidth int
}

func (d dropdown) innerWidth() int {
	w := d.minWidth
	for _, e := range d.entries {
		n := pane.Width(e.Name)
		if e.Shortcut != "" {
			n += 3 + pane.Width(e.Shortcut)
		}
		w = max(w, n)
	}
	return w + 2*padX
}

// height is the rows the box takes: entries, the two edges and the hint row.
func (d dropdown) height() int { return len(d.entries) + 3 }

// entryAt is the index of the entry under the cell (x, y), or -1.
func (d dropdown) entryAt(x, y int) int {
	j := y - d.Y - 1
	if x <= d.X || x >= d.X+d.innerWidth()+1 || j < 0 || j >= len(d.entries) || d.entries[j].Separator {
		return -1
	}
	return j
}

// contains reports whether the cell is inside the box, hint row included.
func (d dropdown) contains(x, y int) bool {
	return x >= d.X && x < d.X+d.innerWidth()+2 && y >= d.Y && y < d.Y+d.height()
}

// step is the next entry from index from in direction dir, skipping
// separators and wrapping; disabled entries can be stopped on so that their
// reason can be read.
func (d dropdown) step(from, dir int) int {
	n := len(d.entries)
	for k := 1; k <= n; k++ {
		i := ((from+dir*k)%n + n) % n
		if !d.entries[i].Separator {
			return i
		}
	}
	return -1
}

// view draws the box.
func (d dropdown) view(th theme.Theme) []string {
	w := d.innerWidth()
	rule := func(s string) string { return th.Border(s, true) }
	rows := []string{rule("╭" + strings.Repeat("─", w) + "╮")}
	for j, e := range d.entries {
		if e.Separator {
			rows = append(rows, rule("├"+strings.Repeat("─", w)+"┤"))
			continue
		}
		gap := max(w-2*padX-pane.Width(e.Name)-pane.Width(e.Shortcut), 0)
		pad := strings.Repeat(" ", padX)
		text := pad + e.Name + strings.Repeat(" ", gap) + e.Shortcut + pad
		switch {
		case j == d.cursor && e.Disabled != "":
			text = th.Hover(th.Muted(text))
		case j == d.cursor:
			text = th.Selected(text)
		case e.Disabled != "":
			text = th.Muted(text)
		default:
			text = pad + e.Name + strings.Repeat(" ", gap) + th.Muted(e.Shortcut) + pad
		}
		rows = append(rows, rule("│")+text+rule("│"))
	}
	hint := ""
	if d.cursor >= 0 && d.cursor < len(d.entries) {
		e := d.entries[d.cursor]
		hint = e.Hint
		if e.Disabled != "" {
			hint = "unavailable: " + e.Disabled
		}
	}
	rows = append(rows, rule("│")+th.Muted(pane.Fit(" "+hint, w))+rule("│"))
	rows = append(rows, rule("╰"+strings.Repeat("─", w)+"╯"))
	return rows
}
