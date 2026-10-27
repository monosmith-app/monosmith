package testrunning

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"
)

// Args returns the arguments pytest takes besides its own fixed ones. When the
// project, or a directory above it up to the workspace root, configures pytest,
// pytest finds that configuration itself and nothing is passed; otherwise
// smith's default is written to .smith/defaults/ and passed with -c, together
// with --rootdir=. so the project, not the default's folder, stays the root.
// The second result says which of the two applies, for the task's description.
func Args(workspaceRoot, projectRoot string) (args []string, own bool, err error) {
	if tooldefaults.FindConfig(workspaceRoot, projectRoot, configures) {
		return nil, true, nil
	}
	rel, err := tooldefaults.WriteDefault(workspaceRoot, DefaultFile, defaultConfig)
	if err != nil {
		return nil, false, err
	}
	return []string{"-c", "{workspaceRoot}/" + rel, "--rootdir=."}, false, nil
}

// configures reports whether pytest would read its configuration from dir.
func configures(dir string) bool {
	if tooldefaults.HasFile("pytest.ini", ".pytest.ini")(dir) {
		return true
	}
	if has(dir, "pyproject.toml", "[tool.pytest") || has(dir, "tox.ini", "[pytest]") {
		return true
	}
	return has(dir, "setup.cfg", "[tool:pytest]")
}

func has(dir, file, marker string) bool {
	raw, err := os.ReadFile(filepath.Join(dir, file))
	return err == nil && strings.Contains(string(raw), marker)
}
