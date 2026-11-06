package detail

import ()

// Row is a label and its value.
type Row struct{ Label, Value string }

// Section is a heading and its rows.
type Section struct {
	Heading string
	Rows    []Row
	// Items are plain lines under the heading (a list of tasks or projects).
	Items []string
	// OneLine keeps each item on one line, ending in an ellipsis when it is
	// too long, instead of wrapping it (a command is not prose).
	OneLine bool
	// Bullets puts a "•" before each item and wraps it under its text.
	Bullets bool
	// Collapse is how many items show until the panel is Expanded; the rest
	// is told as a count. Zero shows them all.
	Collapse int
	// Tail are plain lines drawn after the items and never folded, so the
	// last line of a section (a terminal command) stays visible.
	Tail []string
}

// Action is a clickable line.
type Action struct {
	ID       string
	Label    string
	Shortcut string
	Disabled bool
	// Hint says, under the label, what the action does or what it needs.
	Hint string
	// Button draws the action as a button, ` Label (key) `, instead of a line with
	// the key at the right edge. The mouse clicks it the same way.
	Button bool
}

// Panel is what the detail shows.
type Panel struct {
	Sections []Section
	Actions  []Action
	// ActionsTitle is the heading over the actions; "Actions" when empty.
	ActionsTitle string
	// After are the sections drawn under the actions: what is long and
	// secondary, so it can never push the actions off the screen.
	After []Section
	// Expanded shows every item of a Collapse section.
	Expanded bool
}
