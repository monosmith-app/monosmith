package palette

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Entry is a shell command (an id) listed with its shortcut, under a section.
// Hint is the command line it runs. Unavailable is the reason it cannot be
// used now, or "".
type Entry struct {
	ID, Section, Title, Hint, Shortcut, Unavailable string
}

// Result is what a key or a click did.
type Result struct {
	// CommandID is the shell command chosen.
	CommandID string
	// Refused is the reason a shell command could not be chosen.
	Refused string
	// Closed is true when the palette should close.
	Closed bool
}

func (e Entry) unavailable() bool { return e.Unavailable != "" }

// row is one line of the list: a section heading (index -1) or a match.
type row struct {
	heading string
	index   int
}

// Palette finds any command by typing part of its title, its command line or
// its section. With nothing typed the commands are listed by section.
type Palette struct {
	items  []Entry
	filter textinput.Model
	cursor int
	top    int
}

// New builds a palette of the shell's commands.
func New(entries []Entry) Palette {
	in := textinput.New()
	in.Placeholder = "type to find a command"
	in.Focus()
	return Palette{filter: in}.WithEntries(entries)
}

// WithEntries replaces the commands (their availability and scope change with
// the state) and keeps what was typed.
func (p Palette) WithEntries(entries []Entry) Palette {
	p.items = append([]Entry(nil), entries...)
	p.cursor = min(p.cursor, max(len(p.matches())-1, 0))
	return p
}

// matches are the entries the filter finds: titles starting with it first,
// then titles containing it, then hints and sections mentioning it. Nothing
// typed lists the usable ones in section order.
func (p Palette) matches() []Entry {
	q := strings.ToLower(strings.TrimSpace(p.filter.Value()))
	if q == "" {
		var usable []Entry
		for _, e := range p.items {
			if !e.unavailable() {
				usable = append(usable, e)
			}
		}
		return usable
	}
	var prefix, name, described, blocked []Entry
	for _, e := range p.items {
		n := strings.ToLower(e.Title)
		var bucket *[]Entry
		switch {
		case strings.HasPrefix(n, q):
			bucket = &prefix
		case strings.Contains(n, q):
			bucket = &name
		case strings.Contains(strings.ToLower(e.Hint), q) || strings.Contains(strings.ToLower(e.Section), q):
			bucket = &described
		default:
			continue
		}
		if e.unavailable() {
			bucket = &blocked
		}
		*bucket = append(*bucket, e)
	}
	return append(append(append(prefix, name...), described...), blocked...)
}

// rows are the lines of the list: the matches, under their section headings
// when nothing is typed.
func (p Palette) rows() []row {
	found := p.matches()
	grouped := strings.TrimSpace(p.filter.Value()) == ""
	var out []row
	last := ""
	for i, e := range found {
		if grouped && e.Section != last && e.Section != "" {
			out = append(out, row{heading: e.Section, index: -1})
			last = e.Section
		}
		out = append(out, row{index: i})
	}
	return out
}

// Entries are the commands the palette was given.
func (p Palette) Entries() []Entry { return append([]Entry(nil), p.items...) }

// hidden counts the commands left out of the list because they cannot be used now.
func (p Palette) hidden() int {
	n := 0
	for _, e := range p.items {
		if e.unavailable() {
			n++
		}
	}
	return n
}

// Reset clears the filter, for reopening.
func (p Palette) Reset() Palette {
	p.filter.SetValue("")
	p.filter.Focus()
	p.cursor, p.top = 0, 0
	return p
}

// Key handles a key press.
func (p Palette) Key(msg tea.KeyMsg) (Palette, Result) {
	switch msg.String() {
	case "esc", "ctrl+c":
		return p, Result{Closed: true}
	case "up":
		p.cursor = max(p.cursor-1, 0)
	case "down":
		p.cursor = min(p.cursor+1, max(len(p.matches())-1, 0))
	case "pgup":
		p.cursor = max(p.cursor-8, 0)
	case "pgdown":
		p.cursor = min(p.cursor+8, max(len(p.matches())-1, 0))
	case "enter":
		return p.choose(p.cursor)
	default:
		p.filter, _ = p.filter.Update(msg)
		p.cursor = min(p.cursor, max(len(p.matches())-1, 0))
	}
	return p, Result{}
}

