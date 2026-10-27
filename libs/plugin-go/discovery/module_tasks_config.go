package discovery

import (
	"runtime"

	"github.com/monosmith-app/monosmith/libs/plugin-go/linting"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"
)

var inputs = []string{"**/*", "{workspaceRoot}/go.work", "{workspaceRoot}/go.work.sum"}

// lintInputs add the repository's golangci-lint configuration to the sources,
// so editing it reruns the lint. A configuration inside the module is already
// one of its files.
var lintInputs = append(append([]string{}, inputs...), "{workspaceRoot}/.golangci.{yml,yaml,toml,json}")

// NativeTag marks a module built on each OS (cgo): smith ci verify leaves it
// to smith ci native, which runs on every runner.
const NativeTag = "native"

// Tasks returns the tasks of a Go module. Go builds its dependencies itself,
// so no task depends on another project's build; the cache key still follows
// every dependency's sources. A native module builds with cgo. Lint is
// golangci-lint: with smith's default configuration unless the repository has
// its own (see the linting slice).
func Tasks(name, root string, kind protocol.ProjectKind, native bool, lint linting.Lint) map[string]protocol.Task {
	tasks := map[string]protocol.Task{
		"test": {Command: []string{"go", "test", "./..."}, Inputs: inputs, Cache: true, Description: "go test ./..."},
		"lint": {Command: lint.Command, Inputs: tooldefaults.WithDefaultInputs(lintInputs, lint.Command), Cache: true, Description: lint.Description},
	}
	if kind == protocol.ProjectApp {
		binary := name
		if runtime.GOOS == "windows" {
			binary += ".exe"
		}
		// The binary goes outside the module, under the workspace's dist/ by
		// the project's path: a directory inside it would be walked by go
		// test ./... and go vet ./..., which fail if it changes under them
		// while the build runs beside them.
		out := protocol.ProjectBuildDir(root)
		tasks["build"] = protocol.Task{Command: []string{"go", "build", "-trimpath", "-o", "{workspaceRoot}/" + out + "/" + binary, "."}, Inputs: inputs, Outputs: []string{protocol.BuildDirFromProject(root)}, Cache: true, Description: "Build the binary into " + out + "/"}
		tasks["start"] = protocol.Task{Command: []string{"go", "run", "."}, Description: "Run the app"}
		tasks["package"] = protocol.Task{
			Command:     []string{"go", "build", "-trimpath", "-o", "{workspaceRoot}/dist/drop/" + name + "/" + binary, "."},
			Description: "Build the binary into dist/drop/" + name + "/ (smith ci pack)",
		}
	} else {
		tasks["build"] = protocol.Task{Command: []string{"go", "build", "./..."}, Inputs: inputs, Cache: true, Description: "Compile every package"}
	}
	if native {
		for n, t := range tasks {
			t.Env = map[string]string{"CGO_ENABLED": "1"}
			tasks[n] = t
		}
	}
	return tasks
}
