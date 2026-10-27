package linting

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"
)

// Args returns the arguments `ruff check` takes before its target. When the
// project, or a directory above it up to the workspace root, configures ruff,
// ruff finds that configuration itself and nothing is passed; otherwise smith's
// default is written to .smith/defaults/ and passed with --config. The second
// result says which of the two applies, for the task's description.
func Args(workspaceRoot, projectRoot string) (args []string, own bool, err error) {
	if tooldefaults.FindConfig(workspaceRoot, projectRoot, configures) {
		return nil, true, nil
	}
	rel, err := tooldefaults.WriteDefault(workspaceRoot, tooldefaults.Ruff.File, defaultConfig)
	if err != nil {
		return nil, false, err
	}
	return []string{"--config", "{workspaceRoot}/" + rel}, false, nil
}

// configures reports whether ruff would read its configuration from dir.
func configures(dir string) bool {
	if tooldefaults.HasFile("ruff.toml", ".ruff.toml")(dir) {
		return true
	}
	raw, err := os.ReadFile(filepath.Join(dir, "pyproject.toml"))
	return err == nil && strings.Contains(string(raw), "[tool.ruff")
}
