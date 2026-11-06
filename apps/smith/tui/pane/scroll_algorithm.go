package pane

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// Clamp keeps v between lo and hi. When hi is below lo, lo wins, so a range
// that has no room yet never gives a value under its floor.
func Clamp(v, lo, hi int) int { return min(max(v, lo), max(lo, hi)) }

// ClampTop keeps the first visible line inside the panel: from 0 to the
// line that puts the last one on the bottom row.
func ClampTop(top, total, h int) int {
	return Clamp(top, 0, total-h)
}

// PageStep is how many lines a page key scrolls in h rows: the two rows the
// "more" markers take are not counted, so no line is skipped between pages.
func PageStep(h int) int { return max(h-3, 1) }

// WindowMarked shows h rows of lines from top on. When lines are hidden above or
// below, the first or the last row says so ("↑ more", "↓ more"), so the
// panel is never silently cut. hits are re-keyed to the visible rows; a hit
// under a marker is dropped. A panel that fits is returned as it is.
func WindowMarked(th theme.Theme, lines []string, hits map[int]string, top, h int) ([]string, map[int]string) {
	if h <= 0 || len(lines) <= h {
		return lines, hits
	}
	top = ClampTop(top, len(lines), h)
	out := append([]string(nil), lines[top:top+h]...)
	above, below := top > 0, top+h < len(lines)
	if above {
		out[0] = th.Muted(words.Say("detail.scroll.up"))
	}
	if below {
		out[h-1] = th.Muted(words.Say("detail.scroll.down"))
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
