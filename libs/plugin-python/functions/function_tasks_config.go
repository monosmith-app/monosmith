package functions

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// WithFunctionTasks gives a function app the tasks of its host in place of the
// plain app's: start runs the function locally (start.py), build stages the
// deployed files in the workspace's dist/, and package zips them there.
//
// Parameters: tasks are the app's tasks, which it changes; root is the project
// folder; python is the interpreter; inputs are the files the build reads.
func WithFunctionTasks(tasks map[string]protocol.Task, root, python string, inputs []string) {
	tasks["start"] = protocol.Task{Command: []string{python, "start.py"}, Description: "Run the function locally"}
	tasks["build"] = protocol.Task{
		Command: []string{python, "package.py", "build"}, Inputs: inputs, Cache: true,
		Outputs:     []string{protocol.BuildDirFromProject(root)},
		Description: "Stage the deployed files into " + protocol.ProjectBuildDir(root) + "/app",
	}
	tasks["package"] = protocol.Task{
		Command: []string{python, "package.py", "zip"}, DependsOn: []string{"build"},
		Description: "Zip the staged files into " + protocol.ProjectBuildDir(root) + "/",
	}
}
