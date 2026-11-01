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
// program of the crate above its folder (src/bin/<name>/main.rs), with no
// Cargo.toml of its own.
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
	result.SetToolchain(func() string { return toolrun.Toolchain("rustc", "rustc", "--version") })
	return result, nil
}

// describeDeclared describes one declared folder as a binary target of the crate
// whose Cargo.toml is above it. Cargo runs a binary of a crate with --bin, from
// any folder of the crate, so the tasks are the crate's scoped to that binary.
// A binary has no name other crates import it by, so the project exports
// nothing; it imports the crates the manifest above depends on by path.
func describeDeclared(workspaceRoot string, d protocol.DeclaredProject) (protocol.Project, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(d.Root))
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return protocol.Project{}, fmt.Errorf("the folder does not exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "Cargo.toml")); err == nil {
		return protocol.Project{}, fmt.Errorf("the folder has a Cargo.toml of its own, so it is a crate smith finds itself: remove the declaration")
	}
	crateDir := ""
	var crate Crate
	for up := path.Dir(d.Root); ; up = path.Dir(up) {
		if src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(up), "Cargo.toml")); err == nil {
			parsed, perr := ParseCargo(src)
			if perr != nil {
				return protocol.Project{}, fmt.Errorf("%s: %w", path.Join(up, "Cargo.toml"), perr)
			}
			if parsed.Name != "" {
				crateDir, crate = up, parsed
				break
			}
		}
		if up == "." || up == "/" {
			return protocol.Project{}, fmt.Errorf("no Cargo.toml with a [package] in a folder above it: a declared Rust project is a program of a crate")
		}
	}
	if info, err := os.Stat(filepath.Join(dir, "main.rs")); path.Dir(d.Root) != path.Join(crateDir, "src", "bin") || err != nil || info.IsDir() {
		return protocol.Project{}, fmt.Errorf("cargo builds a folder only as a program of the crate, in %s/src/bin/<name>/main.rs: move it there, or give the folder a Cargo.toml", crateDir)
	}
	kind := d.Kind
	switch kind {
	case "":
		kind = protocol.ProjectApp
	case protocol.ProjectPackage:
		return protocol.Project{}, fmt.Errorf("a program of a crate cannot be released on its own (a crate has one version): declare it as a lib or an app")
	}
	tasks := binTasks(crate.Name, path.Base(d.Root), kind)
	crateInputs := []string{"{workspaceRoot}/" + path.Join(crateDir, "Cargo.toml"), "{workspaceRoot}/" + path.Join(crateDir, "src", "**", "*")}
	for name, t := range tasks {
		if len(t.Inputs) > 0 {
			t.Inputs = slices.Concat(t.Inputs, crateInputs)
			tasks[name] = t
		}
	}
	return protocol.Project{
		Name:     d.Name,
		Root:     d.Root,
		Language: "rust",
		Kind:     kind,
		Imports:  crate.PathDependencies,
		Tasks:    tasks,
		Stack:    crate.Stack,
	}, nil
}

// binTasks are cargo's, scoped to one binary of a crate with -p and --bin: build,
// test, lint (clippy), typecheck and, for an app, start. There is no format task:
// cargo fmt covers a whole crate, not one binary.
func binTasks(crate, bin string, kind protocol.ProjectKind) map[string]protocol.Task {
	scoped := func(args ...string) []string {
		return slices.Concat([]string{"cargo"}, args, []string{"-p", crate, "--bin", bin})
	}
	tasks := map[string]protocol.Task{
		"build":     {Command: scoped("build"), Inputs: inputs, Description: "cargo build --bin " + bin},
		"test":      {Command: scoped("test"), Inputs: inputs, Cache: true, Description: "cargo test --bin " + bin},
		"lint":      {Command: append(scoped("clippy"), "--", "-D", "warnings"), Inputs: inputs, Cache: true, Description: "cargo clippy, warnings are errors"},
		"typecheck": {Command: scoped("check"), Inputs: inputs, Cache: true, Description: "cargo check --bin " + bin},
	}
	if kind == protocol.ProjectApp {
		tasks["start"] = protocol.Task{Command: scoped("run"), Description: "Run the app"}
	}
	return tasks
}
