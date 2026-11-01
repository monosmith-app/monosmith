package discovery

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/monosmith-app/monosmith/libs/plugin-php/linting"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// DiscoverWithDeclared is Discover plus the projects smith.toml declares: a
// folder of the package above it (a src/ folder, a bin/ program), with no
// composer.json of its own.
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
	result.SetToolchain(func() string { return toolrun.Toolchain("php", "php", "--version") })
	return result, nil
}

// describeDeclared describes one declared folder as a project of the package
// whose composer.json is in a folder above it. PHPStan and PHPUnit take paths,
// and the workspace's vendor/ and autoloader serve every folder, so the tasks are
// those of a package run in the folder over the folder. A package is required by
// its name, not by a folder, so the project exports nothing; it imports what the
// composer.json requires. The composer.json is a task input, since the folder
// does not hold it.
func describeDeclared(workspaceRoot string, d protocol.DeclaredProject) (protocol.Project, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(d.Root))
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return protocol.Project{}, fmt.Errorf("the folder does not exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "composer.json")); err == nil {
		return protocol.Project{}, fmt.Errorf("the folder has a composer.json of its own, so it is a package smith finds itself: remove the declaration")
	}
	packageDir := ""
	for up := path.Dir(d.Root); ; up = path.Dir(up) {
		if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(up), "composer.json")); err == nil && up != "." {
			packageDir = up
			break
		}
		if up == "." || up == "/" {
			return protocol.Project{}, fmt.Errorf("no composer.json in a folder above it, other than the workspace's: a declared PHP project is a folder of a package")
		}
	}
	manifest := path.Join(packageDir, "composer.json")
	src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(manifest)))
	if err != nil {
		return protocol.Project{}, err
	}
	m, err := ParseManifest(src)
	if err != nil {
		return protocol.Project{}, fmt.Errorf("%s: %w", manifest, err)
	}
	kind := d.Kind
	switch kind {
	case "":
		kind = protocol.ProjectLib
	case protocol.ProjectPackage:
		return protocol.Project{}, fmt.Errorf("a folder of a package cannot be released on its own (a package has one version): declare it as a lib or an app")
	}
	// An app runs the package's executable that is in the folder.
	bin := ""
	for _, b := range m.Bin {
		if rel, ok := relativeBelow(packageDir, d.Root, b); ok {
			bin = rel
			break
		}
	}
	if kind == protocol.ProjectApp && bin == "" {
		return protocol.Project{}, fmt.Errorf("an app runs an executable of %s (\"bin\"), and none is in the folder: declare it as a lib, or point it at the folder of its bin", manifest)
	}
	phpstan, _, err := linting.Args(workspaceRoot, d.Root)
	if err != nil {
		return protocol.Project{}, err
	}
	_, testsErr := os.Stat(filepath.Join(dir, "tests"))
	tasks := Tasks(Shape{Kind: kind, Tests: testsErr == nil, Bin: bin, PHPStan: phpstan, Sources: []string{"."}})
	for name, t := range tasks {
		if len(t.Inputs) > 0 {
			t.Inputs = append(slices.Clone(t.Inputs), "{workspaceRoot}/"+manifest)
			tasks[name] = t
		}
	}
	return protocol.Project{
		Name:     d.Name,
		Root:     d.Root,
		Language: "php",
		Kind:     kind,
		Imports:  m.Requires,
		Tasks:    tasks,
		Stack:    m.Stack,
	}, nil
}

// relativeBelow is file, a path relative to packageDir, as a path relative to
// folder (also a path from the workspace root), and false when file is not in
// folder.
func relativeBelow(packageDir, folder, file string) (string, bool) {
	full := path.Join(packageDir, file)
	if len(full) <= len(folder)+1 || full[:len(folder)+1] != folder+"/" {
		return "", false
	}
	return full[len(folder)+1:], true
}
