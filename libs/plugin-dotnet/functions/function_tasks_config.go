package functions

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// WithFunctionTasks gives a function app the tasks of its host: package
// publishes the app into the workspace's dist/dotnet (the artifacts layout of
// Directory.Build.props), and start runs it locally. Azure and AWS run the
// published folder with their tool; Google runs the app, which is its own host.
//
// Parameters: tasks are the app's tasks, which it changes; host is the detected
// host; root is the project folder, file its project file; dir its absolute
// folder, to read the template of an AWS function.
func WithFunctionTasks(tasks map[string]protocol.Task, host, root, file, dir string, inputs []string) {
	id := strings.TrimSuffix(file, ".csproj")
	published := PublishFolder(id)
	tasks["package"] = protocol.Task{
		Command: []string{"dotnet", "publish", file, "-c", "Release", "--nologo"}, Inputs: inputs,
		DependsOn: []string{"^build"}, Cache: true,
		Outputs:     []string{strings.Repeat("../", strings.Count(root, "/")+1) + published},
		Description: "dotnet publish into " + published + "/",
	}
	switch host {
	case functionhost.Azure:
		tasks["start"] = protocol.Task{Command: []string{"func", "start"}, Cwd: published, DependsOn: []string{"package"}, Description: "Run the function locally with Azure Functions Core Tools"}
	case functionhost.AWS:
		command := []string{"sam", "local", "invoke", "GreetingFunction", "--event", "Events/greeting.json"}
		if template, _ := os.ReadFile(filepath.Join(dir, "template.yaml")); strings.Contains(string(template), "Type: Api") {
			command = []string{"sam", "local", "start-api"}
		}
		tasks["start"] = protocol.Task{Command: command, DependsOn: []string{"package"}, Description: "Run the function locally with the SAM CLI"}
	}
}
