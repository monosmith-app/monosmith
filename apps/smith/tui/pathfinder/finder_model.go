package pathfinder

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
)

// Model is the list under a path field. It holds the folder read last, the
// rows the typed text leaves and the highlighted row; the text itself stays in
// the host's field.
type Model struct {
	root     string
	settings []Setting
	rules    ignoreRules
	ruled    bool

	dir        string
	listed     bool
	raw        []rawEntry
	truncated  bool
	unreadable bool

	value  string
	rows   []row
	cursor int
	// active is true once the person moved or clicked in the list: only then
	// does Enter act on the highlighted row, otherwise it confirms the text.
	active bool
}

// Outcome is what a key did to the field.
type Outcome struct {
	// Handled is true when the list took the key: the host must not give it to
	// the field.
	Handled bool
	// Set is true when the field's text must become Value.
	Set   bool
	Value string
	// Confirm is true when the value is chosen: the host does what Enter does.
	Confirm bool
}

// New makes a list over the folder root.
//
// Parameters: root is the folder paths start from ("" or "." for the folder
// smith started in); settings are names, not files, that the field also takes.
// Returns: the list, which reads nothing until Sync.
func New(root string, settings []Setting) Model {
	if root == "" {
		root = "."
	}
	return Model{root: root, settings: settings}
}

// Sync tells the list what is typed now. It reads the folder only when the
// typed text enters another one, and puts the highlight back on the first row
// when the text changed.
//
// Parameters: value is the text of the field.
// Returns: the list for that text.
func (m Model) Sync(value string) Model {
	dir, prefix := splitValue(value)
	if !m.ruled {
		m.rules, m.ruled = loadIgnoreRules(m.root), true
	}
	if !m.listed || dir != m.dir {
		raw, truncated, err := listFolder(m.root, dir)
		m.dir, m.listed, m.raw, m.truncated, m.unreadable = dir, true, raw, truncated, err != nil
	}
	if value != m.value {
		m.cursor, m.active = 0, false
	}
	m.value = value
	m.rows = buildRows(dir, prefix, m.raw, m.settings, m.rules.Ignored)
	return m
}

// Count is how many rows the list has.
func (m Model) Count() int { return len(m.rows) }

// Cursor is the highlighted row.
func (m Model) Cursor() int { return m.cursor }

// Select highlights row i, as a click does; Enter then acts on it.
func (m Model) Select(i int) Model {
	if i >= 0 && i < len(m.rows) {
		m.cursor, m.active = i, true
	}
	return m
}

// Handle gives the list a key. Up and Down (and the page keys) move the
// highlight; Tab puts the highlighted name in the field; Enter, after the
// highlight was moved, opens a folder or chooses a file, and otherwise is left
// to the host, which confirms the typed text.
//
// Parameters: key is the key press.
// Returns: the list and what the key did to the field.
func (m Model) Handle(key tea.KeyMsg) (Model, Outcome) {
	switch commands.NavNamed(key.String()) {
	case commands.NavUp:
		return m.move(-1), Outcome{Handled: true}
	case commands.NavDown:
		return m.move(1), Outcome{Handled: true}
	case commands.NavPageUp:
		return m.move(-pageRows), Outcome{Handled: true}
	case commands.NavPageDown:
		return m.move(pageRows), Outcome{Handled: true}
	case commands.NavNext:
		if len(m.rows) == 0 {
			return m, Outcome{Handled: true}
		}
		r := m.rows[m.cursor]
		return m.Sync(r.value), Outcome{Handled: true, Set: true, Value: r.value}
	case commands.NavConfirm:
		if !m.active || len(m.rows) == 0 {
			return m, Outcome{}
		}
		r := m.rows[m.cursor]
		out := Outcome{Handled: true, Set: true, Value: r.value, Confirm: r.kind == rowFile || r.kind == rowSetting}
		return m.Sync(r.value), out
	}
	return m, Outcome{}
}

// pageRows is how far a page key moves the highlight.
const pageRows = 5

func (m Model) move(by int) Model {
	if len(m.rows) == 0 {
		return m
	}
	m.cursor, m.active = min(max(m.cursor+by, 0), len(m.rows)-1), true
	return m
}
