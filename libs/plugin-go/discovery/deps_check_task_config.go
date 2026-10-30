package discovery

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// DepsCheckTask is the name of the task that checks a module requires what it
// imports (smith go-deps-check).
const DepsCheckTask = "deps-check"

// addDepsCheck gives a package the deps-check task, cached on its files and
// the workspace's go.work, and makes its lint run after it, so CI's verify
// phase runs it.
func addDepsCheck(tasks map[string]protocol.Task, root string) {
	tasks[DepsCheckTask] = protocol.Task{
		Command:     []string{"{smith}", "go-deps-check", root},
		Cache:       true,
		Inputs:      inputs,
		Description: "Check go.mod requires every module the package imports",
	}
	for _, name := range []string{"lint", "build", "test"} {
		if t, ok := tasks[name]; ok {
			t.DependsOn = append(t.DependsOn, DepsCheckTask)
			tasks[name] = t
			return
		}
	}
}
