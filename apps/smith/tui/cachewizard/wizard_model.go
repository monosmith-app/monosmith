package cachewizard

import (
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// maxNotes bounds the lines the wizard keeps.
const maxNotes = 200

// Model is the screen of the remote cache setup. It draws what the engine asks
// (see Screen) and turns keys and clicks into answers (Outcome); it runs
// nothing and holds no secret: the key lives in the engine's session only.
type Model struct {
	rect   layout.Rect
	scr    Screen
	notes  []string
	cursor int
	// list keeps where the window of a long list starts.
	list  *pane.Scroll
	input []rune
	// focus is the button of a Confirm screen: 0 yes, 1 no.
	focus int
	// ticked is the commit box of a Confirm screen that offers it.
	ticked bool
	// scroll is the first line of a Page.
	scroll int
	since  time.Time
	// typed is what the person has typed to jump in a list, and typedAt when
	// the last letter came: a pause starts a new word.
	typed   string
	typedAt time.Time
}

// New is the wizard before the engine has said anything: a spinner.
func New(now time.Time) Model { return Model{since: now, list: &pane.Scroll{}} }

// Place sizes the dialog for a screen.
func (m Model) Place(sw, sh int) Model {
	w := min(max(sw-6, 40), 100)
	h := min(max(sh-2, 12), 22)
	m.rect = layout.Rect{X: max((sw-w)/2, 0), Y: max((sh-h)/2, 0), W: w, H: h}
	return m
}

// Rect is where the dialog sits after Place.
func (m Model) Rect() layout.Rect { return m.rect }

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
		m.cursor, m.list, m.input, m.focus, m.scroll, m.ticked = 0, &pane.Scroll{}, nil, 0, 0, false
		for i, o := range ev.Screen.Options {
			if o.Value == ev.Screen.Default {
				m.cursor = i
			}
		}
		if ev.Screen.Kind == Secret {
			m.notes = nil
		}
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

func (m Model) chooseKey(name string) (Model, Outcome) {
	n := len(m.scr.Options)
	if r, ok := typedLetter(name); ok {
		// A letter jumps to the first entry that starts with it, as in a file
		// browser. When no entry does, it keeps its old meaning (j, k, q...).
		if next, found := m.jumpTo(r, time.Now()); found {
			return next, Outcome{}
		}
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
	case tea.KeyEsc, tea.KeyCtrlC:
		return m, Outcome{Cancel: true}
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
	case "esc", "ctrl+c":
		return m, Outcome{Cancel: true}
	case "c", "C":
		m.ticked = m.ticked != m.scr.Confirmation.OfferCommit
	case "y", "Y":
		return m.answered(Reply{Yes: true, Commit: m.ticked})
	case "n", "N":
		return m.answered(Reply{})
	case "tab", "left", "right", "h", "l":
		m.focus = 1 - m.focus
	case "enter":
		return m.answered(Reply{Yes: m.focus == 0, Commit: m.ticked && m.focus == 0})
	}
	return m, Outcome{}
}

func (m Model) pageKey(name string) (Model, Outcome) {
	switch commands.Nav(name) {
	case commands.NavCancel:
		return m, Outcome{Cancel: true}
	case commands.NavConfirm:
		return m.answered(Reply{})
	case commands.NavUp:
		m.scroll = max(m.scroll-1, 0)
	case commands.NavDown:
		m.scroll++
	case commands.NavPageUp:
		m.scroll = max(m.scroll-8, 0)
	case commands.NavPageDown:
		m.scroll += 8
	}
	return m, Outcome{}
}

// Wheel scrolls a page, or the window of a list: the wheel never moves the
// selection of a list, only the keys and a click do.
func (m Model) Wheel(delta int) Model {
	switch m.scr.Kind {
	case Page:
		m.scroll = max(m.scroll+delta, 0)
	case Choose:
		m.list.By(delta)
	}
	return m
}

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

// Click handles a left press: a press on an option moves the cursor to it and
// a second press (double is true) answers; a press on a button does what the
// button says. A press elsewhere does nothing.
//
// Parameters: th is the theme the dialog is drawn with; x and y are screen cells.
func (m Model) Click(th theme.Theme, x, y int, double bool) (Model, Outcome) {
	if !m.rect.Contains(x, y) {
		return m, Outcome{}
	}
	row, col := y-m.rect.Y-1, x-m.rect.X-2
	_, hits := m.body(th, 0, time.Time{})
	for _, h := range hits {
		if h.row != row || col < h.x0 || col >= h.x1 {
			continue
		}
		switch {
		case h.option >= 0:
			m.cursor = h.option
			if double {
				return m.pick(h.option)
			}
			return m, Outcome{}
		case h.yes != nil:
			return m.answered(Reply{Yes: *h.yes, Commit: m.ticked && *h.yes})
		case h.toggle:
			m.ticked = !m.ticked
			return m, Outcome{}
		case h.page:
			return m.answered(Reply{})
		default:
			return m, Outcome{Press: h.press}
		}
	}
	return m, Outcome{}
}
