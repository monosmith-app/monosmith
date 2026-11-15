package choicerow

// Row is one row of a pick list or of a list of tick boxes: the one row every
// checkbox list, choice list and option list of the TUI draws (see Lines).
type Row struct {
	// Lead is drawn between the cursor mark and the label: a tick box from Box,
	// a radio mark from Radio and a space, a state icon, or nothing.
	Lead string
	// Label is what the row is; Hint is said after it, muted.
	Label, Hint string
	// Selected is the row at the keyboard cursor: the cursor mark and the
	// accent label. Hovered is the row under the pointer: the row hover
	// background of the theme. Disabled is a row that cannot be chosen now: a
	// muted label.
	Selected, Hovered, Disabled bool
}

// Ticks is a row with a tick box in front of the label. It is the row of every
// list where several things may be ticked.
//
// Parameters: label is the row's text; checked and tri are as for Box.
// Returns: the row, not selected and not hovered.
func Ticks(label string, checked, tri bool) Row {
	return Row{Lead: Box(checked, tri) + " ", Label: label}
}
