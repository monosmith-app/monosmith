package discovery

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// DiscoverWithDeclared is Discover plus the projects smith.toml declares: a
// folder of the gem above it (its exe/, a lib/ folder, a folder of specs), with
// no gemspec of its own.
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
	result.SetToolchain(func() string { return toolrun.Toolchain("ruby", "ruby", "--version") })
	return result, nil
}

// describeDeclared describes one declared folder as a project of the gem whose
// gemspec is in a folder above it. Bundler finds the Gemfile above the folder, and
// RuboCop and RSpec run in a folder, so the tasks are those of a gem run there.
// A gem is required by its name, not by a folder, so the project exports
// nothing; it imports what the gemspec depends on. The gemspec is a task input,
// since the folder does not hold it.
func describeDeclared(workspaceRoot string, d protocol.DeclaredProject) (protocol.Project, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(d.Root))
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return protocol.Project{}, fmt.Errorf("the folder does not exist")
	}
	if own, _ := filepath.Glob(filepath.Join(dir, "*.gemspec")); len(own) > 0 {
		return protocol.Project{}, fmt.Errorf("the folder has a gemspec of its own, so it is a gem smith finds itself: remove the declaration")
	}
	gemDir, spec := "", ""
	for up := path.Dir(d.Root); ; up = path.Dir(up) {
		if found, _ := filepath.Glob(filepath.Join(workspaceRoot, filepath.FromSlash(up), "*.gemspec")); len(found) > 0 {
			gemDir, spec = up, filepath.Base(found[0])
			break
		}
		if up == "." || up == "/" {
			return protocol.Project{}, fmt.Errorf("no gemspec in a folder above it: a declared Ruby project is a folder of a gem")
		}
	}
	manifest := path.Join(gemDir, spec)
	src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(manifest)))
	if err != nil {
		return protocol.Project{}, err
	}
	gem := ParseGemspec(string(src))
	kind := d.Kind
	switch kind {
	case "":
		kind = protocol.ProjectLib
	case protocol.ProjectPackage:
		return protocol.Project{}, fmt.Errorf("a folder of a gem cannot be released on its own (a gem has one version): declare it as a lib or an app")
	}
	// An app runs the gem's executable that is in the folder (exe/ holds them).
	start := ""
	for _, e := range gem.Executables {
		if info, err := os.Stat(filepath.Join(dir, e)); err == nil && !info.IsDir() {
			start = e
			break
		}
	}
	if kind == protocol.ProjectApp && start == "" {
		return protocol.Project{}, fmt.Errorf("an app runs an executable of %s, and none is in the folder: declare it as a lib, or point it at the gem's exe/ folder", manifest)
	}
	_, specErr := os.Stat(filepath.Join(dir, "spec"))
	tasks := Tasks(kind, start, specErr == nil)
	for name, t := range tasks {
		if len(t.Inputs) > 0 {
			t.Inputs = append(slices.Clone(t.Inputs), "{workspaceRoot}/"+manifest)
			tasks[name] = t
		}
	}
	return protocol.Project{
		Name:     d.Name,
		Root:     d.Root,
		Language: "ruby",
		Kind:     kind,
		Imports:  gem.Dependencies,
		Tasks:    tasks,
		Stack:    gem.Stack,
	}, nil
}
