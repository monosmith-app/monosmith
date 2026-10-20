package palette

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Entry is a shell command (an id) listed with its shortcut. Unavailable is
// the reason it cannot be used now, or "".
type Entry struct {
	ID, Title, Hint, Shortcut, Unavailable string
}

// Result is what a key or a click did.
type Result struct {
	// CommandID is the shell command chosen.
	CommandID string
	// Argv is the smith command line confirmed (without the program name).
	Argv []string
	// Refused is the reason a shell command could not be chosen.
	Refused string
	// Closed is true when the palette should close.
	Closed bool
}

// item is one line of the list: a shell command or a smith command.
type item struct {
	entry *Entry
	cli   *clicontract.CommandDescription
}

func (i item) name() string {
	if i.entry != nil {
		return i.entry.Title
	}
	return i.cli.Name
}

func (i item) description() string {
	if i.entry != nil {
		return i.entry.Hint
	}
	return i.cli.Description
}

// Palette finds any command by typing part of its name or description;
// Enter takes it, then, for a smith command, the rest of the command line can
// be typed.
type Palette struct {
	items  []item
	cli    []clicontract.CommandDescription
	filter textinput.Model
	cursor int
	top    int
	// chosen is set once a smith command is picked; filter then holds its arguments.
	chosen *clicontract.CommandDescription
}

// New builds a palette of the shell's commands followed by smith's.
func New(entries []Entry, commands []clicontract.CommandDescription) Palette {
	in := textinput.New()
	in.Placeholder = "type to find a command"
	in.Focus()
	return Palette{filter: in, cli: commands}.WithEntries(entries)
}

// WithEntries replaces the shell commands (their availability changes with
// the state) and keeps what was typed.
func (p Palette) WithEntries(entries []Entry) Palette {
	p.items = nil
	for i := range entries {
		p.items = append(p.items, item{entry: &entries[i]})
	}
	for i := range p.cli {
		p.items = append(p.items, item{cli: &p.cli[i]})
	}
	p.cursor = min(p.cursor, max(len(p.items)-1, 0))
	return p
}

// matches are the items the filter finds: names starting with it first, then
// names containing it, then descriptions mentioning it.
func (p Palette) matches() []item {
	q := strings.ToLower(strings.TrimSpace(p.filter.Value()))
	var prefix, name, described []item
	for _, it := range p.items {
		n := strings.ToLower(it.name())
		switch {
		case strings.HasPrefix(n, q):
			prefix = append(prefix, it)
		case strings.Contains(n, q):
			name = append(name, it)
		case strings.Contains(strings.ToLower(it.description()), q):
			described = append(described, it)
		}
	}
	return append(append(prefix, name...), described...)
}

// Reset clears the filter and the choice, for reopening.
func (p Palette) Reset() Palette {
	p.filter.SetValue("")
	p.filter.Placeholder = "type to find a command"
	p.filter.Focus()
	p.chosen, p.cursor, p.top = nil, 0, 0
	return p
}

