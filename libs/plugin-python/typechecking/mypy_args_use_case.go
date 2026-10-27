package typechecking

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"
)

// Args returns the arguments `mypy` takes before its target. When the project,
// or a directory above it up to the workspace root, configures mypy, mypy finds
// that configuration itself and nothing is passed; otherwise smith's default is
// written to .smith/defaults/ and passed with --config-file. The second result
// says which of the two applies, for the task's description.
func Args(workspaceRoot, projectRoot string) (args []string, own bool, err error) {
	if tooldefaults.FindConfig(workspaceRoot, projectRoot, configures) {
		return nil, true, nil
	}
	rel, err := tooldefaults.WriteDefault(workspaceRoot, tooldefaults.Mypy.File, defaultConfig)
	if err != nil {
		return nil, false, err
	}
	return []string{"--config-file", "{workspaceRoot}/" + rel}, false, nil
}

// configures reports whether mypy would read its configuration from dir.
func configures(dir string) bool {
	if tooldefaults.HasFile("mypy.ini", ".mypy.ini")(dir) {
		return true
	}
	if has(dir, "pyproject.toml", "[tool.mypy") {
		return true
	}
	return has(dir, "setup.cfg", "[mypy")
}

func has(dir, file, marker string) bool {
	raw, err := os.ReadFile(filepath.Join(dir, file))
	return err == nil && strings.Contains(string(raw), marker)
}
