package discovery

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// pythonToolchain is the version of the Python the project's tasks run with:
// the workspace venv's interpreter when there is one, else the first system
// interpreter that answers (py -3 on Windows, python, python3).
func pythonToolchain(workspaceRoot string) string {
	candidates := [][]string{{"py", "-3", "--version"}, {"python", "--version"}, {"python3", "--version"}}
	if python := venvPython(workspaceRoot); python != "" {
		candidates = append([][]string{{python, "--version"}}, candidates...)
	}
	for _, argv := range candidates {
		if v := toolrun.Toolchain("python", argv...); v != "" {
			return v
		}
	}
	return ""
}

// venvPython is the venv's interpreter when it exists, else "".
func venvPython(workspaceRoot string) string {
	python := filepath.FromSlash(strings.ReplaceAll(venv.Interpreter(), "{workspaceRoot}", filepath.ToSlash(workspaceRoot)))
	if _, err := os.Stat(python); err != nil {
		return ""
	}
	return python
}