// Key handles a key press.
func (p Palette) Key(msg tea.KeyMsg) (Palette, Result) {
	switch msg.String() {
	case "esc", "ctrl+c":
		if p.chosen != nil {
			return p.Reset(), Result{}
		}
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

// Choose picks the item at index i of the matches (a click).
func (p Palette) Choose(i int) (Palette, Result) { return p.choose(i) }

func (p Palette) choose(i int) (Palette, Result) {
	if p.chosen != nil {
		return p, Result{Argv: append([]string{p.chosen.Name}, strings.Fields(p.filter.Value())...), Closed: true}
	}
	found := p.matches()
	if len(found) == 0 || i < 0 {
		return p, Result{}
	}
	it := found[min(i, len(found)-1)]
	if it.entry != nil {
		if it.entry.Unavailable != "" {
			return p, Result{Refused: it.entry.Title + " is unavailable: " + it.entry.Unavailable}
		}
		return p, Result{CommandID: it.entry.ID, Closed: true}
	}
	c := *it.cli
	p.chosen = &c
	p.filter.SetValue("")
	p.filter.Placeholder = usage(c)
	return p, Result{}
}

// Scroll moves the list by delta rows (the wheel).
func (p Palette) Scroll(delta int) Palette {
	p.cursor = min(max(p.cursor+delta, 0), max(len(p.matches())-1, 0))
	return p
}

// Hover highlights the match at index i.
func (p Palette) Hover(i int) Palette {
	if p.chosen == nil && i >= 0 && i < len(p.matches()) {
		p.cursor = i
	}
	return p
}

// RowAt maps a line of the list area (0 is the first list line, below the
// input) to the index of the match drawn there, or -1; height is the number of
// list lines.
func (p Palette) RowAt(y, height int) int {
	top := p.windowTop(height)
	i := top + y
	if p.chosen != nil || y < 0 || y >= height || i >= len(p.matches()) {
		return -1
	}
	return i
}

func (p Palette) windowTop(height int) int {
	top := p.top
	if p.cursor < top {
		top = p.cursor
	}
	if p.cursor >= top+height {
		top = p.cursor - height + 1
	}
	return max(top, 0)
}

// usage is a command's arguments and options as a placeholder.
func usage(c clicontract.CommandDescription) string {
	var parts []string
	for _, a := range c.Arguments {
		if a.Required {
			parts = append(parts, "<"+a.Name+">")
		} else {
			parts = append(parts, "["+a.Name+"]")
		}
	}
	for _, o := range c.Options {
		parts = append(parts, "[--"+o.Name+"]")
		if len(parts) > 6 {
			parts = append(parts, "…")
			break
		}
	}
	if len(parts) == 0 {
		return "enter to run"
	}
	return strings.Join(parts, " ")
}

// Chosen is the smith command being completed, or nil.
func (p Palette) Chosen() *clicontract.CommandDescription { return p.chosen }

// Lines draws the palette in w cells and h lines: the input, a rule, the
// list, and a footer of key hints.
func (p Palette) Lines(th theme.Theme, w, h int) []string {
	var out []string
	if p.chosen != nil {
		out = append(out, th.Title("smith "+p.chosen.Name)+" "+p.filter.View(), "", th.Muted(p.chosen.Description))
		out = append(out, "", th.Muted("enter run · esc back"))
		return pane.Block(out, w, h)
	}
	out = append(out, th.Accent(th.Icons.Pointer+" ")+p.filter.View())
	height := max(h-3, 1)
	found := p.matches()
	if len(found) == 0 {
		out = append(out, th.Muted(" no command matches"))
	}
	top := p.windowTop(height)
	for i := top; i < len(found) && i < top+height; i++ {
		out = append(out, p.line(th, found[i], i == p.cursor, w))
	}
	for len(out) < h-1 {
		out = append(out, "")
	}
	out = append(out, th.Muted("up/down choose · enter pick · esc close"))
	return pane.Block(out, w, h)
}

func (p Palette) line(th theme.Theme, it item, selected bool, w int) string {
	right, desc := "", it.description()
	switch {
	case it.entry != nil && it.entry.Unavailable != "":
		right, desc = "unavailable", it.entry.Unavailable
	case it.entry != nil:
		right = it.entry.Shortcut
	}
	name := it.name()
	room := max(w-3-pane.Width(name)-2-pane.Width(right), 0)
	desc = pane.Cut(desc, room)
	gap := max(w-3-pane.Width(name)-2-pane.Width(desc)-pane.Width(right), 0)
	if selected {
		return th.Selected(pane.Fit(" "+name+"  "+desc+strings.Repeat(" ", gap)+right+" ", w))
	}
	return pane.Fit(" "+name+"  "+th.Muted(desc)+strings.Repeat(" ", gap)+th.Accent(right)+" ", w)
}
