package projectlist

// Widths of the Projects table's columns, in cells.
const (
	KindW  = 8
	LangW  = 10
	StackW = 14
	DepsW  = 6
	// TasksFloorW is the least the Tasks column may get while a droppable
	// column is still shown.
	TasksFloorW = 30
	// LeadW is the dot and the spaces before the name, and the one after.
	LeadW = 4
)

// Columns says how the Projects table spends w cells: the project
// name (nameWanted at most, ellipsised when short of room), which of the
// optional columns Stack, Language and Kind stay (they go in that order; Stack
// only when some project has one, an all-empty column reads as a bug), and
// what is left for Tasks. Tasks is never squeezed to nothing to keep an
// optional column.
type Columns struct {
	Name              int
	Kind, Lang, Stack bool
	Tasks             int
}

// ColumnsFor lays the table out for w cells.
func ColumnsFor(w, nameWanted, tasksWanted int, hasStack bool) Columns {
	name := min(nameWanted, max(w/3, 8))
	need := min(tasksWanted, TasksFloorW)
	for _, c := range []Columns{
		{Kind: true, Lang: true, Stack: hasStack},
		{Kind: true, Lang: true},
		{Kind: true},
		{},
	} {
		optional := 0
		if c.Kind {
			optional += KindW
		}
		if c.Lang {
			optional += LangW
		}
		if c.Stack {
			optional += StackW
		}
		left := w - LeadW - name - DepsW - optional
		if left >= need {
			c.Name, c.Tasks = name, left
			return c
		}
	}
	// Even the bare table is short of room: the name gives way.
	name = max(w-LeadW-DepsW-need, 8)
	return Columns{Name: name, Tasks: max(w-LeadW-name-DepsW, 0)}
}
