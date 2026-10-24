package tabs

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// cells is one row of the strip being built. The pane's top edge is drawn
// glyph by glyph, with a flag for each saying whether it is in the accent; the
// other rows are made of styled pieces of text placed at a column.
type cells struct {
	ch     []rune
	accent []bool
	blocks []block
}

type block struct {
	x, w int
	s    string
}

func newCells(w int) *cells {
	return &cells{ch: make([]rune, max(w, 1)), accent: make([]bool, max(w, 1))}
}

// text places s, which is w cells wide, at column x.
func (c *cells) text(x int, s string, w int) { c.blocks = append(c.blocks, block{x: x, w: w, s: s}) }

// set puts one glyph at column x.
func (c *cells) set(x int, r rune, accent bool) {
	if x >= 0 && x < len(c.ch) {
		c.ch[x], c.accent[x] = r, accent
	}
}

// render joins the row into a line of width cells. glyphs says the row is made
// of glyphs (the pane's edge) and not of placed text.
func (c *cells) render(th theme.Theme, width int, glyphs bool) string {
	if glyphs {
		var b strings.Builder
		for x := 0; x < width; {
			y := x
			var run strings.Builder
			for y < width && c.accent[y] == c.accent[x] {
				run.WriteRune(c.ch[y])
				y++
			}
			b.WriteString(th.Border(run.String(), c.accent[x]))
			x = y
		}
		return b.String()
	}
	var b strings.Builder
	x := 0
	for _, blk := range c.blocks {
		if blk.x < x || blk.x >= width {
			continue
		}
		b.WriteString(strings.Repeat(" ", blk.x-x))
		b.WriteString(blk.s)
		x = blk.x + blk.w
	}
	if x < width {
		b.WriteString(strings.Repeat(" ", width-x))
	}
	return b.String()
}
