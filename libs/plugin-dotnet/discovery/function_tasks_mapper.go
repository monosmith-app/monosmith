package discovery

import (
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/functions"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// tasksOf are the tasks of a project: a function app (ADR 0072) packages and
// starts the way its host needs, in place of the plain app's tasks.
func tasksOf(workspaceRoot, root, file, tests string, kind protocol.ProjectKind) map[string]protocol.Task {
	tasks := Tasks(root, file, tests, kind)
	if kind != protocol.ProjectApp {
		return tasks
	}
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(root))
	if host := functions.DetectHost(dir, functions.ProjectFile(dir)); host != "" {
		functions.WithFunctionTasks(tasks, host, root, file, dir, inputs)
	}
	return tasks
}
