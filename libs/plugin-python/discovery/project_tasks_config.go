package discovery

import (
	"slices"

	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"
)

var inputs = []string{
	"**/*", "{workspaceRoot}/requirements-dev.txt",
	"{workspaceRoot}/ruff.toml", "{workspaceRoot}/.ruff.toml",
	"{workspaceRoot}/mypy.ini", "{workspaceRoot}/.mypy.ini",
	"{workspaceRoot}/pytest.ini", "{workspaceRoot}/.pytest.ini",
	"{workspaceRoot}/pyproject.toml", "{workspaceRoot}/setup.cfg", "{workspaceRoot}/tox.ini",
}

// ToolArgs are the arguments each tool takes from the plugin's defaults slices
// (linting, typechecking, testing): --config and the like with smith's default
// file, or nothing when the repository configures the tool itself.
type ToolArgs struct {
	Ruff, Mypy, Pytest []string
}

// Tasks returns the tasks of a Python project. Every project is installed in
// the venv in editable mode, so none needs another project built first. Lint,
// typecheck and test run on smith's default configuration unless the repository
// configures the tool; the default's content is an input of the task.
func Tasks(root string, kind protocol.ProjectKind, importName string, args ToolArgs) map[string]protocol.Task {
	py := venv.Interpreter()
	task := func(command []string, tool string, toolArgs []string) protocol.Task {
		description := tool + " (the repository's config)"
		if len(toolArgs) > 0 {
			description = tool + " (smith's default config)"
		}
		return protocol.Task{Command: command, Inputs: tooldefaults.WithDefaultInputs(inputs, toolArgs), Cache: true, Description: description}
	}
	tasks := map[string]protocol.Task{
		"lint":      task(slices.Concat([]string{py, "-m", "ruff", "check"}, args.Ruff, []string{"."}), "ruff check", args.Ruff),
		"typecheck": task(slices.Concat([]string{py, "-m", "mypy"}, args.Mypy, []string{"."}), "mypy", args.Mypy),
		"test":      task(slices.Concat([]string{py, "-m", "pytest", "-q", "-p", "no:cacheprovider"}, args.Pytest), "pytest", args.Pytest),
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
