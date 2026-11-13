package pipeline

import (
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolpin"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// probe runs a tool for its version; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Probe(dir, argv) }

// CheckPythonVersion compares the venv's Python with the version CI
// installs. A newer one warns, an older one fails: the code may use what the
// venv's Python lacks.
func CheckPythonVersion(workspaceRoot string) protocol.Finding {
	pin := toolpin.Version(workspaceRoot, "python")
	f := protocol.Finding{
		Check: "the venv's Python is " + pin + ", as in CI", Area: "toolchain", Warning: true,
		Remedy: "recreate .venv with Python " + pin + " (delete it, then `smith ci setup`)",
	}
	python := strings.ReplaceAll(venv.Interpreter(), "{workspaceRoot}", filepath.ToSlash(workspaceRoot))
	got, err := probe(workspaceRoot, []string{python, "-c", "import sys; print('%d.%d' % sys.version_info[:2])"})
	if err != nil {
		f.Problem, f.Detail = "the venv has no working Python, CI installs "+pin, "the venv has no working python"
		return f
	}
	have := strings.TrimSpace(got)
	if found, ok := toolrun.CIPinFinding("Python", have, pin); ok {
		found.Check, found.Remedy = f.Check, f.Remedy
		return found
	}
	f.Problem, f.Detail = "the venv's Python is "+have+", CI uses "+pin, "found Python "+have
	return f
}
