package discovery

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/monosmith-app/monosmith/libs/plugin-go/linting"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// DiscoverWithDeclared is Discover plus the projects smith.toml declares: a
// program or a library that is a package of the Go module above its folder,
// with no go.mod of its own. settings are the projects' Go settings by project
// name (DiscoverParams.Settings); their tags go to the tasks.
func DiscoverWithDeclared(workspaceRoot string, manifests []string, declared []protocol.DeclaredProject, settings map[string]map[string]any) (protocol.DiscoverResult, error) {
	result, err := discover(workspaceRoot, manifests, settings)
	if err != nil || len(declared) == 0 {
		return result, err
	}
	for _, d := range declared {
		project, err := describeDeclared(workspaceRoot, d, settings)
		if err != nil {
			return result, fmt.Errorf("[projects.%s] at %s: %w", d.Name, d.Root, err)
		}
		result.Projects = append(result.Projects, project)
	}
	result.SetToolchain(func() string { return toolrun.Toolchain("go", "go", "version") })
	return result, nil
}

// describeDeclared describes one declared folder as a project of the module
// whose go.mod is in it or above it. The module's path and the folder give the
// package path the project exports, so the graph finds the projects that import
// it; the tasks are those of a module project, run in the folder, whose go
// commands work from a folder of the module; the module's go.mod and go.sum are
// task inputs, since the folder does not hold them.
func describeDeclared(workspaceRoot string, d protocol.DeclaredProject, settings map[string]map[string]any) (protocol.Project, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(d.Root))
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return protocol.Project{}, fmt.Errorf("the folder does not exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		return protocol.Project{}, fmt.Errorf("the folder has a go.mod of its own, so it is a module project smith finds itself: remove the declaration")
	}
	moduleDir := ""
	for up := path.Dir(d.Root); ; up = path.Dir(up) {
		if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(up), "go.mod")); err == nil {
			moduleDir = up
			break
		}
		if up == "." || up == "/" {
			return protocol.Project{}, fmt.Errorf("no go.mod in a folder above it: a declared Go project is a package of a module")
		}
	}
	manifest := path.Join(moduleDir, "go.mod")
	src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(manifest)))
	if err != nil {
		return protocol.Project{}, err
	}
	mod, err := ParseModule(string(src))
	if err != nil {
		return protocol.Project{}, fmt.Errorf("%s: %w", manifest, err)
	}
	kind := d.Kind
	switch kind {
	case "":
		kind = fallbackKind(dir)
	case protocol.ProjectPackage:
		return protocol.Project{}, fmt.Errorf("a package of a module cannot be released on its own (a module has one version): declare it as a lib or an app")
	}
	imports, err := ScanImports(dir)
	if err != nil {
		return protocol.Project{}, err
	}
	lint, err := linting.Command(workspaceRoot, d.Root)
	if err != nil {
		return protocol.Project{}, err
	}
	exported := mod.Path
	if rel := relativeTo(moduleDir, d.Root); rel != "" {
		exported += "/" + rel
	}
	tasks := Tasks(d.Name, d.Root, kind, mod.Native, lint)
	tags, err := projectBuildTags(settings, d.Name)
	if err != nil {
		return protocol.Project{}, err
	}
	withBuildTags(tasks, tags)
	for name, t := range tasks {
		if len(t.Inputs) == 0 {
			continue
		}
		t.Inputs = append(slices.Clone(t.Inputs), moduleFile(moduleDir, "go.mod"), moduleFile(moduleDir, "go.sum"))
		tasks[name] = t
	}
	project := protocol.Project{
		Name:     d.Name,
		Root:     d.Root,
		Language: "go",
		Kind:     kind,
		Exports:  []string{exported},
		Imports:  imports,
		Tasks:    tasks,
		Stack:    mod.Stack,
	}
	if mod.Native {
		project.Tags = append(project.Tags, NativeTag)
	}
	return project, nil
}

// relativeTo is the path of target below base ("." is the workspace root).
func relativeTo(base, target string) string {
	if base == "." {
		return target
	}
	return target[len(base)+1:]
}

// moduleFile is a task input naming a file of the module, from the workspace root.
func moduleFile(moduleDir, name string) string {
	if moduleDir == "." {
		return "{workspaceRoot}/" + name
	}
	return "{workspaceRoot}/" + moduleDir + "/" + name
}
