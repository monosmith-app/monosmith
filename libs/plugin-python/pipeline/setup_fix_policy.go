package pipeline

import (
	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// setupFix is the fix of a missing venv: the two commands that create .venv and
// install requirements-dev.txt into it. It does not run `smith ci setup`, which
// also installs every other plugin's dependencies and each Python project.
//
// Parameters: missing is true when the venv is missing; workspaceRoot is where
// the venv goes.
// Returns: the fix, or nil when the venv exists. It is never Safe: it creates
// the venv and installs the requirements from the network.
func setupFix(missing bool, workspaceRoot string) *protocol.Fix {
	if !missing {
		return nil
	}
	create := append(append([]string{}, venv.Launchers()[0]...), "-m", "venv", "--upgrade-deps", ".venv")
	install := []string{venv.Path(workspaceRoot), "-m", "pip", "install", "--quiet", "--disable-pip-version-check", "-r", venv.RequirementsFile}
	return &protocol.Fix{
		Description: "Create the Python venv and install the requirements",
		Steps: []protocol.FixStep{
			{Description: "Create .venv at the workspace root (" + create[0] + " -m venv)", Argv: create},
			{Description: "Install requirements-dev.txt into .venv (uses the network)", Argv: install},
		},
	}
}
