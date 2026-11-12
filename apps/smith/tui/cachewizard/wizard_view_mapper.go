package cachewizard

import (
	"fmt"
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

// Title is the frame title of the dialog.
const Title = "Set up the remote cache"

// hit is a clickable span of the body: row is the body line, x0 and x1 the
// columns (from the start of the inner text).
type hit struct {
	row, x0, x1 int
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
	if r.W < 24 || r.H < 8 {
		return nil
	}
	body, _ := m.body(th, tick, now)
	inner, height := r.W-4, r.H-2
	rows := make([]string, 0, height)
	for _, l := range body {
		if len(rows) >= height {
			break
		}
		rows = append(rows, " "+pane.Fit(l, inner))
	}
	// The box is as tall as its content, up to the room Place gave it.
	return pane.Frame(th, Title, rows, r.W, min(max(len(rows)+2, 8), r.H), true)
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

// body is the lines of the screen and where a click lands.
func (m Model) body(th theme.Theme, tick int, now time.Time) ([]string, []hit) {
	inner, height := m.rect.W-4, m.rect.H-2
	var lines []string
	var hits []hit
	add := func(l ...string) { lines = append(lines, l...) }
	spin := func(text string) string {
		return th.Accent(th.Spin(tick)) + " " + progress.Waiting(text, now.Sub(m.since))
	}
	switch m.scr.Kind {
	case Choose:
		head := append(m.noteLines(th, inner, max(height/4, 2)), "")
		head = append(head, pane.Wrap(th.Bold(m.scr.Prompt), inner, 0)...)
		add(head...)
		foot := []string{"", th.Muted(words.Say("cachewizard.hint.choose", "escape", m.escapeHint())), th.Muted(words.Say("cachewizard.hint.choose.click"))}
		hint := m.scr.Options[m.cursor].Hint
		if hint != "" {
			foot = append([]string{""}, append(pane.Wrap(th.Muted(hint), inner, 2), foot[1:]...)...)
		}
		room := max(height-len(lines)-len(foot)-2, 3)
		top := m.list.Show(m.cursor, room, len(m.scr.Options))
		for i := top; i < min(top+room, len(m.scr.Options)); i++ {
			r := choicerow.Row{Label: fmt.Sprintf("%d. %s", i+1, m.scr.Options[i].Label), Selected: i == m.cursor, Hovered: i == m.over-1 && i != m.cursor}
			hits = append(hits, hit{row: len(lines), x0: 0, x1: inner, option: i})
			add(choicerow.Line(th, r, inner))
		}
		if n := len(m.scr.Options); n > room {
			add(th.Muted(fmt.Sprintf("  %d of %d", m.cursor+1, n)))
		}
		add(foot...)
		add("")
		hits = append(hits, m.navHits(th, len(lines))...)
		add(m.navRow(th))
	case Text:
		add(m.noteLines(th, inner, max(height/3, 2))...)
		add("")
		add(pane.Wrap(th.Bold(m.scr.Prompt), inner, 0)...)
		add("", th.Accent(choicerow.Cursor(th, true))+string(m.input)+th.Accent("▌"), "", th.Muted(words.Say("cachewizard.hint.text", "escape", m.escapeHint())))
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
		if c.OfferCommit {
			line := choicerow.Toggle(th, dialog.CommitLabel+"  (C)", m.ticked, m.overToggle)
			hits = append(hits, hit{row: len(lines), x0: 0, x1: ansi.StringWidth(line), option: -1, toggle: true})
			add(line, "")
		}
		yes, no := true, false
		labels, keys := []string{c.Yes, c.No}, []string{"y", "n"}
		x := 0
		var row string
		for i, label := range labels {
			state := button.Rest
			if i == m.focus {
				state = button.Active
			}
			styled := button.Render(th, label, keys[i], state)
			w := button.Width(th, label, keys[i])
			h := hit{row: len(lines), x0: x, x1: x + w, option: -1, yes: &yes}
			if i == 1 {
				h.yes = &no
			}
			hits = append(hits, h)
			row += styled + "  "
			x += w + 2
		}
		add(row, "")
		hint := c.YesHint
		if m.focus == 1 {
			hint = c.NoHint
		}
		if hint != "" {
			add(pane.Wrap(th.Muted(hint), inner, 2)...)
		}
		hits = append(hits, m.navHits(th, len(lines))...)
		add(m.navRow(th), "")
		add(th.Muted(words.Say("cachewizard.hint.confirm", "escape", m.escapeHint())))
	case Page:
		var all []string
		for _, l := range m.scr.Lines {
			all = append(all, pane.Wrap(l, inner, 3)...)
		}
		notes := m.noteLines(th, inner, 2)
		room := max(height-4-len(notes), 3)
		top := min(m.scroll, max(len(all)-room, 0))
		add(notes...)
		add(th.Bold(m.scr.Prompt), "")
		add(all[top:min(top+room, len(all))]...)
		if len(all) > room {
			for len(lines) < height-3 {
				add("")
			}
		} else {
			add("")
		}
		more := ""
		if len(all) > room {
			more = fmt.Sprintf("up/down scroll (%d of %d) · ", min(top+room, len(all)), len(all))
		}
		next := words.Say("cachewizard.button.next")
		hits = append(hits, hit{row: len(lines), x0: 0, x1: button.Width(th, next, "Enter"), option: -1, page: true})
		hits = append(hits, m.navHits(th, len(lines), button.Width(th, next, "Enter")+2)...)
		add(button.Render(th, next, "Enter", stateOf(true)) + "  " + m.navRow(th))
		add(th.Muted(more + words.Say("cachewizard.hint.page", "escape", m.escapeHint())))
	case Secret:
		noun := m.scr.Noun
		if len(m.notes) == 0 {
			add(spin("Making the " + cachekey.InSentence(noun)))
			break
		}
		add(th.Bold(noun+" ready"), "")
		add(m.noteLines(th, inner, max(height-8, 3))...)
		add("")
		x := 0
		var row string
		for _, b := range m.scr.Buttons {
			w := button.Width(th, b.Label, keyName(b.Key))
			if x > 0 && x+w > inner {
				add(row)
				row, x = "", 0
			}
			hits = append(hits, hit{row: len(lines), x0: x, x1: x + w, option: -1, press: b.Key})
			row += button.Render(th, b.Label, keyName(b.Key), stateOf(false)) + "  "
			x += w + 2
		}
		add(row, "")
		add(pane.Wrap(th.Muted("The "+cachekey.InSentence(noun)+" is on the clipboard and is not shown here. smith keeps it in memory only while this screen is open, and clears the clipboard if it still holds it."), inner, 0)...)
	case Working:
		add(m.noteLines(th, inner, max(height-6, 2))...)
		add("", spin(m.scr.Prompt), "", th.Muted("smith waits for the program to answer. esc closes"))
	default:
		add(m.noteLines(th, inner, max(height-4, 2))...)
		add("", spin("One moment"), "", th.Muted("esc closes"))
	}
	return lines, hits
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
func (m Model) navButtons() (labels, keys []string) {
	cancelKey := "esc"
	if m.scr.CanBack {
		labels, keys = append(labels, words.Say("cachewizard.button.back")), append(keys, "b")
		cancelKey = "ctrl+c"
	}
	return append(labels, words.Say("cachewizard.button.cancel")), append(keys, cancelKey)
}

// navRow draws the Back and Cancel buttons on one line.
func (m Model) navRow(th theme.Theme) string {
	labels, keys := m.navButtons()
	var row string
	for i := range labels {
		row += button.Render(th, labels[i], keys[i], stateOf(false)) + "  "
	}
	return row
}

// navHits are the clickable spans of navRow on body line row, from column x0.
func (m Model) navHits(th theme.Theme, row int, x0 ...int) []hit {
	x := 0
	if len(x0) > 0 {
		x = x0[0]
	}
	labels, keys := m.navButtons()
	var hits []hit
	for i := range labels {
		w := button.Width(th, labels[i], keys[i])
		hits = append(hits, hit{row: row, x0: x, x1: x + w, option: -1, back: i < len(labels)-1, cancel: i == len(labels)-1})
		x += w + 2
	}
	return hits
}

// keyName is the key of a button as the person presses it.
func keyName(r rune) string {
	if r == 27 {
		return "Esc"
	}
	return string(r)
}

// stateOf is the state of a button: Active while it is the one Enter takes,
// Rest otherwise.
func stateOf(takesEnter bool) button.State {
	if takesEnter {
		return button.Active
	}
	return button.Rest
}
