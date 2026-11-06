package releasepick

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// bumpLevels are the answers of the bump column: the commits decide (empty), or
// the person picks a level.
var bumpLevels = []string{"", "patch", "minor", "major"}

const minWidth = 56

// buttonKeys are the keys that do what a click on Continue and Cancel does.
var buttonKeys = [2]string{"enter", "esc"} // words:literal

// buttonGo is the label of the button that continues the release.
func buttonGo() string { return words.Say("releasepick.button.go") }

// buttonStop is the label of the button that cancels the release.
func buttonStop() string { return words.Say("releasepick.button.stop") }

// Model is the question before a release is confirmed: which packages of the
// plan to hold back (any number) and which bump to change (pick the package,
// then the level). It ends as the `smith release` line with --hold and --bump;
// the shell then shows the usual confirmation, which says the release cannot be
// undone.
//
// Keys: Up and Down pick a package; Space holds it back or lets it go; B, Left
// and Right change its bump; Enter goes on; Esc cancels. A click on a row's box
// holds it, on its bump changes it.
type Model struct {
	plan   release.Plan
	held   map[string]bool
	bump   map[string]int
	cursor int
	// top is the first row of the list on screen.
	top int
	// focus is -1 for the list, 0 for Continue and 1 for Cancel; hover is the
	// button under the pointer, or -1.
	focus, hover int
	rect         layout.Rect
	// sw and sh are the screen the dialog was placed on.
	sw, sh int
	// topMargin and bottomMargin are the screen rows the dialog stays clear of.
	topMargin, bottomMargin int
	// head is how many lines come before the list; room is how many rows the list has.
	head, room int
}

// New asks about the packages of a plan.
//
// Parameters: plan is the release plan as the commits decide it.
// Returns: the dialog, with nothing held and no bump changed.
func New(plan release.Plan) Model {
	return Model{plan: plan, held: map[string]bool{}, bump: map[string]int{}, focus: -1, hover: -1}
}

// Within keeps the dialog off the top rows (header, tab bar) and the bottom rows (status bar).
func (m Model) Within(top, bottom int) Model {
	m.topMargin, m.bottomMargin = max(top, 0), max(bottom, 0)
	return m
}

// Title is the heading of the dialog.
func (m Model) Title() string { return words.Say("releasepick.title") }

// canBump is true when the commits decide the version of a package that was
// released before. A first release, a forced version and a channel release
// have a version of their own.
func (m Model) canBump(it release.Item) bool {
	return it.Current != "" && !it.Forced && m.plan.Channel == ""
}

// Argv is the release command line for the answers so far, without --yes.
//
// Returns: release, with --hold <names> and one --bump <name>=<level> per change.
func (m Model) Argv() []string {
	argv := []string{"release"}
	var hold []string
	for _, it := range m.plan.Items {
		if m.held[it.Project] {
			hold = append(hold, it.Project)
		}
	}
	if len(hold) > 0 {
		argv = append(argv, "--hold", strings.Join(hold, ","))
	}
	for _, it := range m.plan.Items {
		if level := bumpLevels[m.bump[it.Project]]; level != "" && !m.held[it.Project] && m.canBump(it) {
			argv = append(argv, "--bump", it.Project+"="+level)
		}
	}
	return argv
}

// adjusted is the plan after the answers, by project, so each row shows the version it gets.
func (m Model) adjusted() map[string]release.Item {
	out := map[string]release.Item{}
	plan, _, err := release.Adjust(m.plan, release.Adjustments{Hold: m.holds(), Bump: m.bumps()})
	if err != nil {
		return out
	}
	for _, it := range plan.Items {
		out[it.Project] = it
	}
	return out
}

func (m Model) holds() []string {
	var out []string
	for _, it := range m.plan.Items {
		if m.held[it.Project] {
			out = append(out, it.Project)
		}
	}
	return out
}

func (m Model) bumps() map[string]string {
	out := map[string]string{}
	for _, it := range m.plan.Items {
		if level := bumpLevels[m.bump[it.Project]]; level != "" && !m.held[it.Project] && m.canBump(it) {
			out[it.Project] = level
		}
	}
	return out
}

// Place centres the dialog on a screen of sw by sh cells, between the margins.
func (m Model) Place(sw, sh int) Model {
	m.sw, m.sh = sw, sh
	w := min(max(sw-8, minWidth), min(max(sw-2, 10), 84))
	inner := w - 4
	head := append(intro(inner), "")
	tail := m.tailLines(inner)
	avail := max(sh-m.topMargin-m.bottomMargin, 3)
	want := len(head) + len(m.plan.Items) + len(tail) + 4
	h := min(want, avail)
	m.head = len(head)
	m.room = max(h-4-len(head)-len(tail), 1)
	m.rect = layout.Rect{X: max((sw-w)/2, 0), Y: m.topMargin + max((avail-h)/2, 0), W: w, H: h}
	return m.follow()
}

