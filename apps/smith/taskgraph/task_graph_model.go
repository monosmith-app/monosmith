package taskgraph

import (
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Task is one runnable unit: a project's task.
type Task struct {
	// ID is "project:task".
	ID      string
	Project *projectgraph.Project
	Name    string
	Def     protocol.Task
	// Deps are IDs of tasks that must succeed first, sorted.
	Deps []string
	// Requested is true for tasks the user asked for, false for tasks pulled
	// in as dependencies.
	Requested bool
	// Fixing is true after UseFix for a lint task that runs its fix.
	Fixing bool
}

// Graph is a set of tasks in an order that respects dependencies.
type Graph struct {
	Tasks map[string]*Task
	// Order lists task IDs so each comes after its dependencies.
	Order []string
}

// UseRelease makes every task of the graph run its release profile
// (protocol.Task.ForRelease): what `smith run --release` runs. A task without
// one stays as it is. The cache key holds the command, so a task's two
// profiles never share a cache entry.
func (g *Graph) UseRelease() {
	for _, t := range g.Tasks {
		t.Def = t.Def.ForRelease()
	}
}

// LintTask is the name of the task `smith lint --fix` fixes.
const LintTask = "lint"

// UseFix prepares the graph for `smith lint --fix`: no lint task reads or
// stores a cached result, since the fix changes the files, and a lint task
// whose plugin declares a fix (protocol.Task.Fix) is marked Fixing, so the run
// fixes in place of linting. A lint task without a fix runs as it is.
func (g *Graph) UseFix() {
	for _, t := range g.Tasks {
		if t.Name != LintTask {
			continue
		}
		t.Def.Cache = false
		t.Fixing = t.Def.Fix != nil && len(t.Def.Fix.Command) > 0
	}
}

// ID builds a task id.
func ID(project, task string) string { return project + ":" + task }