// Choose picks the match at index i (a click).
func (p Palette) Choose(i int) (Palette, Result) { return p.choose(i) }

func (p Palette) choose(i int) (Palette, Result) {
	found := p.matches()
	if len(found) == 0 || i < 0 {
		return p, Result{}
	}
	e := found[min(i, len(found)-1)]
	if e.unavailable() {
		return p, Result{Refused: e.Title + " is unavailable: " + e.Unavailable}
	}
	return p, Result{CommandID: e.ID, Closed: true}
}

// Scroll moves the cursor by delta matches (the wheel).
func (p Palette) Scroll(delta int) Palette {
	p.cursor = min(max(p.cursor+delta, 0), max(len(p.matches())-1, 0))
	return p
}

// Hover highlights the match at index i.
func (p Palette) Hover(i int) Palette {
	if i >= 0 && i < len(p.matches()) {
		p.cursor = i
	}
	return p
}

// RowAt maps a line of the list area (0 is the first list line, below the
// input) to the index of the match drawn there, or -1 (a heading is no match);
// height is the number of list lines.
func (p Palette) RowAt(y, height int) int {
	rows := p.rows()
	i := p.windowTop(rows, height) + y
	if y < 0 || y >= height || i >= len(rows) {
		return -1
	}
	return rows[i].index
}

// windowTop is the first row shown so that the cursor is in the window; the
// heading above the first match of a section stays with it.
func (p Palette) windowTop(rows []row, height int) int {
	at := 0
	for i, r := range rows {
		if r.index == p.cursor {
			at = i
		}
	}
	top := min(p.top, at)
	if at >= top+height {
		top = at - height + 1
	}
	if top == at && at > 0 && rows[at-1].index < 0 {
		top = at - 1
	}
	return max(top, 0)
}

// Lines draws the palette in w cells and h lines: the input, the list, and a
// footer of key hints.
func (p Palette) Lines(th theme.Theme, w, h int) []string {
	out := []string{th.Accent(th.Icons.Pointer+" ") + p.filter.View()}
	height := max(h-2, 1)
	found := p.matches()
	if len(found) == 0 {
		out = append(out, th.Muted(" no command matches"))
	}
	rows := p.rows()
	top := p.windowTop(rows, height)
	for i := top; i < len(rows) && i < top+height; i++ {
		if r := rows[i]; r.index < 0 {
			out = append(out, " "+th.Bold(th.Accent(r.heading)))
		} else {
			out = append(out, p.line(th, found[r.index], r.index == p.cursor, w))
		}
	}
	footer := "up/down choose · enter pick · esc close"
	if n := p.hidden(); n > 0 && strings.TrimSpace(p.filter.Value()) == "" {
		footer = fmt.Sprintf("%d more not usable here: type to find them", n)
	}
	for len(out) < h-1 {
		out = append(out, "")
	}
	out = append(out, th.Muted(pane.Cut(footer, w)))
	return pane.Block(out, w, h)
}

func (p Palette) line(th theme.Theme, e Entry, selected bool, w int) string {
	right, desc := e.Shortcut, e.Hint
	if e.unavailable() {
		right, desc = "unavailable", e.Unavailable
	}
	// One cell on each side, two after the name and one before the shortcut.
	room := max(w-2-2-1-pane.Width(e.Title)-pane.Width(right), 0)
	if pane.Width(desc) > room {
		desc = strings.TrimRight(pane.Cell(desc, max(room-1, 0)), " ") + "…"
		if room < 2 {
			desc = ""
		}
	}
	gap := max(w-2-2-pane.Width(e.Title)-pane.Width(desc)-pane.Width(right), 1)
	if selected {
		return th.Selected(pane.Fit(" "+e.Title+"  "+desc+strings.Repeat(" ", gap)+right+" ", w))
	}
	return pane.Fit(" "+e.Title+"  "+th.Muted(desc)+strings.Repeat(" ", gap)+th.Bold(right)+" ", w)
}
