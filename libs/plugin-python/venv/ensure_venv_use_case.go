package venv

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Runner runs a command; tests replace it.
var Runner = func(dir string, argv []string) error { return toolrun.Run(dir, argv) }

// Path expands Interpreter() for a workspace.
func Path(workspaceRoot string) string {
	return filepath.FromSlash(strings.Replace(Interpreter(), "{workspaceRoot}", filepath.ToSlash(workspaceRoot), 1))
}

// Ensure creates .venv and the requirements file when missing and installs the
// pinned development tools. It returns the files it created.
func Ensure(workspaceRoot string) ([]string, error) {
	var created []string
	if added, err := ignoreVenv(workspaceRoot); err != nil {
		return nil, err
	} else if added {
		created = append(created, ".gitignore")
	}
	req := filepath.Join(workspaceRoot, RequirementsFile)
	if _, err := os.Stat(req); errors.Is(err, fs.ErrNotExist) {
		if err := os.WriteFile(req, []byte(Requirements), 0o644); err != nil {
			return nil, err
		}
		created = append(created, RequirementsFile)
	}
	if _, err := os.Stat(Path(workspaceRoot)); errors.Is(err, fs.ErrNotExist) {
		toolrun.Step("Creating the Python environment (.venv)")
		if err := create(workspaceRoot); err != nil {
			return created, err
		}
	}
	py := Path(workspaceRoot)
	toolrun.Step("Installing the Python tools (pytest, ruff, mypy and the rest of " + RequirementsFile + ")")
	if err := Runner(workspaceRoot, []string{py, "-m", "pip", "install", "--quiet", "--disable-pip-version-check", "-r", RequirementsFile}); err != nil {
		return created, err
	}
	return created, nil
}

// InstallEditable installs one project into the venv in editable mode, which
// is what makes every workspace project importable by every other one.
func InstallEditable(workspaceRoot, projectRoot string) error {
	toolrun.Step("Installing " + projectRoot + " into the environment")
	return Runner(workspaceRoot, []string{Path(workspaceRoot), "-m", "pip", "install", "--quiet", "--disable-pip-version-check", "-e", "./" + projectRoot})
}

func create(workspaceRoot string) error {
	var lastErr error
	for _, launcher := range Launchers() {
		if !toolrun.Available(launcher[0]) {
			continue
		}
		argv := append(append([]string{}, launcher...), "-m", "venv", "--upgrade-deps", ".venv")
		if lastErr = Runner(workspaceRoot, argv); lastErr == nil {
			return nil
		}
	}
	if lastErr != nil {
		return lastErr
	}
	return errors.New("no Python found on PATH (tried " + strings.Join(names(), ", ") + "); install Python " + PythonVersion + " or newer")
}

func names() []string {
	var out []string
	for _, l := range Launchers() {
		out = append(out, strings.Join(l, " "))
	}
	return out
}

// Uninstall removes a distribution from the venv.
func Uninstall(workspaceRoot, distribution string) error {
	return Runner(workspaceRoot, []string{Path(workspaceRoot), "-m", "pip", "uninstall", "--yes", "--quiet", "--disable-pip-version-check", distribution})
}
