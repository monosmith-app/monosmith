package palette

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/footer"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Entry is a shell command (an id) listed with its shortcut, under a section.
// Hint is the command line it runs. Keywords are words a person may type for
// the intent (synonyms): the search finds them, the list never shows them.
// Unavailable is the reason it cannot be used now, or "".
type Entry struct {
	ID, Section, Title, Hint, Shortcut, Keywords, Unavailable string
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

// selectedMark is the mark of the selected entry, the one of the Projects list.
const selectedMark = "▶ "

// Palette finds any command by typing part of its title, its command line or
// its section. The results are grouped under headings with their counts, and
// each shows the command line it runs under its title.
type Palette struct {
	items  []Entry
	filter textinput.Model
	cursor int
	top    int
	// guessed is true after Enter on one of several "Did you mean" rows: the
	// next Enter runs it, so a typo never runs a command at the first press.
	guessed bool
}

// New builds a palette of the shell's commands.
func New(entries []Entry) Palette {
	in := textinput.New()
	in.Placeholder = words.Say("palette.placeholder")
	in.Prompt = "" // the line draws its own pointer
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

// typed is the search as typed, trimmed.
func (p Palette) typed() string { return strings.TrimSpace(p.filter.Value()) }

// matches are the entries in the order the list draws them (see view): the
// cursor counts in this order.
func (p Palette) matches() []Entry {
	found, _ := p.view()
	return found
}

// view is the list as drawn: the matches in their order and the rows that
// draw them. A search lists its best match first, then the others by section,
// then the ones that cannot be used now; nothing typed lists the usable
// commands by section. The "Did you mean" list has no headings.
func (p Palette) view() ([]Entry, []row) {
	found, exact := p.search()
	if !exact {
		return found, flatRows(found)
	}
	return groupRows(found, p.typed() != "", headings{best: words.Say("palette.group.best"), unavailable: words.Say("palette.group.unavailable")})
}

// suggested is true when the filter found nothing and the list holds the
// closest commands instead.
func (p Palette) suggested() bool {
	_, exact := p.search()
	return !exact
}

// search finds the entries for the filter: titles starting with it first,
// then titles with it as the smith word in brackets, then titles containing
// it, then entries whose keywords carry every typed word, then hints and
// sections mentioning it, and the unusable ones last. Nothing typed gives the
// usable ones in section order. When nothing is found it gives the closest
// three commands and false.
func (p Palette) search() ([]Entry, bool) {
	q := strings.ToLower(p.typed())
	if q == "" {
		var usable []Entry
		for _, e := range p.items {
			if !e.unavailable() {
				usable = append(usable, e)
			}
		}
		return usable, true
	}
	var prefix, term, name, intent, described, blocked []Entry
	for _, e := range p.items {
		n := strings.ToLower(e.Title)
		var bucket *[]Entry
		switch {
		case strings.HasPrefix(n, q):
			bucket = &prefix
		case strings.Contains(n, "("+q):
			bucket = &term
		case strings.Contains(n, q):
			bucket = &name
		case hasAllWords(strings.ToLower(e.Keywords), q):
			bucket = &intent
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
	found := slices.Concat(prefix, term, name, intent, described, blocked)
	if len(found) == 0 {
		return closest(p.items, q), false
	}
	return found, true
}

// hasAllWords is true when every word of query is part of text.
func hasAllWords(text, query string) bool {
	for _, w := range strings.Fields(query) {
		if !strings.Contains(text, w) {
			return false
		}
	}
	return text != ""
}

// line is one line of the list area: the row of the heading or entry it draws
// (index -1 for a heading or a note), and second for the command line of an entry.
type line struct {
	r      row
	second bool
	note   string
}

// lines are the lines of the list area: a note on top for the "Did you mean"
// list, a heading on one line, an entry on two (its title, then its command line).
func (p Palette) lines() []line {
	_, rows := p.view()
	var out []line
	if p.suggested() && len(rows) > 0 {
		out = append(out, line{r: row{index: -1}, note: words.Say("palette.closest", "typed", p.typed())})
	}
	for _, r := range rows {
		out = append(out, line{r: r})
		if r.index >= 0 {
			out = append(out, line{r: r, second: true})
		}
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
	guessed := p.guessed
	p.guessed = false
	switch commands.NavNamed(msg.String()) {
	case commands.NavCancel:
		return p, Result{Closed: true}
	case commands.NavUp:
		p.cursor = max(p.cursor-1, 0)
	case commands.NavDown:
		p.cursor = min(p.cursor+1, max(len(p.matches())-1, 0))
	case commands.NavPageUp:
		p.cursor = max(p.cursor-8, 0)
	case commands.NavPageDown:
		p.cursor = min(p.cursor+8, max(len(p.matches())-1, 0))
	case commands.NavConfirm:
		if found := p.matches(); p.suggested() && len(found) > 1 && !guessed && !found[min(p.cursor, len(found)-1)].unavailable() {
			p.guessed = true
			return p, Result{Refused: words.Say("palette.guess.confirm", "title", found[min(p.cursor, len(found)-1)].Title)}
		}
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
		return p, Result{Refused: words.Say("palette.unavailable", "title", e.Title, "reason", e.Unavailable)}
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
// input and its rule) to the index of the match drawn there, or -1 (a heading
// or a note is no match). Both lines of an entry are the entry.
//
// Parameters: y is the line of the list area; height is the number of list lines.
// Returns: the index of the match, or -1.
func (p Palette) RowAt(y, height int) int {
	ls := p.lines()
	i := p.windowTop(ls, height) + y
	if y < 0 || y >= height || i >= len(ls) {
		return -1
	}
	return ls[i].r.index
}

// At maps a line inside the palette's frame to the index of the match drawn
// there, or -1: the input, the rules and the key hints are no match.
//
// Parameters: y is the line inside the frame (0 is the input); h is the number
// of lines inside the frame, as given to Lines.
// Returns: the index of the match, or -1.
func (p Palette) At(y, h int) int { return p.RowAt(y-listTop, listHeight(h)) }

// listTop is the first line of the list area inside the frame: under the input and its rule.
const listTop = 2

// listHeight is the number of list lines in h lines: all but the input, the
// two rules and the key hints.
func listHeight(h int) int { return max(h-4, 1) }

// windowTop is the first line shown so that both lines of the selected entry
// are in the window; the heading above the first entry of a group stays with it.
func (p Palette) windowTop(ls []line, height int) int {
	at := -1
	for i, l := range ls {
		if l.r.index == p.cursor && !l.second {
			at = i
			break
		}
	}
	if at < 0 {
		return 0
	}
	top := min(p.top, at)
	if at+1 >= top+height {
		top = at + 2 - height
	}
	if top == at && at > 0 && ls[at-1].r.index < 0 {
		top = at - 1
	}
	return max(min(top, at), 0)
}

// Lines draws the palette in w cells and h lines: the input with its rule,
// the grouped results, a rule and the key hints. The selected entry has the
// mark ▶ and the selected background; the typed letters show in the accent in
// every title.
//
// Parameters: th is the theme; w and h are the cells and lines inside the frame.
// Returns: h lines of w cells.
func (p Palette) Lines(th theme.Theme, w, h int) []string {
	esc := th.Muted(words.Say("palette.key.close"))
	input := th.Accent(th.Icons.Pointer+" ") + p.filter.View()
	if gap := w - pane.Width(input) - pane.Width(esc) - 1; gap > 0 {
		input += strings.Repeat(" ", gap) + esc
	}
	out := []string{input, th.Border(strings.Repeat("─", max(w, 0)), false)}
	height := listHeight(h)
	found := p.matches()
	ls := p.lines()
	if len(found) == 0 {
		out = append(out, th.Muted(pane.Cut(words.Say("palette.empty"), w)), "", th.Muted(pane.Cut(words.Say("palette.empty.hint"), w)))
	}
	top := p.windowTop(ls, height)
	for i := top; i < len(ls) && i < top+height; i++ {
		out = append(out, p.drawLine(th, ls[i], found, w))
	}
	for len(out) < h-2 {
		out = append(out, "")
	}
	out = append(out[:max(h-2, 0)], p.bottomRule(th, w), p.keyHints(th, w))
	return pane.Block(out, w, h)
}

// bottomRule is the rule over the key hints, with the count of the commands
// that cannot be used now while nothing is typed.
func (p Palette) bottomRule(th theme.Theme, w int) string {
	n := p.hidden()
	if n == 0 || p.typed() != "" {
		return th.Border(strings.Repeat("─", max(w, 0)), false)
	}
	note := " " + words.Say("palette.footer.hidden", "count", n) + " "
	if pane.Width(note)+4 > w {
		return th.Border(strings.Repeat("─", max(w, 0)), false)
	}
	return th.Border("──", false) + th.Muted(note) + th.Border(strings.Repeat("─", w-2-pane.Width(note)), false)
}

// keyHints is the line of key hints, in the look of the footer.
func (p Palette) keyHints(th theme.Theme, w int) string {
	hints := []commands.Hint{
		{Glyph: words.Say("palette.key.move"), Label: words.Say("palette.key.move.label")},
		{Glyph: words.Say("palette.key.run"), Label: words.Say("palette.key.run.label")},
		{Glyph: words.Say("palette.key.close"), Label: words.Say("palette.key.close.label")},
	}
	return footer.Render(th, footer.Place(th, hints, nil, w), -1, w)
}

// drawLine draws one line of the list area in w cells.
func (p Palette) drawLine(th theme.Theme, l line, found []Entry, w int) string {
	switch {
	case l.note != "":
		return th.Muted(pane.Cut(l.note, w))
	case l.r.index < 0:
		return pane.Fit(" "+th.Bold(th.Accent(l.r.heading))+"  "+th.Muted(words.Say("palette.group.count", "count", l.r.count)), w)
	}
	e := found[l.r.index]
	selected := l.r.index == p.cursor
	var text string
	if l.second {
		text = p.secondLine(th, e, w)
	} else {
		text = p.titleLine(th, e, selected, w)
	}
	if selected {
		return th.Selected(pane.Fit(text, w))
	}
	return pane.Fit(text, w)
}

// titleLine is the first line of an entry: the mark, the title with the typed
// letters in the accent, and the shortcut at the right end.
func (p Palette) titleLine(th theme.Theme, e Entry, selected bool, w int) string {
	mark := "  "
	plain := func(s string) string { return s }
	lit := func(s string) string { return th.Bold(th.Accent(s)) }
	if selected {
		mark = th.Accent(selectedMark)
		plain = th.Accent
	}
	right := th.Bold(e.Shortcut)
	if e.unavailable() {
		right = th.Muted(words.Say("palette.unavailable.tag"))
	}
	room := max(w-2-1-pane.Width(right)-1, 1)
	title := e.Title
	if pane.Width(title) > room {
		title = strings.TrimRight(pane.Cell(title, room), " ")
	}
	text := mark + highlight(title, matchSpans(title, p.typed()), plain, lit)
	gap := max(w-pane.Width(text)-pane.Width(right)-1, 1)
	return text + strings.Repeat(" ", gap) + right
}

// secondLine is the second line of an entry: the smith command line it runs,
// dim, or why it cannot be used now.
func (p Palette) secondLine(th theme.Theme, e Entry, w int) string {
	text := commandLine(e.Hint)
	if e.unavailable() {
		text = e.Unavailable
	}
	if room := max(w-4-1, 1); pane.Width(text) > room {
		text = strings.TrimRight(pane.Cell(text, room), " ")
	}
	return "    " + th.Faint(text)
}
