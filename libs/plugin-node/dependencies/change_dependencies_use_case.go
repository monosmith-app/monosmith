package dependencies

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// NpmRunner runs npm; tests replace it.
var NpmRunner = func(dir string, argv []string) error { return toolrun.Run(dir, argv) }

// QuietNpmRunner runs npm keeping its output unless it fails; tests replace it.
var QuietNpmRunner = func(dir string, argv []string) error { return toolrun.Quiet(dir, argv) }

var quiet = []string{"--no-audit", "--no-fund"}

// Install adds packages to one project's package.json with npm's workspace
// flag, or installs every workspace when there is no project.
func Install(workspaceRoot string, params protocol.InstallParams) (protocol.DependencyResult, error) {
	argv := []string{"npm", "install"}
	files := []string{"package-lock.json"}
	if params.Project != nil {
		argv = append(append(argv, params.Packages...), "--workspace", params.Project.Root)
		if params.Dev {
			argv = append(argv, "--save-dev")
		}
		if len(params.Packages) > 0 {
			files = append([]string{params.Project.Root + "/package.json"}, files...)
		}
	}
	run := NpmRunner
	if params.Quiet {
		run = QuietNpmRunner
	}
	if err := run(workspaceRoot, append(argv, quiet...)); err != nil {
		return protocol.DependencyResult{}, err
	}
	return protocol.DependencyResult{Files: files}, nil
}

// Uninstall removes packages from one project's package.json.
func Uninstall(workspaceRoot string, params protocol.UninstallParams) (protocol.DependencyResult, error) {
	argv := append(append([]string{"npm", "uninstall"}, params.Packages...), "--workspace", params.Project.Root)
	if err := NpmRunner(workspaceRoot, append(argv, quiet...)); err != nil {
		return protocol.DependencyResult{}, err
	}
	return protocol.DependencyResult{Files: []string{params.Project.Root + "/package.json", "package-lock.json"}}, nil
}
