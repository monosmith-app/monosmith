package projectlist

import (
	"slices"
	"sort"
)

// groupOrder is the order the kinds are listed in; other kinds follow.
var groupOrder = []string{"app", "lib", "package"}

// List is the cursor over the projects. It is a value: every method returns
// the new list.
type List struct {
	ordered []Project
	byName  map[string]Project
	cursor  int
	// filter is the text that chooses the projects listed (see Matches).
	filter string
}

// New builds the list: projects by kind (apps, libs, packages, then other
// kinds), then by name.
func New(projects []Project) List {
	l := List{byName: map[string]Project{}}
	for _, p := range projects {
		l.byName[p.Name] = p
	}
	l.ordered = append([]Project(nil), projects...)
	sort.SliceStable(l.ordered, func(i, j int) bool {
		a, b := l.ordered[i], l.ordered[j]
		if ra, rb := kindRank(a.Kind), kindRank(b.Kind); ra != rb {
			return ra < rb
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Name < b.Name
	})
	return l
}

func kindRank(kind string) int {
	for i, k := range groupOrder {
		if k == kind {
			return i
		}
	}
	return len(groupOrder)
}

// Projects returns the projects in the order they are listed.
func (l List) Projects() []Project { return l.ordered }

// Len is the number of projects.
func (l List) Len() int { return len(l.ordered) }

// Project returns the project with a name.
func (l List) Project(name string) (Project, bool) { p, ok := l.byName[name]; return p, ok }

// Cursor is the index of the selected project.
func (l List) Cursor() int { return l.cursor }

// Selected is the project under the cursor. It reports false when the list
// is empty, or when the filter lists no project.
func (l List) Selected() (Project, bool) {
	if l.cursor < 0 || l.cursor >= len(l.ordered) || !Matches(l.ordered[l.cursor], l.filter) {
		return Project{}, false
	}
	return l.ordered[l.cursor], true
}

// SelectedName is the name of the selected project, or "".
func (l List) SelectedName() string {
	p, _ := l.Selected()
	return p.Name
}

// SelectProject moves the cursor to a project.
func (l List) SelectProject(name string) List {
	for i, p := range l.ordered {
		if p.Name == name {
			l.cursor = i
		}
	}
	return l
}

// Select moves the cursor to index i of Projects.
func (l List) Select(i int) List {
	if i >= 0 && i < len(l.ordered) {
		l.cursor = i
	}
	return l
}

// Move moves the cursor by delta rows of the listed projects (those the
// filter keeps), clamped at the first and the last.
func (l List) Move(delta int) List {
	shown := l.shown()
	if len(shown) == 0 {
		return l
	}
	at := 0
	for i, idx := range shown {
		if idx == l.cursor {
			at = i
		}
	}
	l.cursor = shown[min(max(at+delta, 0), len(shown)-1)]
	return l
}

// Home and End jump to the first and last listed project.
func (l List) Home() List { return l.Move(-len(l.ordered)) }

// End jumps to the last listed project.
func (l List) End() List { return l.Move(len(l.ordered)) }

// WithFilter lists only the projects that pass filter (see Matches). The
// cursor stays on its project when it passes, else it moves to the first one
// that does.
//
// Parameters: filter is the text the person typed; "" lists every project.
// Returns: the filtered list.
func (l List) WithFilter(filter string) List {
	l.filter = filter
	if shown := l.shown(); len(shown) > 0 && !slices.Contains(shown, l.cursor) {
		l.cursor = shown[0]
	}
	return l
}

// Filter is the text that chooses the listed projects; "" lists them all.
func (l List) Filter() string { return l.filter }

// Listed are the projects the filter keeps, in the order of Projects, each
// with its index there.
//
// Returns: the projects and their indexes, in the same order.
func (l List) Listed() (projects []Project, indexes []int) {
	for _, i := range l.shown() {
		projects, indexes = append(projects, l.ordered[i]), append(indexes, i)
	}
	return projects, indexes
}

// shown are the indexes of the projects the filter keeps.
func (l List) shown() []int {
	var out []int
	for i, p := range l.ordered {
		if Matches(p, l.filter) {
			out = append(out, i)
		}
	}
	return out
}

// Page moves by a page of the given number of rows.
func (l List) Page(dir, rows int) List { return l.Move(dir * max(rows-1, 1)) }
