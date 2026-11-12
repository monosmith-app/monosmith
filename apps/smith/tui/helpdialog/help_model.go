package helpdialog

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// HelpRow is one line of the shortcuts overlay.
type HelpRow struct {
	// Group is the menu the command sits in.
	Group, Title, Shortcut string
	// Disabled is why the key cannot run now ("" when it can): the row is muted
	// and the reason follows it.
	Disabled string
}

// Help lists every shortcut, grouped by menu, with the mouse equivalents. It
// scrolls with the arrows, Page keys and the wheel, and closes with Esc, F1 or ?.
type Help struct {
	rows  []HelpRow
	extra []string
	top   int
	rect  layout.Rect
	// margins are the rows above and below the dialog that stay free.
	margins layout.Margins
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
		if r.Disabled != "" {
			out = append(out, th.Muted("  "+r.Title+" ("+r.Shortcut+")"))
			out = append(out, pane.Wrap(th.Muted("    "+words.Say("dialog.help.disabled", "reason", r.Disabled)), w, 4)...)
			continue
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
		out = append(out, "", th.Title(words.Say("dialog.help.mouse")))
		for _, l := range h.extra {
			out = append(out, pane.Wrap("  "+l, w, 4)...)
		}
	}
	return out
}

// Within keeps the dialog off the top rows (the header and the tab strip) and
// the bottom rows (the footer and the status bar) of the screen, as every
// dialog does: Place centres it in the rows between.
//
// Parameters: top and bottom are the rows to leave free.
// Returns: the dialog with the margins set.
func (h Help) Within(top, bottom int) Help {
	h.margins = layout.Margins{Top: max(top, 0), Bottom: max(bottom, 0)}
	return h
}

// Place centres the overlay on a screen of the given size, between the margins
// Within set: 28 rows, or the rows between the margins when there are fewer.
func (h Help) Place(sw, sh int) Help {
	w := min(max(sw-6, 10), 92)
	h.rect = h.margins.Centre(sw, sh, w, 28)
	return h
}

// Rect is where the overlay sits after Place.
func (h Help) Rect() layout.Rect { return h.rect }

func (h Help) maxTop(n int) int { return max(n-(h.rect.H-2), 0) }

// Lines draws the overlay.
func (h Help) Lines(th theme.Theme) []string {
	r := h.rect
	if r.W < 6 || r.H < 3 {
		return nil
	}
	content := h.content(th, r.W-4)
	top := min(max(h.top, 0), h.maxTop(len(content)))
	// The shared scrollbar takes the cell next to the border, and the "more"
	// rows say what is hidden, as in every dialog.
	rows := pane.DialogBody(th, content, nil, r.W-3, r.H-2, top)
	for i, l := range rows {
		rows[i] = " " + l
	}
	return pane.Frame(th, words.Say("dialog.help.title"), rows, r.W, r.H, true)
}

// Key handles a key name; the overlay ends on Esc, F1, ? or q.
func (h Help) Key(name string, total int) (Help, Outcome) {
	switch commands.Nav(name) {
	case commands.NavUp:
		h.top--
	case commands.NavDown:
		h.top++
	case commands.NavPageUp:
		h.top -= h.rect.H - 3
	case commands.NavPageDown:
		h.top += h.rect.H - 3
	case commands.NavHome:
		h.top = 0
	case commands.NavEnd:
		h.top = total
	default:
		switch name {
		case "esc", "f1", "?", "q", "enter", "ctrl+c":
			return h, Outcome{Done: true}
		}
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
