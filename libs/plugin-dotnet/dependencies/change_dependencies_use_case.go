package dependencies

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Dotnet runs the dotnet command in a directory; tests replace it.
var Dotnet = func(dir string, argv []string) error {
	return toolrun.Run(dir, append([]string{"dotnet"}, argv...))
}

// Install adds packages to one project with `dotnet add package` (name or
// name@version), or, with no packages, restores every project of the plugin
// so lockfiles and the package cache follow the project files.
func Install(workspaceRoot string, params protocol.InstallParams) (protocol.DependencyResult, error) {
	if params.Project == nil || len(params.Packages) == 0 {
		projects := params.Projects
		if params.Project != nil {
			projects = []protocol.Project{*params.Project}
		}
		for _, p := range projects {
			if err := Dotnet(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root)), []string{"restore"}); err != nil {
				return protocol.DependencyResult{}, err
			}
		}
		return protocol.DependencyResult{}, nil
	}
	file, err := projectFile(workspaceRoot, *params.Project)
	if err != nil {
		return protocol.DependencyResult{}, err
	}
	for _, pkg := range params.Packages {
		name, version, _ := strings.Cut(pkg, "@")
		argv := []string{"add", file, "package", name}
		if version != "" {
			argv = append(argv, "--version", version)
		}
		if err := Dotnet(workspaceRoot, argv); err != nil {
			return protocol.DependencyResult{}, err
		}
	}
	return protocol.DependencyResult{Files: []string{file}}, nil
}

// Uninstall removes packages from one project with `dotnet remove package`.
func Uninstall(workspaceRoot string, params protocol.UninstallParams) (protocol.DependencyResult, error) {
	file, err := projectFile(workspaceRoot, params.Project)
	if err != nil {
		return protocol.DependencyResult{}, err
	}
	for _, pkg := range params.Packages {
		if err := Dotnet(workspaceRoot, []string{"remove", file, "package", pkg}); err != nil {
			return protocol.DependencyResult{}, fmt.Errorf("%s: %w (does it reference %s?)", params.Project.Name, err, pkg)
		}
	}
	return protocol.DependencyResult{Files: []string{file}}, nil
}

// ProjectFileOf is the project's own project file, workspace-relative, or an
// error that names the folder when it has none.
func ProjectFileOf(workspaceRoot string, p protocol.Project) (string, error) {
	return projectFile(workspaceRoot, p)
}

// projectFile is the project's own .csproj or .fsproj, workspace-relative.
func projectFile(workspaceRoot string, p protocol.Project) (string, error) {
	for _, f := range ProjectFiles(workspaceRoot, p) {
		if path.Dir(f) == p.Root {
			if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(f))); err == nil {
				return f, nil
			}
		}
	}
	return "", fmt.Errorf("%s has no .csproj or .fsproj in %s", p.Name, p.Root)
}
