package detail

// Row is a label and its value.
type Row struct {
	Label, Value string
	// ShownByList marks a fact the highlighted row of the list beside the panel
	// already draws. The panel leaves the row out while ListVisible is set.
	ShownByList bool
}

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
	// Fold hides the whole section body until the panel is Expanded; the
	// heading stays, with a note that says how to show it.
	Fold bool
	// Tail are plain lines drawn after the items and never folded, so the
	// last line of a section (a terminal command) stays visible.
	Tail []string
}

// Action is a button: its label and its key, drawn by package button.
type Action struct {
	ID    string
	Label string
	// Shortcut is the key that does the same as a click; it shows in
	// parentheses after the label.
	Shortcut string
	Disabled bool
	// Hint says, under the label, what the action does or what it needs.
	Hint string
}

// Panel is what the detail shows.
type Panel struct {
	Sections []Section
	Actions  []Action
	// ActionsTitle is the heading over the actions; "Actions" when empty.
	ActionsTitle string
	// More are a second group of actions, drawn under the first with the heading
	// MoreTitle (the runs that cover every project, under those of one project).
	More      []Action
	MoreTitle string
	// After are the sections drawn under the actions: what is long and
	// secondary, so it can never push the actions off the screen.
	After []Section
	// Expanded shows every item of a Collapse section.
	Expanded bool
	// Pressed is the ID of the action the pointer holds down; it shows the
	// Active state while the pointer is still over it.
	Pressed string
	// ListVisible is true when the list the panel tells about is on screen next
	// to it; the rows marked ShownByList are then left out. It is false when the
	// panel fills the pane, so those rows stay.
	ListVisible bool
	// Cursor is the ID of the action the keyboard cursor is on. When it is set,
	// every action leads with the cursor mark (see package choicerow) and the one
	// at the cursor looks as if hovered. RunID puts the cursor on the run button
	// of a split panel.
	Cursor string
	// Chosen is the ID of the chosen action of a split panel (see Splits). An ID
	// that names no action chooses the first one.
	Chosen string
	// PinRun leaves the second part of a split panel (the description of the
	// chosen action and its run button) out of View, for a caller that draws it
	// with PinnedRun under the part that scrolls, so it stays in view.
	PinRun bool
}

// RunID is the hit, the hover and the cursor ID of the run button of a split
// panel: the one button under the description that runs the chosen action.
const RunID = "\x00run"

// Splits is true when the panel draws its actions in two parts: a compact
// column of buttons that only choose, then the description of the chosen action
// and one button that runs it. The rule is the same for every panel: two or
// more actions, and at least one with a description. One action, or actions
// with no description, keep buttons that run at a click.
func (p Panel) Splits() bool {
	all := p.all()
	if len(all) < 2 {
		return false
	}
	for _, a := range all {
		if a.Hint != "" {
			return true
		}
	}
	return false
}

// ChosenAction is the action the run button of a split panel runs: the one
// Chosen names, else the first. It reports false when the panel has no actions.
func (p Panel) ChosenAction() (Action, bool) {
	all := p.all()
	if len(all) == 0 {
		return Action{}, false
	}
	for _, a := range all {
		if a.ID == p.Chosen {
			return a, true
		}
	}
	return all[0], true
}

// all are the actions of both groups, in the order the keyboard visits them.
func (p Panel) all() []Action {
	return append(append([]Action(nil), p.Actions...), p.More...)
}
