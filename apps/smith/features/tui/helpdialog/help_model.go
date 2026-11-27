package helpdialog

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// HelpRow is one line of the shortcuts overlay.
type HelpRow struct {
	// Group is the menu the command sits in.
	Group, Title, Shortcut string
	// Disabled is why the key cannot run now ("" when it can): the row is muted
	// and the reason follows it.
	Disabled string
	// View is true for a key of the open view; the others work everywhere.
	View bool
}

// Help lists every shortcut by scope (the keys of the open view, then the keys
// that work everywhere), with the mouse actions after them. It scrolls with
// the arrows, Page keys and the wheel, and closes with Esc, F1 or ?.
type Help struct {
	rows  []HelpRow
	extra []string
	// view is the name of the open view, for the title and the heading of its keys.
	view string
	top  int
	rect layout.Rect
	// margins are the rows above and below the dialog that stay free.
	margins layout.Margins
}

// NewHelp builds the overlay from the shortcuts and some free-text lines
// (the mouse actions) shown after them.
func NewHelp(rows []HelpRow, extra []string) Help { return Help{rows: rows, extra: extra} }

// Titled names the open view: the title says "Keys · <view>" and its keys sit
// under that name.
//
// Parameters: view is the name of the open view.
// Returns: the overlay with the name.
func (h Help) Titled(view string) Help {
	h.view = view
	return h
}

// maxKeyW is the widest the key cell gets; a longer key takes its own line.
const maxKeyW = 16

// content is the part of the overlay that scrolls, in w cells: the keys by
// scope (the open view's first), then the notes and the mouse.
func (h Help) content(th theme.Theme, w int) []string {
	var everywhere, here []HelpRow
	for _, r := range h.rows {
		if r.View {
			here = append(here, r)
		} else {
			everywhere = append(everywhere, r)
		}
	}
	// One column: the titles are sentences, which two columns side by side
	// would wrap into many short lines.
	var out []string
	if len(here) > 0 {
		out = append(out, h.column(th, h.viewHeading(), here, w, false)...)
		out = append(out, "")
	}
	out = append(out, h.column(th, words.Say("helppage.scope.everywhere"), everywhere, w, true)...)
	out = append(out, "", th.Faint(strings.Repeat("─", max(w, 1))), "")
	for _, note := range []string{words.Say("helppage.note.commands"), words.Say("helppage.note.plain")} {
		out = append(out, pane.Wrap(th.Muted(note), w, 0)...)
	}
	if len(h.extra) > 0 {
		out = append(out, "", th.Title(words.Say("dialog.help.mouse")))
		for _, l := range h.extra {
			out = append(out, pane.Wrap("  "+l, w, 4)...)
		}
	}
	return out
}

// viewHeading is the heading of the keys of the open view.
func (h Help) viewHeading() string {
	if h.view == "" {
		return words.Say("helppage.scope.view.unnamed")
	}
	return words.Say("helppage.scope.view", "view", h.view)
}

// column draws the keys of one scope in w cells: its heading in the accent,
// then each key in the accent and what it does beside it, muted with its reason
// when it cannot run now. grouped puts the menu of each key over its keys.
func (h Help) column(th theme.Theme, heading string, rows []HelpRow, w int, grouped bool) []string {
	out := []string{th.Accent(th.Bold(heading))}
	kw := 0
	for _, r := range rows {
		if n := pane.Width(r.Shortcut); n <= maxKeyW {
			kw = max(kw, n)
		}
	}
	last := ""
	for _, r := range rows {
		if grouped && r.Group != last {
			out = append(out, "", th.Muted(r.Group))
			last = r.Group
		}
		key, title := th.Accent(th.Bold(r.Shortcut)), r.Title
		if r.Disabled != "" {
			key, title = th.Muted(r.Shortcut), th.Muted(r.Title)
		}
		textW := max(w-kw-4, 8)
		pad := strings.Repeat(" ", kw+4)
		lines := pane.Wrap(title, textW, 0)
		if pane.Width(r.Shortcut) > kw {
			// A key too long for the column goes on its own line, its text under it.
			out = append(out, "  "+key)
			for _, l := range lines {
				out = append(out, pad+l)
			}
		} else {
			for i, l := range lines {
				lead := pad
				if i == 0 {
					lead = "  " + pane.Fit(key, kw) + "  "
				}
				out = append(out, lead+l)
			}
		}
		if r.Disabled != "" {
			for _, l := range pane.Wrap(words.Say("dialog.help.disabled", "reason", r.Disabled), textW, 0) {
				out = append(out, pad+th.Muted(l))
			}
		}
	}
	return out
}

// foot is the line pinned under the keys: how to find a command by name, and
// how to scroll and close the overlay, in the footer's style.
func (h Help) foot(th theme.Theme, w int) []string {
	left := th.Muted(words.Say("helppage.foot.find")) + "  " + th.Accent(th.Bold(commands.KeyLabel("ctrl+k")))
	right := th.Accent(th.Bold("↑/↓")) + // words:literal
		" " + th.Muted(words.Say("helppage.foot.scroll")) + th.Border(" • ", false) + th.Accent(th.Bold(commands.KeyLabel("esc"))) + " " + th.Muted(words.Say("helppage.foot.close"))
	gap := w - pane.Width(left) - pane.Width(right)
	if gap < 2 {
		return []string{th.Faint(strings.Repeat("─", max(w, 1))), pane.Fit(right, w)}
	}
	return []string{th.Faint(strings.Repeat("─", max(w, 1))), left + strings.Repeat(" ", gap) + right}
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
// Within set: 32 rows and at most 104 columns, or the room between the margins
// when there is less.
func (h Help) Place(sw, sh int) Help {
	w := min(max(sw-6, 10), 104)
	h.rect = h.margins.Centre(sw, sh, w, 32)
	return h
}

// Rect is where the overlay sits after Place.
func (h Help) Rect() layout.Rect { return h.rect }

// footRows are the rows the pinned foot takes.
const footRows = 2

func (h Help) maxTop(n int) int { return max(n-(h.rect.H-2-footRows), 0) }

// Lines draws the overlay: a rounded frame titled "Keys · <view>", the keys
// that scroll, and the foot pinned under them.
func (h Help) Lines(th theme.Theme) []string {
	r := h.rect
	if r.W < 6 || r.H < 3 {
		return nil
	}
	content := h.content(th, r.W-4)
	top := min(max(h.top, 0), h.maxTop(len(content)))
	var foot []string
	if r.H-2 > footRows+2 {
		foot = h.foot(th, pane.DialogRowWidth(r.W)-1)
	}
	// The shared scrollbar takes the cell next to the border, and the "more"
	// rows say what is hidden, as in every dialog.
	rows := pane.DialogBody(th, content, foot, pane.DialogRowWidth(r.W), r.H-2, top)
	for i, l := range rows {
		rows[i] = " " + l
	}
	title := words.Say("helppage.title.unnamed")
	if h.view != "" {
		title = words.Say("helppage.title", "view", h.view)
	}
	return pane.Frame(th, title, rows, r.W, r.H, true)
}

// Key handles a key name; the overlay ends on Esc, F1, ? or q.
func (h Help) Key(name string, total int) (Help, Outcome) {
	switch commands.Nav(name) {
	case commands.NavUp:
		h.top--
	case commands.NavDown:
		h.top++
	case commands.NavPageUp:
		h.top -= h.rect.H - 3 - footRows
	case commands.NavPageDown:
		h.top += h.rect.H - 3 - footRows
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
