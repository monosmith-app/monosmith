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

// entryPoint is the file a declared Dart app runs, in its folder.
const entryPoint = "main.dart"

// DiscoverWithDeclared is Discover plus the projects smith.toml declares: a
// folder of the package above it (a bin/ program, a folder of lib/), with no
// pubspec.yaml of its own.
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
	result.SetToolchain(func() string { return toolrun.Toolchain("dart", "dart", "--version") })
	return result, nil
}

// describeDeclared describes one declared folder as a project of the package whose
// pubspec.yaml is above it (a workspace root, which only lists its members, is
// not a package). dart analyze, format, test and compile work from a folder of a
// package, which they resolve through the pubspec above, so the tasks run in the
// folder. Dart packages are imported by name, not by folder, so the project
// exports nothing; it imports what the pubspec depends on. The pubspec is a task
// input, since the folder does not hold it. A Flutter package is refused: its
// tasks need the package root.
func describeDeclared(workspaceRoot string, d protocol.DeclaredProject) (protocol.Project, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(d.Root))
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return protocol.Project{}, fmt.Errorf("the folder does not exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "pubspec.yaml")); err == nil {
		return protocol.Project{}, fmt.Errorf("the folder has a pubspec.yaml of its own, so it is a package smith finds itself: remove the declaration")
	}
	var pub Pubspec
	manifest := ""
	for up := path.Dir(d.Root); ; up = path.Dir(up) {
		rel := path.Join(up, "pubspec.yaml")
		if src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(rel))); err == nil {
			parsed, perr := ParsePubspec(src)
			if perr != nil {
				return protocol.Project{}, fmt.Errorf("%s: %w", rel, perr)
			}
			if len(parsed.Workspace) == 0 {
				pub, manifest = parsed, rel
				break
			}
		}
		if up == "." || up == "/" {
			return protocol.Project{}, fmt.Errorf("no pubspec.yaml of a package in a folder above it: a declared Dart project is a folder of a package")
		}
	}
	if pub.Flutter {
		return protocol.Project{}, fmt.Errorf("%s is a Flutter package, and flutter runs from a package's own folder: give the folder a pubspec.yaml of its own", manifest)
	}
	_, mainErr := os.Stat(filepath.Join(dir, entryPoint))
	kind := d.Kind
	switch kind {
	case "":
		kind = protocol.ProjectLib
		if mainErr == nil {
			kind = protocol.ProjectApp
		}
	case protocol.ProjectPackage:
		return protocol.Project{}, fmt.Errorf("a folder of a package cannot be released on its own (a package has one version): declare it as a lib or an app")
	}
	if kind == protocol.ProjectApp && mainErr != nil {
		return protocol.Project{}, fmt.Errorf("an app runs %s, and the folder has none: declare it as a lib, or add the file", entryPoint)
	}
	_, testErr := os.Stat(filepath.Join(dir, "test"))
	tasks := declaredTasks(d.Root, kind, testErr == nil)
	for name, t := range tasks {
		if len(t.Inputs) > 0 {
			t.Inputs = append(slices.Clone(t.Inputs), "{workspaceRoot}/"+manifest)
			tasks[name] = t
		}
	}
	return protocol.Project{
		Name:     d.Name,
		Root:     d.Root,
		Language: "dart",
		Kind:     kind,
		Imports:  pub.Dependencies,
		Tasks:    tasks,
		Stack:    pub.Stack,
	}, nil
}

// declaredTasks are dart's for a folder of a package, run in the folder: lint,
// format, test (when the folder has a test/ directory) and, for an app, build
// (a native executable in dist/) and start.
func declaredTasks(root string, kind protocol.ProjectKind, tests bool) map[string]protocol.Task {
	sources := []string{"**/*", "!.dart_tool/**", "!build/**", "{workspaceRoot}/pubspec.lock"}
	tasks := map[string]protocol.Task{
		"lint":   {Command: []string{"dart", "analyze", "--fatal-infos"}, Inputs: sources, Cache: true, Description: "dart analyze --fatal-infos"},
		"format": {Command: []string{"dart", "format", "--output=none", "--set-exit-if-changed", "."}, Inputs: sources, Cache: true, Description: "dart format, failing on unformatted files"},
	}
	if tests {
		tasks["test"] = protocol.Task{Command: []string{"dart", "test", "test"}, Inputs: sources, Cache: true, Description: "dart test"}
	}
	if kind == protocol.ProjectApp {
		out := protocol.ProjectBuildDir(root)
		tasks["build"] = protocol.Task{Command: []string{"dart", "compile", "exe", entryPoint, "-o", "{workspaceRoot}/" + out + "/" + path.Base(root)}, Inputs: sources, Outputs: []string{protocol.BuildDirFromProject(root)}, Cache: true, Description: "dart compile exe into " + out + "/"}
		tasks["start"] = protocol.Task{Command: []string{"dart", "run", entryPoint}, Description: "Run the app"}
	}
	return tasks
}
