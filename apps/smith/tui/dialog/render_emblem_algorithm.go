package dialog

import "strconv"

// rgb is the colour of one pixel of the emblem.
type rgb struct{ r, g, b uint8 }

// hex is the colour as #rrggbb.
func (c rgb) hex() string {
	const digits = "0123456789abcdef"
	return string([]byte{'#',
		digits[c.r>>4], digits[c.r&15], digits[c.g>>4], digits[c.g&15], digits[c.b>>4], digits[c.b&15]})
}

// emblemGrid reads emblemRows into pixels: nil where the emblem has none.
func emblemGrid() [][]*rgb {
	grid := make([][]*rgb, len(emblemRows))
	for y, row := range emblemRows {
		grid[y] = make([]*rgb, 0, len(row))
		for _, ch := range row {
			grid[y] = append(grid[y], parseEmblemColour(emblemPalette[ch]))
		}
	}
	return grid
}

func parseEmblemColour(hex string) *rgb {
	if len(hex) != 7 {
		return nil
	}
	v, err := strconv.ParseUint(hex[1:], 16, 24)
	if err != nil {
		return nil
	}
	return &rgb{uint8(v >> 16), uint8(v >> 8), uint8(v)}
}

// shrink halves a grid: each 2 by 2 block becomes one pixel of the average
// colour of its pixels, and is left out when fewer than two of them exist.
func shrink(grid [][]*rgb) [][]*rgb {
	var out [][]*rgb
	for y := 0; y+1 < len(grid); y += 2 {
		row := make([]*rgb, 0, len(grid[y])/2)
		for x := 0; x+1 < len(grid[y]); x += 2 {
			row = append(row, average(grid[y][x], grid[y][x+1], grid[y+1][x], grid[y+1][x+1]))
		}
		out = append(out, row)
	}
	return out
}

func average(px ...*rgb) *rgb {
	var r, g, b, n int
	for _, p := range px {
		if p != nil {
			r, g, b, n = r+int(p.r), g+int(p.g), b+int(p.b), n+1
		}
	}
	if n < 2 {
		return nil
	}
	return &rgb{uint8(r / n), uint8(g / n), uint8(b / n)}
}

// halfBlocks draws a grid two pixel rows to a text row, with the half blocks
// ▀ and ▄. paint renders a glyph in a foreground and, when there is one, a
// background colour; a pixel pair needs both, and a text row of blanks stays
// plain spaces.
func halfBlocks(grid [][]*rgb, paint func(glyph string, fg, bg *rgb) string) []string {
	var lines []string
	for y := 0; y < len(grid); y += 2 {
		var line string
		for x := range grid[y] {
			top, bottom := grid[y][x], (*rgb)(nil)
			if y+1 < len(grid) && x < len(grid[y+1]) {
				bottom = grid[y+1][x]
			}
			switch {
			case top == nil && bottom == nil:
				line += " "
			case bottom == nil:
				line += paint("▀", top, nil)
			case top == nil:
				line += paint("▄", bottom, nil)
			default:
				line += paint("▀", top, bottom)
			}
		}
		lines = append(lines, line)
	}
	return lines
}

// emblemSize is how big the emblem is drawn.
type emblemSize int

const (
	// emblemAnvil is the small anvil, for a short terminal.
	emblemAnvil emblemSize = iota
	// emblemSmall is the emblem at 16 columns by 8 rows.
	emblemSmall
	// emblemLarge is the emblem at 32 columns by 16 rows.
	emblemLarge
)

// emblemSizeFor picks the largest emblem that leaves the text of the About
// screen its room in a terminal of the given height.
func emblemSizeFor(height int) emblemSize {
	switch {
	case height >= 44:
		return emblemLarge
	case height >= 34:
		return emblemSmall
	default:
		return emblemAnvil
	}
}
