package dependencies

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const (
	runtimeTable = "project"
	runtimeKey   = "dependencies"
	devTable     = "project.optional-dependencies"
	devKey       = "dev"
)

// Install records requirements in one project's pyproject.toml (the dev
// extra when Dev) and installs them into the workspace venv. Without a
// project it sets up the venv and installs every project into it.
func Install(workspaceRoot string, params protocol.InstallParams) (protocol.DependencyResult, error) {
	created, err := venv.Ensure(workspaceRoot)
	if err != nil {
		return protocol.DependencyResult{}, err
	}
	res := protocol.DependencyResult{Files: created}
	if params.Project == nil {
		for _, p := range venv.DependencyOrder(params.Projects) {
			if err := venv.InstallEditable(workspaceRoot, p.Root); err != nil {
				return res, err
			}
		}
		return res, nil
	}
	p := *params.Project
	if len(params.Packages) > 0 {
		table, key := runtimeTable, runtimeKey
		if params.Dev {
			table, key = devTable, devKey
		}
		if err := editPyproject(workspaceRoot, p, func(src string) (string, error) {
			return AddRequirements(src, table, key, params.Packages), nil
		}); err != nil {
			return res, err
		}
		res.Files = append(res.Files, p.Root+"/pyproject.toml")
	}
	// The editable install pulls the runtime requirements; dev ones are an
	// extra, installed by name.
	if err := venv.InstallEditable(workspaceRoot, p.Root); err != nil {
		return res, err
	}
	if params.Dev && len(params.Packages) > 0 {
		if err := venv.Runner(workspaceRoot, append([]string{venv.Path(workspaceRoot), "-m", "pip", "install", "--quiet", "--disable-pip-version-check"}, params.Packages...)); err != nil {
			return res, err
		}
	}
	return res, nil
}

// Uninstall drops requirements from one project's pyproject.toml, runtime
// and dev alike. They stay installed in the shared venv, where another
// project may still need them.
func Uninstall(workspaceRoot string, params protocol.UninstallParams) (protocol.DependencyResult, error) {
	var removed []string
	err := editPyproject(workspaceRoot, params.Project, func(src string) (string, error) {
		out, fromRuntime := RemoveRequirements(src, runtimeTable, runtimeKey, params.Packages)
		out, fromDev := RemoveRequirements(out, devTable, devKey, params.Packages)
		removed = append(fromRuntime, fromDev...)
		if len(removed) == 0 {
			return "", fmt.Errorf("%s does not require %s", params.Project.Name, strings.Join(params.Packages, ", "))
		}
		return out, nil
	})
	if err != nil {
		return protocol.DependencyResult{}, err
	}
	return protocol.DependencyResult{
		Files: []string{params.Project.Root + "/pyproject.toml"},
		Notes: []string{"left installed in .venv, which every project shares; recreate it to drop " + strings.Join(removed, ", ")},
	}, nil
}

func editPyproject(workspaceRoot string, p protocol.Project, change func(string) (string, error)) error {
	path := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "pyproject.toml")
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out, err := change(string(raw))
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(out), 0o644)
}
