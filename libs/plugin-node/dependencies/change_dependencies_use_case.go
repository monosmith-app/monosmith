package dependencies

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// FoldedNpmRunner runs npm and returns its warnings folded into one-line
// notes; its output is shown only when it fails. Tests replace it.
var FoldedNpmRunner = func(dir string, argv []string) ([]string, error) { return toolrun.Folded(dir, argv) }

var quiet = []string{"--no-audit", "--no-fund"}

// Install adds packages to one project's package.json with npm's workspace
// flag, or installs every workspace when there is no project. A project whose
// root is "." is the workspace root: its package.json is edited without
// --workspace, so the packages are shared by every project.
func Install(workspaceRoot string, params protocol.InstallParams) (protocol.DependencyResult, error) {
	argv := []string{"npm", "install"}
	files := []string{"package-lock.json"}
	if params.Project != nil {
		manifest := "package.json"
		argv = append(argv, params.Packages...)
		if params.Project.Root != "." {
			argv = append(argv, "--workspace", params.Project.Root)
			manifest = params.Project.Root + "/package.json"
		}
		if params.Dev {
			argv = append(argv, "--save-dev")
		}
		if len(params.Packages) > 0 {
			files = append([]string{manifest}, files...)
		}
	}
	notes, err := FoldedNpmRunner(workspaceRoot, append(argv, quiet...))
	if err != nil {
		return protocol.DependencyResult{}, err
	}
	return protocol.DependencyResult{Files: files, Notes: notes}, nil
}

// Uninstall removes packages from one project's package.json, whatever section
// declares them (dependencies or devDependencies). A project whose root is "."
// is the workspace root: its package.json is edited without --workspace.
func Uninstall(workspaceRoot string, params protocol.UninstallParams) (protocol.DependencyResult, error) {
	argv := append([]string{"npm", "uninstall"}, params.Packages...)
	manifest := "package.json"
	if params.Project.Root != "." {
		argv = append(argv, "--workspace", params.Project.Root)
		manifest = params.Project.Root + "/package.json"
	}
	notes, err := FoldedNpmRunner(workspaceRoot, append(argv, quiet...))
	if err != nil {
		return protocol.DependencyResult{}, err
	}
	return protocol.DependencyResult{Files: []string{manifest, "package-lock.json"}, Notes: notes}, nil
}
