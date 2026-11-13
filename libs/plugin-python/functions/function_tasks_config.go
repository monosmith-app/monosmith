package functions

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// WithFunctionTasks gives a function app the tasks of its host in place of the
// plain app's. The app holds no script of smith's: start runs the tool of the
// host (func, sam, the Functions Framework), build runs `smith func stage` to
// stage the deployed files in the workspace's dist/, and package runs
// `smith func zip` to zip them there.
//
// Parameters: tasks are the app's tasks, which it changes; dir is the absolute
// project folder; root the workspace-relative one; python is the interpreter;
// inputs are the files the build reads.
func WithFunctionTasks(tasks map[string]protocol.Task, dir, root, python string, inputs []string) {
	host := DetectHost(dir)
	tasks["start"] = startTask(host, dir, python)
	tasks["build"] = protocol.Task{
		Command: []string{"{smith}", "func", "stage", "--host", host, "--python", python}, Inputs: inputs, Cache: true,
		Outputs:     []string{protocol.BuildDirFromProject(root)},
		Description: "Stage the deployed files into " + protocol.ProjectBuildDir(root) + "/app",
	}
	tasks["package"] = protocol.Task{
		Command: []string{"{smith}", "func", "zip"}, DependsOn: []string{"build"},
		Description: "Zip the staged files into " + protocol.ProjectBuildDir(root) + "/",
	}
}

// startTask runs the function locally with the tool of its host.
func startTask(host, dir, python string) protocol.Task {
	const description = "Run the function locally"
	switch host {
	case functionhost.Azure:
		// Core Tools starts the Python worker with the interpreter it is told.
		return protocol.Task{Command: []string{"func", "start"}, Env: map[string]string{"languageWorkers__python__defaultExecutablePath": python}, Description: description}
	case functionhost.AWS:
		// sam reads the staged code: the build runs first. An HTTP function is a
		// local API, any other trigger is invoked once with the sample event.
		command := []string{"sam", "local", "invoke", "GreetingFunction", "--event", "events/greeting.json"}
		if strings.Contains(readFile(dir, "template.yaml"), "Type: Api") {
			command = []string{"sam", "local", "start-api"}
		}
		return protocol.Task{Command: command, DependsOn: []string{"build"}, Description: description}
	}
	command := []string{python, "-m", "functions_framework"}
	target, cloudEvent := frameworkTarget(dir)
	if target != "" {
		command = append(command, "--target="+target)
	}
	if cloudEvent {
		command = append(command, "--signature-type=cloudevent")
	}
	return protocol.Task{Command: command, Description: description}
}

// frameworkDecorator matches a function the Functions Framework serves: an
// HTTP one, or a CloudEvent one (which needs --signature-type=cloudevent).
var frameworkDecorator = regexp.MustCompile(`@functions_framework\.(http|cloud_event)\s+def\s+(\w+)`)

// frameworkTarget finds the function that the Functions Framework serves in
// the app's functions slice, and whether it answers CloudEvents.
func frameworkTarget(dir string) (target string, cloudEvent bool) {
	files, _ := filepath.Glob(filepath.Join(dir, "src", "*", "functions", "*.py"))
	for _, file := range files {
		if m := frameworkDecorator.FindStringSubmatch(readFile("", file)); m != nil {
			return m[2], m[1] == "cloud_event"
		}
	}
	return "", false
}

// readFile reads a file under dir; an unreadable file is empty.
func readFile(dir, name string) string {
	raw, _ := os.ReadFile(filepath.Join(dir, name))
	return string(raw)
}
