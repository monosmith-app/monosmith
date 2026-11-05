package cachewizard

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
	// scroll is the first line of a Page.
	scroll int
	since  time.Time
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
		m.cursor, m.list, m.input, m.focus, m.scroll = 0, &pane.Scroll{}, nil, 0, 0
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
	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1:
		return Outcome{Press: msg.Runes[0]}
	}
	return Outcome{Press: 1}
}

func (m Model) chooseKey(name string) (Model, Outcome) {
	n := len(m.scr.Options)
	switch name {
	case "esc", "ctrl+c":
		return m, Outcome{Cancel: true}
	case "up", "k":
		m.cursor = max(m.cursor-1, 0)
	case "down", "j":
		m.cursor = min(m.cursor+1, n-1)
	case "home":
		m.cursor = 0
	case "end":
		m.cursor = n - 1
	case "pgup":
		m.cursor = max(m.cursor-8, 0)
	case "pgdown":
		m.cursor = min(m.cursor+8, n-1)
	case "enter":
		return m.pick(m.cursor)
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
	case "y", "Y":
		return m.answered(Reply{Yes: true})
	case "n", "N":
		return m.answered(Reply{})
	case "tab", "left", "right", "h", "l":
		m.focus = 1 - m.focus
	case "enter":
		return m.answered(Reply{Yes: m.focus == 0})
	}
	return m, Outcome{}
}

func (m Model) pageKey(name string) (Model, Outcome) {
	switch name {
	case "esc", "ctrl+c":
		return m, Outcome{Cancel: true}
	case "enter", " ":
		return m.answered(Reply{})
	case "up", "k":
		m.scroll = max(m.scroll-1, 0)
	case "down", "j":
		m.scroll++
	case "pgup":
		m.scroll = max(m.scroll-8, 0)
	case "pgdown":
		m.scroll += 8
	}
	return m, Outcome{}
}

// Wheel scrolls a page, or moves the cursor of a list.
func (m Model) Wheel(delta int) Model {
	switch m.scr.Kind {
	case Page:
		m.scroll = max(m.scroll+delta, 0)
	case Choose:
		m.cursor = min(max(m.cursor+delta, 0), len(m.scr.Options)-1)
	}
	return m
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
			return m.answered(Reply{Yes: *h.yes})
		case h.page:
			return m.answered(Reply{})
		default:
			return m, Outcome{Press: h.press}
		}
	}
	return m, Outcome{}
}
