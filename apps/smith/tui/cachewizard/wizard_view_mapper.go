package cachewizard

import (
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/cachekey"
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/choicerow"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// hit is a clickable span of the body: row is the row it is drawn on inside
// the frame (the body line until layout places it), line the body line, x0 and
// x1 the columns (from the start of the inner text).
type hit struct {
	row, line, x0, x1 int
	// option is the index of an option, or -1.
	option int
	// yes is set for a button of a Confirm screen.
	yes *bool
	// toggle is true for the commit box.
	toggle bool
	// page is true for the button that goes on from a Page.
	page bool
	// back and cancel are the Back and Cancel buttons.
	back, cancel bool
	// press is the key a secret button sends.
	press rune
}

// Lines draws the dialog, framed, at the size Place set.
//
// Parameters: th is the theme; tick moves the spinner; now is the clock.
func (m Model) Lines(th theme.Theme, tick int, now time.Time) []string {
	r := m.rect
	if r.W < 24 || r.H < 4 {
		return nil
	}
	rows, _ := m.layout(th, tick, now)
	for i, l := range rows {
		rows[i] = " " + pane.Fit(l, r.W-4)
	}
	// The box is as tall as its content, up to the room Place gave it.
	return pane.Frame(th, words.Say("cachewizard.title"), rows, r.W, min(max(len(rows)+2, 8), r.H), true)
}

// screen is the body of a screen as body builds it: lines[foot:] are the
// buttons that stay under the part that scrolls, and focus is the lines that
// must stay in view (the option under the cursor, the answer buttons, the
// typed line), from first to last excluded, or {-1, -1}.
type screen struct {
	lines []string
	hits  []hit
	foot  int
	focus [2]int
}

// layout is the rows drawn inside the frame and the hits keyed by those rows.
// A body that fits is drawn as it is; a longer one scrolls from m.top under
// the shared scrollbar (pane.DialogBody), and its buttons stay at the bottom.
func (m Model) layout(th theme.Theme, tick int, now time.Time) ([]string, []hit) {
	sc := m.body(th, tick, now)
	inner, height := m.rect.W-4, m.rect.H-2
	if len(sc.lines) <= height {
		for i := range sc.hits {
			sc.hits[i].line = sc.hits[i].row
		}
		return sc.lines, sc.hits
	}
	body, foot := sc.lines[:sc.foot], sc.lines[sc.foot:]
	room := pane.BodyRoom(height, len(foot))
	rows := pane.DialogBody(th, body, foot, inner, height, m.top)
	var hits []hit
	for _, h := range sc.hits {
		h.line = h.row
		if h.row >= sc.foot {
			h.row = room + h.row - sc.foot
		} else {
			y := h.row - pane.ClampTop(m.top, len(body), room)
			if pane.BodyLine(y, m.top, len(body), room) != h.line {
				continue
			}
			h.row = y
		}
		hits = append(hits, h)
	}
	return rows, hits
}

// follow scrolls the body so the lines in focus show.
func (m Model) follow() Model {
	sc := m.body(theme.Default(), 0, time.Time{})
	if sc.focus[0] < 0 || len(sc.lines) <= m.rect.H-2 {
		return m
	}
	room := pane.BodyRoom(m.rect.H-2, len(sc.lines)-sc.foot)
	m.top = pane.FollowLines(m.top, sc.focus[0], sc.focus[1], sc.foot, room)
	return m
}

// scrollBy moves the body by delta lines, never past its ends.
func (m Model) scrollBy(delta int) Model {
	sc := m.body(theme.Default(), 0, time.Time{})
	room := pane.BodyRoom(m.rect.H-2, len(sc.lines)-sc.foot)
	m.top = pane.ClampTop(m.top+delta, sc.foot, room)
	return m
}

// noteLines are the last lines the engine said, wrapped, at most max.
func (m Model) noteLines(th theme.Theme, inner, max int) []string {
	var lines []string
	for _, n := range m.notes {
		lines = append(lines, pane.Wrap(th.Muted(n), inner, 2)...)
	}
	if len(lines) > max {
		lines = lines[len(lines)-max:]
	}
	return lines
}

// body is the lines of the screen, where a click lands, the buttons that stay
// at the bottom and the lines that must stay in view.
func (m Model) body(th theme.Theme, tick int, now time.Time) screen {
	inner, height := m.rect.W-4, m.rect.H-2
	var lines []string
	var hits []hit
	foot, focus := -1, [2]int{-1, -1}
	add := func(l ...string) { lines = append(lines, l...) }
	pin := func() { foot = len(lines) }
	spin := func(text string) string {
		return th.Accent(th.Spin(tick)) + " " + progress.Waiting(text, now.Sub(m.since))
	}
	switch m.scr.Kind {
	case Choose:
		add(m.noteLines(th, inner, max(height/4, 2))...)
		add("")
		add(pane.Wrap(th.Bold(m.scr.Prompt), inner, 0)...)
		for i, o := range m.scr.Options {
			r := choicerow.Row{Label: words.Say("cachewizard.option", "number", i+1, "label", o.Label), Selected: i == m.cursor, Hovered: i == m.over-1 && i != m.cursor}
			if i == m.cursor {
				focus = [2]int{len(lines), len(lines) + 1}
			}
			hits = append(hits, hit{row: len(lines), x0: 0, x1: inner, option: i})
			add(choicerow.Line(th, r, inner))
		}
		if m.cursor < len(m.scr.Options) && m.scr.Options[m.cursor].Hint != "" {
			add("")
			add(pane.Wrap(th.Muted(m.scr.Options[m.cursor].Hint), inner, 2)...)
		}
		add("", th.Muted(words.Say("cachewizard.hint.choose", "escape", m.escapeHint())), "")
		pin()
		add(m.buttonRow(th, len(lines), m.navButtons(), &hits))
	case Text:
		add(m.noteLines(th, inner, max(height/3, 2))...)
		add("")
		add(pane.Wrap(th.Bold(m.scr.Prompt), inner, 0)...)
		add("")
		focus = [2]int{len(lines), len(lines) + 1}
		add(th.Accent(choicerow.Cursor(th, true))+string(m.input)+th.Accent("▌"), "", th.Muted(words.Say("cachewizard.hint.text", "escape", m.escapeHint())))
	case Confirm:
		c := m.scr.Confirmation
		add(m.noteLines(th, inner, max(height/5, 1))...)
		if c.Text != "" {
			add("")
			add(pane.Wrap(c.Text, inner, 0)...)
		}
		if c.Shown != "" {
			add("")
			for _, l := range strings.Split(c.Shown, "\n") {
				for _, w := range pane.Wrap(l, inner-2, 4) {
					add("  " + th.Bold(w))
				}
			}
		}
		add("", th.Bold(strings.TrimSpace(c.Prompt)), "")
		focus[0] = len(lines)
		if c.OfferCommit {
			line := choicerow.Toggle(th, words.Say("cachewizard.commit", "label", dialog.CommitLabel), m.ticked, m.overToggle)
			hits = append(hits, hit{row: len(lines), x0: 0, x1: ansi.StringWidth(line), option: -1, toggle: true})
			add(line, "")
		}
		yes, no := true, false
		answers := []cell{
			{label: c.Yes, key: "y", hit: hit{option: -1, yes: &yes}, active: m.focus == 0},
			{label: c.No, key: "n", hit: hit{option: -1, yes: &no}, active: m.focus == 1},
		}
		add(m.buttonRow(th, len(lines), answers, &hits))
		focus[1] = len(lines)
		add("")
		hint := c.YesHint
		if m.focus == 1 {
			hint = c.NoHint
		}
		if hint != "" {
			add(pane.Wrap(th.Muted(hint), inner, 2)...)
		}
		add(th.Muted(words.Say("cachewizard.hint.confirm", "escape", m.escapeHint())), "")
		pin()
		add(m.buttonRow(th, len(lines), m.navButtons(), &hits))
	case Page:
		add(m.noteLines(th, inner, 2)...)
		add(th.Bold(m.scr.Prompt), "")
		for _, l := range m.scr.Lines {
			add(pane.Wrap(l, inner, 3)...)
		}
		add("")
		hint := words.Say("cachewizard.hint.page", "escape", m.escapeHint())
		if room := pane.BodyRoom(height, 2); len(lines) > room {
			top := pane.ClampTop(m.top, len(lines), room)
			hint = words.Say("cachewizard.hint.scroll", "shown", min(top+room, len(lines)), "total", len(lines), "escape", m.escapeHint())
		}
		pin()
		next := cell{label: words.Say("cachewizard.button.next"), key: "enter", hit: hit{option: -1, page: true}, active: true}
		add(m.buttonRow(th, len(lines), append([]cell{next}, m.navButtons()...), &hits))
		add(th.Muted(hint))
	case Secret:
		noun := m.scr.Noun
		if len(m.notes) == 0 {
			add(spin(words.Say("cachewizard.secret.making", "noun", cachekey.InSentence(noun))))
			break
		}
		add(th.Bold(words.Say("cachewizard.secret.ready", "noun", noun)), "")
		add(m.noteLines(th, inner, max(height-8, 3))...)
		add("")
		pin()
		// The buttons go on as many centred rows as they need.
		var row []cell
		width := 0
		for _, b := range m.scr.Buttons {
			c := cell{label: b.Label, key: keyName(b.Key), hit: hit{option: -1, press: b.Key}}
			w := button.Width(th, c.label, c.key)
			if len(row) > 0 && width+w > inner {
				add(m.buttonRow(th, len(lines), row, &hits))
				row, width = nil, 0
			}
			row, width = append(row, c), width+w+1
		}
		add(m.buttonRow(th, len(lines), row, &hits), "")
		add(pane.Wrap(th.Muted(words.Say("cachewizard.secret.clipboard", "noun", cachekey.InSentence(noun))), inner, 0)...)
	case Working:
		add(m.noteLines(th, inner, max(height-6, 2))...)
		add("", spin(m.scr.Prompt), "", th.Muted(words.Say("cachewizard.hint.working")))
	default:
		add(m.noteLines(th, inner, max(height-4, 2))...)
		add("", spin(words.Say("cachewizard.idle")), "", th.Muted(words.Say("cachewizard.escape.close")))
	}
	if foot < 0 {
		foot = len(lines)
	}
	return screen{lines: lines, hits: hits, foot: foot, focus: focus}
}

// cell is one button of a row: what it shows, what a click on it does, and
// whether it is the one Enter takes (drawn active).
type cell struct {
	label, key string
	hit        hit
	active     bool
}

// buttonRow draws a row of buttons centred in the dialog, as every other dialog
// places them, and adds their click spans for body line row to hits. The button
// under the pointer has the hover look; the one Enter takes, the active look.
func (m Model) buttonRow(th theme.Theme, row int, cells []cell, hits *[]hit) string {
	if len(cells) == 0 {
		return ""
	}
	labels, keys := make([]string, len(cells)), make([]string, len(cells))
	for i, c := range cells {
		labels[i], keys[i] = c.label, c.key
	}
	// RowSpans counts from the dialog's edge; the body starts two cells in.
	spans := button.RowSpans(th, labels, keys, m.rect.W-4)
	var line string
	col := 0
	for i, c := range cells {
		from, to := spans[i][0]-2, spans[i][1]-2
		h := c.hit
		h.row, h.x0, h.x1 = row, from, to
		*hits = append(*hits, h)
		state := button.Rest
		switch {
		case m.overButton == [2]int{row + 1, from}:
			state = button.Hover
		case c.active:
			state = button.Active
		}
		line += strings.Repeat(" ", max(from-col, 0)) + button.Render(th, c.label, c.key, state)
		col = to
	}
	return line
}

// escapeHint says what Esc does on this screen: back, or close on the first question.
func (m Model) escapeHint() string {
	if m.scr.CanBack {
		return words.Say("cachewizard.escape.back")
	}
	return words.Say("cachewizard.escape.close")
}

// navButtons are the Back and Cancel buttons of a screen: Back only when a
// question before this one exists. Cancel closes the wizard; its key is Esc on
// the first question and Ctrl+C after, where Esc goes back.
func (m Model) navButtons() []cell {
	cancelKey := "esc"
	var cells []cell
	if m.scr.CanBack {
		cells = append(cells, cell{label: words.Say("cachewizard.button.back"), key: "b", hit: hit{option: -1, back: true}})
		cancelKey = "ctrl+c"
	}
	return append(cells, cell{label: words.Say("cachewizard.button.cancel"), key: cancelKey, hit: hit{option: -1, cancel: true}})
}

// keyName is the key of a button as the person presses it.
func keyName(r rune) string {
	if r == 27 {
		return "esc"
	}
	return string(r)
}
