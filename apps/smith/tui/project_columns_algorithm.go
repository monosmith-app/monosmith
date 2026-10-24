package tui

// Widths of the Projects table's columns, in cells.
const (
	kindColW  = 8
	langColW  = 10
	stackColW = 14
	depsColW  = 6
	// tasksFloorW is the least the Tasks column may get while a droppable
	// column is still shown.
	tasksFloorW = 30
	// projectLeadW is the dot and the spaces before the name, and the one after.
	projectLeadW = 4
)

// projectColumns says how the Projects table spends w cells: the project
// name (nameWanted at most, ellipsised when short of room), which of the
// optional columns Stack, Language and Kind stay (they go in that order; Stack
// only when some project has one, an all-empty column reads as a bug), and
// what is left for Tasks. Tasks is never squeezed to nothing to keep an
// optional column.
type projectColumns struct {
	name              int
	kind, lang, stack bool
	tasks             int
}

func columnsForProjects(w, nameWanted, tasksWanted int, hasStack bool) projectColumns {
	name := min(nameWanted, max(w/3, 8))
	need := min(tasksWanted, tasksFloorW)
	for _, c := range []projectColumns{
		{kind: true, lang: true, stack: hasStack},
		{kind: true, lang: true},
		{kind: true},
		{},
	} {
		optional := 0
		if c.kind {
			optional += kindColW
		}
		if c.lang {
			optional += langColW
		}
		if c.stack {
			optional += stackColW
		}
		left := w - projectLeadW - name - depsColW - optional
		if left >= need {
			c.name, c.tasks = name, left
			return c
		}
	}
	// Even the bare table is short of room: the name gives way.
	name = max(w-projectLeadW-depsColW-need, 8)
	return projectColumns{name: name, tasks: max(w-projectLeadW-name-depsColW, 0)}
}
