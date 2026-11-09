package projectlist

// Widths of the Projects table's columns, in cells.
const (
	KindW  = 8
	LangW  = 10
	StackW = 14
	DepsW  = 6
	// LeadW is the dot and the spaces before the name, and the one after.
	LeadW = 4
	// minNameW is the least the name column may get.
	minNameW = 8
)

// Columns says how the Projects table spends w cells: the project name
// (nameWanted at most, ellipsised when short of room) and which of the
// optional columns Kind, Language and Stack stay.
type Columns struct {
	Name              int
	Kind, Lang, Stack bool
}

// ColumnsFor lays the table out for w cells. The optional columns go in the
// order Stack, Language, Kind when the width is short; Stack only shows when
// some project has one, because an all-empty column reads as a bug.
func ColumnsFor(w, nameWanted int, hasStack bool) Columns {
	name := min(max(nameWanted, minNameW), max(w/3, minNameW))
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
		if w-LeadW-name-DepsW-optional >= 0 {
			c.Name = name
			return c
		}
	}
	// Even the bare table is short of room: the name gives way.
	return Columns{Name: max(w-LeadW-DepsW, minNameW)}
}
