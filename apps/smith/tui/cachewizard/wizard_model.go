package cachewizard

import (
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// maxNotes bounds the lines the wizard keeps.
const maxNotes = 200

// Model is the screen of the remote cache setup. It draws what the engine asks
// (see Screen) and turns keys and clicks into answers (Outcome); it runs
// nothing and holds no secret: the key lives in the engine's session only.
type Model struct {
	// margins are the rows above and below the dialog that stay free.
	margins layout.Margins
	// rect is the room Place gave the dialog; the frame is drawn in box.
	rect layout.Rect
	// sw and sh are the screen's size, kept to centre the box.
	sw, sh int
	scr     Screen
	notes   []string
	cursor  int
	// over is 1 + the index of the option under the pointer, 0 for none;
	// overToggle is true while the pointer is over the commit box.
	over       int
	overToggle bool
	// overButton is the button under the pointer: 1 + its body line and its
	// first column, the zero value for none.
	overButton [2]int
	input      []rune
	// focus is the button of a Confirm screen: 0 yes, 1 no.
	focus int
	// ticked is the commit box of a Confirm screen that offers it.
	ticked bool
	// top is the first body line shown when the body is taller than the dialog.
	top   int
	since time.Time
	// typed is what the person has typed to jump in a list, and typedAt when
	// the last letter came: a pause starts a new word.
	typed   string
	typedAt time.Time
}

// New is the wizard before the engine has said anything: a spinner.
func New(now time.Time) Model { return Model{since: now} }

// Within keeps the dialog off the top rows (the header and the tab strip) and
// the bottom rows (the footer and the status bar) of the screen, as every
// dialog does: Place centres it in the rows between.
//
// Parameters: top and bottom are the rows to leave free.
// Returns: the dialog with the margins set.
func (m Model) Within(top, bottom int) Model {
	m.margins = layout.Margins{Top: max(top, 0), Bottom: max(bottom, 0)}
	return m
}

// Place sizes the dialog for a screen: 22 rows, or the rows between the
// margins when there are fewer. A body taller than that scrolls above the
// buttons, and the window follows the cursor.
func (m Model) Place(sw, sh int) Model {
	w := min(max(sw-6, 40), 100)
	m.sw, m.sh = sw, sh
	m.rect = m.margins.Centre(sw, sh, w, 22)
	return m.follow()
}

// Rect is where the dialog is drawn: as tall as the screen's content, up to
// the room Place gave it, centred between the margins as every dialog is.
func (m Model) Rect() layout.Rect {
	rows, _ := m.layout(theme.Default(), 0, time.Time{})
	h := min(max(len(rows)+2, 8), m.rect.H)
	return m.margins.Centre(m.sw, m.sh, m.rect.W, h)
}

// Kind is the kind of the screen now.
func (m Model) Kind() Kind { return m.scr.Kind }

// Handle takes news from the engine.
//
// Parameters: ev is the news; now is the clock, for the elapsed time of slow work.
// Returns: the model after the news.
func (m Model) Handle(ev Event, now time.Time) Model {
	switch ev := ev.(type) {
	case SayEvent:
		m.notes = append(m.notes, ev.Text)
		if len(m.notes) > maxNotes {
			m.notes = m.notes[len(m.notes)-maxNotes:]
		}
	case IdleEvent:
		if m.scr.Kind == Secret {
			// What the secret session printed belongs to its screen.
			m.notes = nil
		}
		m.scr, m.since = Screen{Kind: Idle}, now
	case AskEvent:
		m.scr, m.since = ev.Screen, now
		m.cursor, m.input, m.focus, m.top, m.ticked = 0, nil, 0, 0, false
		for i, o := range ev.Screen.Options {
			if o.Value == ev.Screen.Default {
				m.cursor = i
			}
		}
		m.input = []rune(ev.Screen.Typed)
		if e := ev.Screen.Confirmation.Earlier; e != nil {
			// The person came back here: the old answer is marked again.
			m.focus, m.ticked = 1, e.Commit && ev.Screen.Confirmation.OfferCommit
			if e.Yes {
				m.focus = 0
			}
		}
		if ev.Screen.Kind == Secret {
			m.notes = nil
		}
		return m.follow()
	}
	return m
}

// answered ends the screen: the next one comes from the engine.
func (m Model) answered(r Reply) (Model, Outcome) {
	m.scr, m.notes = Screen{Kind: Idle}, nil
	return m, Outcome{Reply: &r}
}

// Key handles a key.
//
// Returns: the model and what the host must do.
func (m Model) Key(msg tea.KeyMsg) (Model, Outcome) {
	m, out := m.key(msg)
	if m.scr.Kind == Page {
		return m, out
	}
	return m.follow(), out
}

// key is Key before the body follows the cursor.
func (m Model) key(msg tea.KeyMsg) (Model, Outcome) {
	name := msg.String()
	switch m.scr.Kind {
	case Choose:
		return m.chooseKey(name)
	case Text:
		return m.textKey(msg)
	case Confirm:
		return m.confirmKey(name)
	case Page:
		return m.pageKey(name)
	case Secret:
		return m, secretPress(msg)
	}
	if name == "esc" || name == "ctrl+c" {
		return m, Outcome{Cancel: true}
	}
	return m, Outcome{}
}

// secretPress turns a key into the press the secret session reads: a letter as
// it is, Esc and Ctrl+C as the session knows them, any other key as one that
// answers no to a command.
func secretPress(msg tea.KeyMsg) Outcome {
	switch {
	case msg.Type == tea.KeyEsc:
		return Outcome{Press: 27}
	case msg.Type == tea.KeyCtrlC:
		return Outcome{Press: 3}
	case msg.Type == tea.KeyEnter:
		return Outcome{Press: '\r'}
	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1:
		return Outcome{Press: msg.Runes[0]}
	}
	return Outcome{Press: 1}
}

// escape is Esc: the question before when one exists, else the wizard closes.
func (m Model) escape() (Model, Outcome) {
	if m.scr.CanBack {
		return m.back()
	}
	return m, Outcome{Cancel: true}
}

// back asks the engine for the question before this one.
func (m Model) back() (Model, Outcome) {
	if !m.scr.CanBack {
		return m, Outcome{}
	}
	return m.answered(Reply{Back: true})
}

func (m Model) chooseKey(name string) (Model, Outcome) {
	n := len(m.scr.Options)
	switch name {
	case "esc":
		return m.escape()
	case "left":
		return m.back()
	}
	if r, ok := typedLetter(name); ok {
		// A letter jumps to the first entry that starts with it, as in a file
		// browser. When no entry does, it keeps its old meaning (j, k, q...).
		if next, found := m.jumpTo(r, time.Now()); found {
			return next, Outcome{}
		}
	}
	if name == "b" {
		return m.back()
	}
	switch commands.Nav(name) {
	case commands.NavCancel:
		return m, Outcome{Cancel: true}
	case commands.NavUp:
		m.cursor = max(m.cursor-1, 0)
	case commands.NavDown:
		m.cursor = min(m.cursor+1, n-1)
	case commands.NavHome:
		m.cursor = 0
	case commands.NavEnd:
		m.cursor = n - 1
	case commands.NavPageUp:
		m.cursor = max(m.cursor-8, 0)
	case commands.NavPageDown:
		m.cursor = min(m.cursor+8, n-1)
	case commands.NavConfirm:
		if name == "enter" {
			return m.pick(m.cursor)
		}
	default:
		if len(name) == 1 && name[0] >= '1' && name[0] <= '9' && int(name[0]-'1') < n {
			return m.pick(int(name[0] - '1'))
		}
	}
	return m, Outcome{}
}

func (m Model) pick(i int) (Model, Outcome) {
	if i < 0 || i >= len(m.scr.Options) {
		return m, Outcome{}
	}
	return m.answered(Reply{Value: m.scr.Options[i].Value})
}

func (m Model) textKey(msg tea.KeyMsg) (Model, Outcome) {
	switch msg.Type {
	case tea.KeyEsc:
		return m.escape()
	case tea.KeyCtrlC:
		return m, Outcome{Cancel: true}
	case tea.KeyLeft:
		// A typed line has no cursor to move, so Left goes back once nothing is typed.
		if len(m.input) == 0 {
			return m.back()
		}
	case tea.KeyEnter:
		return m.answered(Reply{Value: strings.TrimSpace(string(m.input))})
	case tea.KeyBackspace:
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	case tea.KeyCtrlU:
		m.input = nil
	case tea.KeySpace:
		m.input = append(m.input, ' ')
	case tea.KeyRunes:
		m.input = append(m.input, msg.Runes...)
	}
	return m, Outcome{}
}

func (m Model) confirmKey(name string) (Model, Outcome) {
	switch name {
	case "esc":
		return m.escape()
	case "ctrl+c":
		return m, Outcome{Cancel: true}
	case "b", "B", "left": // words:literal
		return m.back()
	case "c", "C": // words:literal
		m.ticked = m.ticked != m.scr.Confirmation.OfferCommit
	case "y", "Y": // words:literal
		return m.answered(Reply{Yes: true, Commit: m.ticked})
	case "n", "N": // words:literal
		return m.answered(Reply{})
	case "tab", "right", "h", "l":
		m.focus = 1 - m.focus
	case "enter":
		return m.answered(Reply{Yes: m.focus == 0, Commit: m.ticked && m.focus == 0})
	}
	return m, Outcome{}
}

func (m Model) pageKey(name string) (Model, Outcome) {
	switch name {
	case "esc":
		return m.escape()
	case "b", "left":
		return m.back()
	}
	switch commands.Nav(name) {
	case commands.NavCancel:
		return m, Outcome{Cancel: true}
	case commands.NavConfirm:
		return m.answered(Reply{})
	case commands.NavUp:
		return m.scrollBy(-1), Outcome{}
	case commands.NavDown:
		return m.scrollBy(1), Outcome{}
	case commands.NavPageUp:
		return m.scrollBy(-8), Outcome{}
	case commands.NavPageDown:
		return m.scrollBy(8), Outcome{}
	}
	return m, Outcome{}
}

// Wheel scrolls the body of the screen: the wheel never moves the selection
// of a list, only the keys and a click do.
func (m Model) Wheel(delta int) Model { return m.scrollBy(delta) }

// typeAheadWindow is how long a pause ends the word being typed.
const typeAheadWindow = time.Second

// typedLetter is the letter a key name stands for, when it is one.
func typedLetter(name string) (rune, bool) {
	r := []rune(name)
	if len(r) != 1 || !unicode.IsLetter(r[0]) {
		return 0, false
	}
	return unicode.ToLower(r[0]), true
}

// jumpTo moves the cursor to the first entry whose label starts with what has
// been typed (the letter added to the word typed a moment ago); when none does,
// to the next entry that starts with the letter alone, after the cursor, so
// pressing one letter again goes round the entries that start with it.
func (m Model) jumpTo(r rune, now time.Time) (Model, bool) {
	word := string(r)
	if now.Sub(m.typedAt) <= typeAheadWindow {
		word = m.typed + word
	}
	starts := func(i int, prefix string) bool {
		return strings.HasPrefix(strings.ToLower(m.scr.Options[i].Label), prefix)
	}
	n := len(m.scr.Options)
	// The same letter again, on an entry that starts with it, goes on to the next such entry.
	repeat := len(word) == 1 && m.typed == word && m.cursor < n && starts(m.cursor, word)
	if !repeat {
		for i := range n {
			if starts(i, word) {
				m.cursor, m.typed, m.typedAt = i, word, now
				return m, true
			}
		}
	}
	letter := string(r)
	for k := 1; k <= n; k++ {
		if i := (m.cursor + k) % n; starts(i, letter) {
			m.cursor, m.typed, m.typedAt = i, letter, now
			return m, true
		}
	}
	return m, false
}

// Hover notes the option, the commit box or the button under the pointer at
// (x, y): an option and the box take the row hover look, a button the button
// hover look.
//
// Parameters: th is the theme the dialog is drawn with; x and y are screen cells.
// Returns: the dialog with the hover set.
func (m Model) Hover(th theme.Theme, x, y int) Model {
	m.over, m.overToggle, m.overButton = 0, false, [2]int{}
	r := m.Rect()
	if !r.Contains(x, y) {
		return m
	}
	row, col := y-r.Y-1, x-r.X-2
	_, hits := m.layout(th, 0, time.Time{})
	for _, h := range hits {
		if h.row == row && col >= h.x0 && col < h.x1 {
			m.over, m.overToggle = h.option+1, h.toggle
			if h.option < 0 && !h.toggle {
				m.overButton = [2]int{h.line + 1, h.x0}
			}
		}
	}
	return m
}

// Click handles a left press: a press on an option picks it, like a click in
// every single-choice list; a press on the commit box ticks or unticks it; a
// press on a button does what the button says. A press elsewhere does nothing.
//
// Parameters: th is the theme the dialog is drawn with; x and y are screen cells.
func (m Model) Click(th theme.Theme, x, y int) (Model, Outcome) {
	r := m.Rect()
	if !r.Contains(x, y) {
		return m, Outcome{}
	}
	row, col := y-r.Y-1, x-r.X-2
	_, hits := m.layout(th, 0, time.Time{})
	for _, h := range hits {
		if h.row != row || col < h.x0 || col >= h.x1 {
			continue
		}
		switch {
		case h.option >= 0:
			m.cursor = h.option
			return m.pick(h.option)
		case h.yes != nil:
			return m.answered(Reply{Yes: *h.yes, Commit: m.ticked && *h.yes})
		case h.toggle:
			m.ticked = !m.ticked
			return m, Outcome{}
		case h.back:
			return m.back()
		case h.cancel:
			return m, Outcome{Cancel: true}
		case h.page:
			return m.answered(Reply{})
		default:
			return m, Outcome{Press: h.press}
		}
	}
	return m, Outcome{}
}
