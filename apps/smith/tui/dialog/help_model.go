package dialog

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// HelpRow is one line of the shortcuts overlay.
type HelpRow struct {
	// Group is the menu the command sits in.
	Group, Title, Shortcut string
}

// Help lists every shortcut, grouped by menu, with the mouse equivalents. It
// scrolls with the arrows, Page keys and the wheel, and closes with Esc, F1 or ?.
type Help struct {
	rows  []HelpRow
	extra []string
	top   int
	rect  Rect
}

// NewHelp builds the overlay from the shortcuts and some free-text lines
// (the mouse actions) shown after them.
func NewHelp(rows []HelpRow, extra []string) Help { return Help{rows: rows, extra: extra} }

func (h Help) content(th theme.Theme, w int) []string {
	var out []string
	last := ""
	for _, r := range h.rows {
		if r.Group != last {
			if last != "" {
				out = append(out, "")
			}
			out = append(out, th.Title(r.Group))
			last = r.Group
		}
		gap := w - 2 - pane.Width(r.Title) - pane.Width(r.Shortcut)
		if gap < 2 {
			// Too long for one line: the title wraps and the keys follow below it.
			out = append(out, pane.Wrap("  "+r.Title, w, 4)...)
			out = append(out, "    "+th.Bold(r.Shortcut))
			continue
		}
		out = append(out, "  "+r.Title+strings.Repeat(" ", gap)+th.Bold(r.Shortcut))
	}
	if len(h.extra) > 0 {
		out = append(out, "", th.Title("Mouse"))
		for _, l := range h.extra {
			out = append(out, pane.Wrap("  "+l, w, 4)...)
		}
	}
	return out
}

// Place centres the overlay on a screen of the given size.
func (h Help) Place(sw, sh int) Help {
	w := min(max(sw-6, 10), 92)
	hh := min(max(sh-2, 3), 28)
	h.rect = Rect{X: max((sw-w)/2, 0), Y: max((sh-hh)/2, 0), W: w, H: hh}
	return h
}

// Rect is where the overlay sits after Place.
func (h Help) Rect() Rect { return h.rect }

func (h Help) maxTop(n int) int { return max(n-(h.rect.H-2), 0) }

// Lines draws the overlay.
func (h Help) Lines(th theme.Theme) []string {
	r := h.rect
	if r.W < 6 || r.H < 3 {
		return nil
	}
	content := h.content(th, r.W-4)
	top := min(max(h.top, 0), h.maxTop(len(content)))
	rows := make([]string, 0, r.H-2)
	for i := 0; i < r.H-2; i++ {
		line := ""
		if top+i < len(content) {
			line = content[top+i]
		}
		rows = append(rows, " "+line)
	}
	return pane.Frame(th, "Keyboard shortcuts · Esc closes · ↑↓ scroll", rows, r.W, r.H, true)
}

// Key handles a key name; the overlay ends on Esc, F1, ? or q.
func (h Help) Key(name string, total int) (Help, Outcome) {
	switch name {
	case "esc", "f1", "?", "q", "enter", "ctrl+c":
		return h, Outcome{Done: true}
	case "up", "k":
		h.top--
	case "down", "j":
		h.top++
	case "pgup":
		h.top -= h.rect.H - 3
	case "pgdown":
		h.top += h.rect.H - 3
	case "home":
		h.top = 0
	case "end":
		h.top = total
	}
	h.top = min(max(h.top, 0), h.maxTop(total))
	return h, Outcome{}
}

// Total is the number of content lines, for Key.
func (h Help) Total(th theme.Theme) int { return len(h.content(th, max(h.rect.W-4, 1))) }

// Wheel scrolls by delta lines.
func (h Help) Wheel(delta, total int) Help {
	h.top = min(max(h.top+delta, 0), h.maxTop(total))
	return h
}

// Click closes the overlay when the press is outside it.
func (h Help) Click(x, y int) Outcome {
	return Outcome{Done: !h.rect.Contains(x, y)}
}
