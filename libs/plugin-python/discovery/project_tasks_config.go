package discovery

import (
	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var inputs = []string{"**/*", "{workspaceRoot}/requirements-dev.txt", "{workspaceRoot}/ruff.toml", "{workspaceRoot}/.ruff.toml", "{workspaceRoot}/pyproject.toml"}

// Tasks returns the tasks of a Python project. Every project is installed in
// the venv in editable mode, so none needs another project built first. A
// project that configures mypy ([tool.mypy]) is also type-checked.
//
// ruffArgs come from the linting slice: --config with smith's default, or
// nothing when the repository configures ruff itself.
func Tasks(root string, kind protocol.ProjectKind, importName string, mypy bool, ruffArgs []string) map[string]protocol.Task {
	py := venv.Interpreter()
	lint := append([]string{py, "-m", "ruff", "check"}, ruffArgs...)
	lintDescription := "ruff check (the repository's ruff config)"
	if len(ruffArgs) > 0 {
		lintDescription = "ruff check (smith's default config)"
	}
	tasks := map[string]protocol.Task{
		"lint": {Command: append(lint, "."), Inputs: inputs, Cache: true, Description: lintDescription},
		"test": {Command: []string{py, "-m", "pytest", "-q", "-p", "no:cacheprovider"}, Inputs: inputs, Cache: true, Description: "pytest"},
	}
	if mypy {
		tasks["typecheck"] = protocol.Task{Command: []string{py, "-m", "mypy", "."}, Inputs: inputs, Cache: true, Description: "mypy"}
	}
	switch kind {
	case protocol.ProjectPackage:
		// Into the workspace's dist/, like every build: nothing in the source tree.
		tasks["build"] = protocol.Task{Command: []string{py, "-m", "build", "--outdir", "{workspaceRoot}/" + protocol.ProjectBuildDir(root), "."}, Inputs: inputs, Outputs: []string{protocol.BuildDirFromProject(root)}, Cache: true, Description: "Build the wheel and sdist into " + protocol.ProjectBuildDir(root) + "/"}
	case protocol.ProjectApp:
		tasks["start"] = protocol.Task{Command: []string{py, "-m", importName}, Description: "Run the app"}
		tasks["package"] = protocol.Task{
			Command:     []string{py, "-m", "build", "--wheel", "--outdir", "{workspaceRoot}/dist/drop", "."},
			Description: "Build the app's wheel into dist/drop (smith ci pack)",
		}
	}
	return tasks
}
