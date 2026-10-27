package dependencies

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Cargo runs cargo in a directory; tests replace it.
var Cargo = func(dir string, argv []string) error {
	return toolrun.Run(dir, append([]string{"cargo"}, argv...))
}

// Install adds crates to one project with `cargo add` (a dev-dependency when
// Dev), or, with no project, fetches every crate the workspace's lockfile
// needs, which also brings Cargo.lock up to date after a manifest changed.
func Install(workspaceRoot string, params protocol.InstallParams) (protocol.DependencyResult, error) {
	if params.Project == nil {
		return protocol.DependencyResult{}, Cargo(workspaceRoot, []string{"fetch"})
	}
	p := *params.Project
	if len(params.Packages) == 0 {
		return protocol.DependencyResult{}, Cargo(workspaceRoot, []string{"fetch"})
	}
	argv := []string{"add", "--manifest-path", p.Root + "/Cargo.toml"}
	if params.Dev {
		argv = append(argv, "--dev")
	}
	if err := Cargo(workspaceRoot, append(argv, params.Packages...)); err != nil {
		return protocol.DependencyResult{}, err
	}
	return protocol.DependencyResult{Files: []string{p.Root + "/Cargo.toml", "Cargo.lock"}}, nil
}

// Uninstall removes crates from one project with `cargo remove`.
func Uninstall(workspaceRoot string, params protocol.UninstallParams) (protocol.DependencyResult, error) {
	argv := append([]string{"remove", "--manifest-path", params.Project.Root + "/Cargo.toml"}, params.Packages...)
	if err := Cargo(workspaceRoot, argv); err != nil {
		return protocol.DependencyResult{}, fmt.Errorf("%s: %w (does it depend on %s?)", params.Project.Name, err, strings.Join(params.Packages, ", "))
	}
	return protocol.DependencyResult{Files: []string{params.Project.Root + "/Cargo.toml", "Cargo.lock"}}, nil
}