// follow scrolls the list so the row under the cursor is on screen.
func (m Model) follow() Model {
	switch {
	case m.cursor < m.top:
		m.top = m.cursor
	case m.cursor >= m.top+m.room:
		m.top = m.cursor - m.room + 1
	}
	m.top = max(min(m.top, max(len(m.plan.Items)-m.room, 0)), 0)
	return m
}

// Rect is where the dialog sits after Place.
func (m Model) Rect() layout.Rect { return m.rect }

// buttonLabels are the labels and keys of the Continue and Cancel buttons.
func buttonLabels() (labels, keys []string) {
	return []string{buttonGo(), buttonStop()}, buttonKeys[:]
}

// buttons are the spans, from the dialog's left edge, of Continue and Cancel.
func (m Model) buttons(inner int) [][2]int {
	labels, keys := buttonLabels()
	return button.RowSpans(theme.Default(), labels, keys, inner)
}

// rowAt is the package at a cell, or -1.
func (m Model) rowAt(x, y int) int {
	r := m.rect
	i := y - r.Y - 1 - m.head + m.top
	if x <= r.X || x >= r.X+r.W-1 || y-r.Y-1-m.head < 0 || y-r.Y-1-m.head >= m.room || i < 0 || i >= len(m.plan.Items) {
		return -1
	}
	return i
}

// Hover highlights the button under the pointer.
func (m Model) Hover(x, y int) Model {
	m.hover = button.At(m.buttons(m.rect.W-4), m.rect, x, y)
	return m
}

// Click handles a left press: a button goes on or cancels, a row picks that
// package and holds it back, the bump column changes the bump, and a press
// outside the dialog cancels.
func (m Model) Click(x, y int) (Model, dialog.Outcome) {
	switch button.At(m.buttons(m.rect.W-4), m.rect, x, y) {
	case 0:
		return m, dialog.Outcome{Done: true, Confirmed: true}
	case 1:
		return m, dialog.Outcome{Done: true}
	}
	if i := m.rowAt(x, y); i >= 0 {
		m.cursor, m.focus = i, -1
		if x-m.rect.X <= 6 {
			return m.toggle().Place(m.sw, m.sh), dialog.Outcome{}
		}
		return m.cycle(1).Place(m.sw, m.sh), dialog.Outcome{}
	}
	if !m.rect.Contains(x, y) {
		return m, dialog.Outcome{Done: true}
	}
	return m, dialog.Outcome{}
}

// toggle holds the package under the cursor back, or lets it go.
func (m Model) toggle() Model {
	if m.cursor < 0 || m.cursor >= len(m.plan.Items) {
		return m
	}
	held := maps2(m.held)
	name := m.plan.Items[m.cursor].Project
	held[name] = !held[name]
	m.held = held
	return m
}

// cycle changes the bump of the package under the cursor by dir.
func (m Model) cycle(dir int) Model {
	if m.cursor < 0 || m.cursor >= len(m.plan.Items) {
		return m
	}
	it := m.plan.Items[m.cursor]
	if !m.canBump(it) || m.held[it.Project] {
		return m
	}
	bump := maps2(m.bump)
	bump[it.Project] = (bump[it.Project] + dir + len(bumpLevels)) % len(bumpLevels)
	m.bump = bump
	return m
}

func maps2[V any](in map[string]V) map[string]V {
	out := make(map[string]V, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// Key handles a key name.
func (m Model) Key(name string) (Model, dialog.Outcome) {
	switch commands.Nav(name) {
	case commands.NavUp:
		m.focus = -1
		m.cursor = max(m.cursor-1, 0)
	case commands.NavDown:
		m.focus = -1
		m.cursor = min(m.cursor+1, max(len(m.plan.Items)-1, 0))
	default:
		switch name {
		case "esc", "ctrl+c", "q", "n":
			return m, dialog.Outcome{Done: true}
		case "enter":
			return m, dialog.Outcome{Done: true, Confirmed: m.focus != 1}
		case " ", "space", "h":
			m.focus = -1
			m = m.toggle()
		case "b", "right", "l":
			if m.focus >= 0 {
				m.focus = 1 - m.focus
				break
			}
			m = m.cycle(1)
		case "left":
			if m.focus >= 0 {
				m.focus = 1 - m.focus
				break
			}
			m = m.cycle(-1)
		case "tab", "shift+tab":
			m.focus = (m.focus+2)%3 - 1
		}
	}
	return m.Place(m.sw, m.sh), dialog.Outcome{}
}

// Held reports whether a package is held back, for the tests and the shell.
func (m Model) Held(name string) bool { return slices.Contains(m.holds(), name) }
