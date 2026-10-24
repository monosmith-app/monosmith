package detail

import "github.com/monosmith-app/monosmith/apps/smith/tui/theme"

// ClampTop keeps the first visible line inside the panel: from 0 to the
// line that puts the last one on the bottom row.
func ClampTop(top, total, h int) int {
	return min(max(top, 0), max(total-h, 0))
}

// PageStep is how many lines a page key scrolls in h rows: the two rows the
// "more" markers take are not counted, so no line is skipped between pages.
func PageStep(h int) int { return max(h-3, 1) }

// Window shows h rows of lines from top on. When lines are hidden above or
// below, the first or the last row says so ("↑ more", "↓ more"), so the
// panel is never silently cut. hits are re-keyed to the visible rows; a hit
// under a marker is dropped. A panel that fits is returned as it is.
func Window(th theme.Theme, lines []string, hits map[int]string, top, h int) ([]string, map[int]string) {
	if h <= 0 || len(lines) <= h {
		return lines, hits
	}
	top = ClampTop(top, len(lines), h)
	out := append([]string(nil), lines[top:top+h]...)
	above, below := top > 0, top+h < len(lines)
	if above {
		out[0] = th.Muted(" ↑ more (PgUp, Ctrl+U)")
	}
	if below {
		out[h-1] = th.Muted(" ↓ more (PgDn, Ctrl+D)")
	}
	vis := map[int]string{}
	for row, id := range hits {
		y := row - top
		if y < 0 || y >= h || (y == 0 && above) || (y == h-1 && below) {
			continue
		}
		vis[y] = id
	}
	return out, vis
}
