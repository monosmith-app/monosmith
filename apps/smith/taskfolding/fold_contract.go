package taskfolding

import "github.com/monosmith-app/monosmith/apps/smith/workspace"

// Project is what folding needs to know about one project.
type Project struct {
	Name string
	// Language and Kind are what the selectors lang: and kind: name.
	Language string
	Kind     string
	// Has lists the tasks the project has before its own overrides: a group
	// changes only those.
	Has map[string]bool
	// Overrides are the project's own [projects.<name>.tasks.<task>].
	Overrides map[string]workspace.TaskOverride
}

// Folded says that Projects now take Task from the group selected by Select.
type Folded struct {
	Task     string
	Select   []string
	Projects []string
}

// Result is the folded configuration.
type Result struct {
	// Groups are the task groups to add, in order.
	Groups []workspace.TaskGroup
	// Overrides are every project's own task overrides afterwards; a project
	// with none is absent.
	Overrides map[string]map[string]workspace.TaskOverride
	// Folded lists what moved, by task.
	Folded []Folded
}
