package discovery

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/monosmith-app/monosmith/libs/plugin-python/linting"
	"github.com/monosmith-app/monosmith/libs/plugin-python/testrunning"
	"github.com/monosmith-app/monosmith/libs/plugin-python/typechecking"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/pelletier/go-toml/v2"
)

// DiscoverWithDeclared is Discover plus the projects smith.toml declares: a
// lib or an app that is a folder of the distribution above it, with no
// pyproject.toml of its own. A declaration is refused with an error that names
// it when its folder is missing, has a pyproject.toml, has none with a
// [project] table above it, or asks for the kind package.
func DiscoverWithDeclared(workspaceRoot string, manifests []string, declared []protocol.DeclaredProject) (protocol.DiscoverResult, error) {
	result, err := Discover(workspaceRoot, manifests)
	if err != nil || len(declared) == 0 {
		return result, err
	}
	for _, d := range declared {
		project, err := describeDeclared(workspaceRoot, d)
		if err != nil {
			return result, fmt.Errorf("[projects.%s] at %s: %w", d.Name, d.Root, err)
		}
		result.Projects = append(result.Projects, project)
	}
	return result, nil
}

// describeDeclared describes one declared folder as a project of the
// distribution whose pyproject.toml is above it. The folder's import name is
// its export; ruff, mypy and pytest run in the folder, and the manifest above
// is an input of every task, since the folder does not hold it.
func describeDeclared(workspaceRoot string, d protocol.DeclaredProject) (protocol.Project, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(d.Root))
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return protocol.Project{}, fmt.Errorf("the folder does not exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "pyproject.toml")); err == nil {
		return protocol.Project{}, fmt.Errorf("the folder has a pyproject.toml of its own, so it is a project smith finds itself: remove the declaration")
	}
	manifestDir := ""
	for up := path.Dir(d.Root); ; up = path.Dir(up) {
		if hasProjectTable(filepath.Join(workspaceRoot, filepath.FromSlash(up), "pyproject.toml")) {
			manifestDir = up
			break
		}
		if up == "." {
			break
		}
	}
	if manifestDir == "" {
		return protocol.Project{}, fmt.Errorf("no pyproject.toml with a [project] table is above the folder: add one, or give the folder a pyproject.toml of its own")
	}
	kind := d.Kind
	if kind == protocol.ProjectPackage {
		return protocol.Project{}, fmt.Errorf("a folder of a distribution cannot be released on its own (a pyproject.toml has one version): declare it as a lib or an app")
	}
	if kind == "" {
		kind = projectkind.Resolve("", d.Root, protocol.ProjectLib)
		if kind == protocol.ProjectPackage {
			kind = protocol.ProjectLib
		}
	}
	imports, err := ScanImports(dir)
	if err != nil {
		return protocol.Project{}, err
	}
	var args ToolArgs
	if args.Ruff, _, err = linting.Args(workspaceRoot, d.Root); err == nil {
		if args.Mypy, _, err = typechecking.Args(workspaceRoot, d.Root); err == nil {
			args.Pytest, _, err = testrunning.Args(workspaceRoot, d.Root)
		}
	}
	if err != nil {
		return protocol.Project{}, err
	}
	importName := ImportName(path.Base(d.Root))
	manifest := "{workspaceRoot}/pyproject.toml"
	if manifestDir != "." {
		manifest = "{workspaceRoot}/" + manifestDir + "/pyproject.toml"
	}
	tasks := Tasks(d.Root, kind, importName, args)
	for name, t := range tasks {
		if len(t.Inputs) > 0 && !slices.Contains(t.Inputs, manifest) {
			t.Inputs = append(slices.Clone(t.Inputs), manifest)
			tasks[name] = t
		}
	}
	return protocol.Project{
		Name:     d.Name,
		Root:     d.Root,
		Language: "python",
		Kind:     kind,
		Exports:  []string{importName},
		Imports:  imports,
		Tasks:    tasks,
	}, nil
}

// hasProjectTable reports whether file is a pyproject.toml with a [project]
// table that names a distribution.
func hasProjectTable(file string) bool {
	raw, err := os.ReadFile(file)
	if err != nil {
		return false
	}
	var py PyProject
	return toml.Unmarshal(raw, &py) == nil && py.Project.Name != ""
}
