package menubar

// Entry is one line of a drop-down: a name, the shortcut shown beside it, and
// when it cannot be used, the reason why.
type Entry struct {
	// ID is what a choice reports; the shell maps it to a command.
	ID       string
	Name     string
	Shortcut string
	// Hint is a short explanation shown under the box while the entry is highlighted.
	Hint string
	// Disabled is the reason the entry cannot be chosen now; empty when it can.
	Disabled  string
	Separator bool
}

// Menu is one heading of the bar and the entries under it.
type Menu struct {
	Name string
	// Mnemonic is the letter that opens the menu with Alt; it is underlined
	// in the bar. Zero means the first letter of the name.
	Mnemonic rune
	Entries  []Entry
}

// Result is what a key or a click did to the bar.
type Result struct {
	// Chosen is the ID of the entry chosen, or "".
	Chosen string
	// Refused is the reason a disabled entry was not chosen.
	Refused string
	// Handled is true when the bar used the input, so the shell must not.
	Handled bool
}
