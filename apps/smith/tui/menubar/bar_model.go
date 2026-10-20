package menubar

import (
	"strings"
	"unicode"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Bar is the menu bar: headings on one row and, under the open one, a box of
// entries. The zero value has no menus.
type Bar struct {
	menus []Menu
	// open is the menu whose box shows, or -1; cursor is the entry highlighted
	// in it; heading is the heading the keyboard is on; active means the
	// keyboard has taken the bar (F10, Alt+letter, a click).
	open, cursor, heading int
	active                bool
	// width is the room the bar has, to keep the box on screen.
	width int
	// brand is the app name drawn before the headings, after which a pipe
	// separates it from them.
	brand string
}

// New returns a closed bar.
func New(menus []Menu, width int) Bar {
	return Bar{menus: menus, open: -1, cursor: -1, width: width}
}

// SetMenus replaces the menus (their availability changes with the state),
// keeping the open menu and the highlighted entry.
func (b Bar) SetMenus(menus []Menu) Bar {
	b.menus = menus
	if b.open >= len(menus) {
		b.open = -1
	}
	return b
}

// SetWidth tells the bar how wide the screen is.
func (b Bar) SetWidth(w int) Bar { b.width = w; return b }

// SetBrand sets the app name shown before the headings.
func (b Bar) SetBrand(name string) Bar { b.brand = name; return b }

// brandWidth is the room the name and its pipe take: a space, the name, a
// space and the pipe.
func (b Bar) brandWidth() int {
	if b.brand == "" {
		return 1
	}
	return pane.Width(b.brand) + 3
}

// Menus are the menus.
func (b Bar) Menus() []Menu { return b.menus }

// Open is the open menu's index, or -1.
func (b Bar) Open() int { return b.open }

// Active reports whether the bar has the keyboard.
func (b Bar) Active() bool { return b.active || b.open >= 0 }

// Close shuts the box and gives the keyboard back.
func (b Bar) Close() Bar { b.open, b.cursor, b.active = -1, -1, false; return b }

// OpenMenu opens menu i with the first entry highlighted.
func (b Bar) OpenMenu(i int) Bar {
	if i < 0 || i >= len(b.menus) {
		return b.Close()
	}
	b.open, b.heading, b.active = i, i, true
	b.cursor = b.box().step(-1, 1)
	return b
}

// Cursor is the entry highlighted in the open menu, or -1.
func (b Bar) Cursor() int { return b.cursor }

// Tooltip is the reason the highlighted entry is unavailable, or its hint.
func (b Bar) Tooltip() string {
	if b.open < 0 || b.cursor < 0 {
		return ""
	}
	e := b.menus[b.open].Entries[b.cursor]
	if e.Disabled != "" {
		return e.Name + " is unavailable: " + e.Disabled
	}
	return e.Hint
}

func (b Bar) box() dropdown {
	if b.open < 0 {
		return dropdown{}
	}
	d := dropdown{entries: b.menus[b.open].Entries, cursor: b.cursor, Y: 1}
	d.X = min(b.headingX()[b.open], max(b.width-d.innerWidth()-2, 0))
	return d
}

// headingX are the columns where each heading's label (with its padding) starts.
func (b Bar) headingX() []int {
	xs := make([]int, len(b.menus))
	x := b.brandWidth()
	for i, m := range b.menus {
		xs[i] = x
		x += pane.Width(m.Name) + 3
	}
	return xs
}

// HeadingAt is the heading under column x of the bar's row, or -1.
func (b Bar) HeadingAt(x int) int {
	for i, x0 := range b.headingX() {
		if x >= x0 && x < x0+pane.Width(b.menus[i].Name)+2 {
			return i
		}
	}
	return -1
}

// Click handles a left press at (x, y) in screen coordinates (the bar is row 0).
func (b Bar) Click(x, y int) (Bar, Result) {
	if y == 0 {
		if i := b.HeadingAt(x); i >= 0 {
			if b.open == i {
				return b.Close(), Result{Handled: true}
			}
			return b.OpenMenu(i), Result{Handled: true}
		}
	}
	if b.open >= 0 {
		if j := b.box().entryAt(x, y); j >= 0 {
			return b.choose(j)
		}
		if b.box().contains(x, y) {
			return b, Result{Handled: true}
		}
		return b.Close(), Result{Handled: y != 0}
	}
	return b, Result{}
}

// Hover moves the highlight under the pointer; with a menu open, hovering
// another heading opens it, as in a desktop menu bar.
func (b Bar) Hover(x, y int) Bar {
	if b.open < 0 {
		return b
	}
	if y == 0 {
		if i := b.HeadingAt(x); i >= 0 && i != b.open {
			b = b.OpenMenu(i)
			b.cursor = -1
		}
		return b
	}
	if j := b.box().entryAt(x, y); j >= 0 {
		b.cursor = j
	}
	return b
}

// Key handles a key name (Bubble Tea's names: "f10", "alt+f", "down", "enter",
// "esc"). Shortcuts of entries are the shell's business, not the bar's.
func (b Bar) Key(name string) (Bar, Result) {
	if len(b.menus) == 0 {
		return b, Result{}
	}
	if name == "f10" {
		if b.Active() {
			return b.Close(), Result{Handled: true}
		}
		return b.OpenMenu(b.heading), Result{Handled: true}
	}
	if letter, ok := strings.CutPrefix(name, "alt+"); ok && len([]rune(letter)) == 1 {
		r := []rune(letter)[0]
		for i, m := range b.menus {
			if unicode.ToLower(mnemonic(m)) == unicode.ToLower(r) {
				return b.OpenMenu(i), Result{Handled: true}
			}
		}
		return b, Result{}
	}
	if !b.Active() {
		return b, Result{}
	}
	switch name {
	case "left", "right":
		dir := 1
		if name == "left" {
			dir = -1
		}
		next := ((b.heading+dir)%len(b.menus) + len(b.menus)) % len(b.menus)
		if b.open >= 0 {
			return b.OpenMenu(next), Result{Handled: true}
		}
		b.heading = next
	case "down", "up":
		if b.open < 0 {
			if name == "down" {
				return b.OpenMenu(b.heading), Result{Handled: true}
			}
			return b, Result{Handled: true}
		}
		dir := 1
		if name == "up" {
			dir = -1
		}
		from := b.cursor
		if from < 0 && dir < 0 {
			from = 0
		}
		b.cursor = b.box().step(from, dir)
	case "home":
		if b.open >= 0 {
			b.cursor = b.box().step(-1, 1)
		}
	case "end":
		if b.open >= 0 {
			b.cursor = b.box().step(0, -1)
		}
	case "enter", " ":
		if b.open < 0 {
			return b.OpenMenu(b.heading), Result{Handled: true}
		}
		if b.cursor >= 0 {
			return b.choose(b.cursor)
		}
	case "esc":
		if b.open >= 0 {
			b.open, b.cursor = -1, -1
		} else {
			b.active = false
		}
	default:
		// Any other key while a menu is open is swallowed so it cannot reach a view.
		return b, Result{Handled: b.open >= 0}
	}
	return b, Result{Handled: true}
}

func (b Bar) choose(j int) (Bar, Result) {
	e := b.menus[b.open].Entries[j]
	if e.Disabled != "" {
		b.cursor = j
		return b, Result{Handled: true, Refused: e.Name + " is unavailable: " + e.Disabled}
	}
	return b.Close(), Result{Handled: true, Chosen: e.ID}
}

func mnemonic(m Menu) rune {
	if m.Mnemonic != 0 {
		return m.Mnemonic
	}
	for _, r := range m.Name {
		return r
	}
	return 0
}

// Row draws the bar's row: the headings with their mnemonic letters
// underlined, and a title at the right, padded to w cells.
func (b Bar) Row(th theme.Theme, title string, w int) string {
	var s strings.Builder
	if b.brand == "" {
		s.WriteString(" ")
	} else {
		s.WriteString(" " + th.Accent(th.Bold(b.brand)) + " " + th.Faint("│"))
	}
	for i, m := range b.menus {
		label := m.Name
		idx := strings.IndexRune(strings.ToLower(label), unicode.ToLower(mnemonic(m)))
		text := label
		if idx >= 0 {
			text = label[:idx] + "\x1b[4m" + label[idx:idx+1] + "\x1b[24m" + label[idx+1:]
		}
		cell := " " + text + " "
		if i == b.open || (i == b.heading && b.active) {
			cell = th.Selected(cell)
		}
		s.WriteString(cell + " ")
	}
	row := s.String()
	if title != "" {
		if gap := w - pane.Width(row) - pane.Width(title) - 1; gap > 0 {
			row += strings.Repeat(" ", gap) + th.Accent(title) + " "
		}
	}
	return th.Bar(pane.Fit(row, w))
}

// Box draws the open menu's box and says where it goes; ok is false when no
// menu is open.
func (b Bar) Box(th theme.Theme) (lines []string, x, y int, ok bool) {
	if b.open < 0 {
		return nil, 0, 0, false
	}
	d := b.box()
	return d.view(th), d.X, d.Y, true
}
