package projectpage

// ListRow is one project of the list.
type ListRow struct {
	// Index is the project's position in the shell's list; a click gives it back.
	Index                int
	Name, Kind, Language string
	// Version is "current → next" while a release is pending, else the
	// version; "" when there is none.
	Version string
	// Mark is how the project's tasks last ran.
	Mark Mark
	// Changed is true when the current changes touch the project.
	Changed bool
	// Picked is true when the person picked the project with Space.
	Picked bool
}

// ListInput is what the list of projects is drawn from.
type ListInput struct {
	// Rows are the projects the filter keeps, grouped by kind, in the order drawn.
	Rows []ListRow
	// Kinds are the kinds of every project in the order of the groups, and
	// Totals how many projects each kind has before the filter.
	Kinds  []string
	Totals map[string]int
	// Changed is how many projects the current changes touch, before the filter.
	Changed int
	// Filter is the text that chooses the rows; Typing is true while the person
	// types it, which shows the cursor.
	Filter string
	Typing bool
	// Selected is the Index of the selected row. Tick moves the spinner of a
	// running project.
	Selected int
	Tick     int
	// Picked is how many projects the person picked with Space; while it is
	// above zero every row shows its tick box and the count line leads with it.
	Picked int
}
