package pipeline

import (
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// probe runs a tool for its version; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Output(dir, argv) }

// CheckPythonVersion compares the venv's Python with the version CI
// installs. A newer one warns, an older one fails: the code may use what the
// venv's Python lacks.
func CheckPythonVersion(workspaceRoot string) protocol.Finding {
	f := protocol.Finding{
		Check: "the venv's Python is " + venv.PythonVersion + ", as in CI", Warning: true,
		Remedy: "recreate .venv with Python " + venv.PythonVersion + " (delete it, then `smith ci setup`)",
	}
	python := strings.ReplaceAll(venv.Interpreter(), "{workspaceRoot}", filepath.ToSlash(workspaceRoot))
	got, err := probe(workspaceRoot, []string{python, "-c", "import sys; print('%d.%d' % sys.version_info[:2])"})
	if err != nil {
		f.Detail = "the venv has no working python"
		return f
	}
	have := strings.TrimSpace(got)
	if found, ok := toolrun.CIPinFinding("Python", have, venv.PythonVersion); ok {
		found.Check, found.Remedy = f.Check, f.Remedy
		return found
	}
	f.Detail = "found Python " + have
	return f
}
