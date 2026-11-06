package commands

// NavAction is what a key means to a list, a dialog or a menu: the one
// vocabulary of movement and answer keys. A view switches on it and never on
// the raw key name, so a key is bound to an action in one place (Nav) and not
// in every view.
//
// The keymap (keymap_config.go) is where commands are bound to keys for the
// help, the footer and the palette. Nav is deliberately not in it: the
// movement keys already show there as one line per view ("projects.move"),
// and a second entry per action would list them twice.
type NavAction int

// The navigation actions.
const (
	// NavNone is a key that has no navigation meaning.
	NavNone NavAction = iota
	// NavUp and NavDown move by one row.
	NavUp
	NavDown
	// NavPageUp and NavPageDown move by a page.
	NavPageUp
	NavPageDown
	// NavHome and NavEnd move to the first and the last row.
	NavHome
	NavEnd
	// NavConfirm accepts the row or the dialog: enter or space.
	NavConfirm
	// NavCancel leaves: esc, ctrl+c or q.
	NavCancel
	// NavNext and NavPrev move between fields or buttons: tab and shift+tab.
	NavNext
	NavPrev
	// NavYes and NavNo answer a question: y and n.
	NavYes
	NavNo
)

// navNamed are the keys that are not typed characters: they are safe to read
// as navigation while a text field has the keyboard.
var navNamed = map[string]NavAction{
	"up":        NavUp,
	"down":      NavDown,
	"pgup":      NavPageUp,
	"pgdown":    NavPageDown,
	"home":      NavHome,
	"end":       NavEnd,
	"enter":     NavConfirm,
	"esc":       NavCancel,
	"ctrl+c":    NavCancel,
	"tab":       NavNext,
	"shift+tab": NavPrev,
}

// navLetters are the keys that are also letters or the space bar: they mean
// navigation only where no text is being typed.
var navLetters = map[string]NavAction{
	"k":     NavUp,
	"j":     NavDown,
	" ":     NavConfirm,
	"space": NavConfirm,
	"q":     NavCancel,
	"y":     NavYes,
	"n":     NavNo,
}

// NavNamed maps a key name to its navigation action, for a screen that has a
// text field: only the named keys (arrows, page keys, enter, esc, tab) count,
// so the letters stay with the field. It returns NavNone for any other key.
func NavNamed(name string) NavAction { return navNamed[name] }

// Nav maps a key name, as Bubble Tea spells it ("up", "k", "pgdown", " "), to
// its navigation action, for a screen that has no text field: the letters
// k, j, q, y and n and the space bar count too. It returns NavNone for any
// other key.
func Nav(name string) NavAction {
	if a, ok := navNamed[name]; ok {
		return a
	}
	return navLetters[name]
}

// Moves reports whether the action moves a cursor: up, down, a page, home or end.
func (a NavAction) Moves() bool { return a >= NavUp && a <= NavEnd }
