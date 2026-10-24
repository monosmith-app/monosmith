package projectlist

// Task is one task of a project, with what the detail shows of it.
type Task struct {
	Name        string
	Command     string
	DependsOn   []string
	Cache       bool
	Description string
}

// Project is what the Projects list and its detail know of a project.
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
