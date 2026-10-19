package navigator

// Task is one task of a project, with what the inspector shows of it.
type Task struct {
	Name        string
	Command     string
	DependsOn   []string
	Cache       bool
	Description string
}

// Project is what the navigator and the inspector know of a project.
type Project struct {
	Name     string
	Kind     string
	Language string
	Root     string
	Stack    string
	Version  string
	Tasks    []Task
	// Deps are the projects it imports; Dependents are those that import it.
	Deps, Dependents []string
	// Affected is true when the current changes touch it.
	Affected bool
}

// Task returns the project's task with a name.
func (p Project) Task(name string) (Task, bool) {
	for _, t := range p.Tasks {
		if t.Name == name {
			return t, true
		}
	}
	return Task{}, false
}

// RowKind says what a row is.
type RowKind int

// The kinds of row.
const (
	GroupRow RowKind = iota
	ProjectRow
	TaskRow
)

// Row is one visible line of the tree.
type Row struct {
	Kind RowKind
	// Group is the kind a group row stands for ("app", "lib", "package").
	Group string
	// Project and Task name what a project or task row stands for.
	Project string
	Task    string
	// Depth is the indentation level.
	Depth int
	// Expandable and Expanded describe the arrow.
	Expandable, Expanded bool
}

// Selection is the row the cursor is on.
type Selection struct {
	Kind    RowKind
	Group   string
	Project string
	Task    string
	// Valid is false for an empty tree.
	Valid bool
}
