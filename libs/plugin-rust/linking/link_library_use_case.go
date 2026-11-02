package linking

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-rust/dependencies"
	"github.com/monosmith-app/monosmith/libs/plugin-rust/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// PathTo is the folder of the library as seen from the project's folder, with
// forward slashes and a leading ./ or ../, the way a path dependency spells it.
func PathTo(project, library string) string {
	rel, err := filepath.Rel(filepath.FromSlash(project), filepath.FromSlash(library))
	if err != nil {
		return "./" + library
	}
	rel = filepath.ToSlash(rel)
	if !strings.HasPrefix(rel, ".") {
		rel = "./" + rel
	}
	return rel
}

// AddArgs is the cargo command that adds the library as a path dependency of
// the project: cargo writes the key under [dependencies] (with the version,
// when the library has one a registry could resolve) and updates Cargo.lock.
func AddArgs(project, library string) []string {
	return []string{"add", "--path", PathTo(project, library), "--manifest-path", project + "/Cargo.toml"}
}

// RemoveArgs is the cargo command that drops the dependency on a crate.
func RemoveArgs(project, crate string) []string {
	return []string{"remove", "--manifest-path", project + "/Cargo.toml", crate}
}

func crateName(library protocol.Project) (string, error) {
	if len(library.Exports) == 0 {
		return "", fmt.Errorf("%s/Cargo.toml has no [package] name, so it cannot be depended upon", library.Root)
	}
	return library.Exports[0], nil
}

// declares reports whether the project's Cargo.toml has a path dependency on
// the crate.
func declares(workspaceRoot string, project protocol.Project, crate string) (bool, error) {
	manifest := project.Root + "/Cargo.toml"
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(manifest)))
	if err != nil {
		return false, fmt.Errorf("%s: %w", manifest, err)
	}
	c, err := discovery.ParseCargo(raw)
	if err != nil {
		return false, fmt.Errorf("%s: %w", manifest, err)
	}
	return slices.Contains(c.PathDependencies, crate), nil
}

// Link makes the project depend on the library with a path dependency in its
// Cargo.toml, written by cargo add, which also brings Cargo.lock up to date.
// With params.Plan nothing is written.
func Link(workspaceRoot string, params protocol.LinkParams) (protocol.LinkResult, error) {
	crate, err := crateName(params.Library)
	if err != nil {
		return protocol.LinkResult{}, err
	}
	manifest := params.Project.Root + "/Cargo.toml"
	already, err := declares(workspaceRoot, params.Project, crate)
	if err != nil {
		return protocol.LinkResult{}, err
	}
	res := protocol.LinkResult{Manifest: manifest, Line: fmt.Sprintf("%s = { path = %q }", crate, PathTo(params.Project.Root, params.Library.Root))}
	if params.Plan {
		return res, nil
	}
	if already {
		res.Notes = []string{params.Project.Name + " already depends on " + crate}
		return res, nil
	}
	if err := dependencies.Cargo(workspaceRoot, AddArgs(params.Project.Root, params.Library.Root)); err != nil {
		return protocol.LinkResult{}, err
	}
	res.Files = []string{manifest, "Cargo.lock"}
	return res, nil
}

// Unlink removes the project's path dependency on the library with cargo
// remove. When the Cargo.toml did not have it, the result has no files.
func Unlink(workspaceRoot string, params protocol.UnlinkParams) (protocol.DependencyResult, error) {
	crate, err := crateName(params.Library)
	if err != nil {
		return protocol.DependencyResult{}, err
	}
	has, err := declares(workspaceRoot, params.Project, crate)
	if err != nil || !has {
		return protocol.DependencyResult{}, err
	}
	if err := dependencies.Cargo(workspaceRoot, RemoveArgs(params.Project.Root, crate)); err != nil {
		return protocol.DependencyResult{}, err
	}
	return protocol.DependencyResult{Files: []string{params.Project.Root + "/Cargo.toml", "Cargo.lock"}}, nil
}
