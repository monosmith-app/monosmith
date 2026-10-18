package discovery

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Discover describes the crate behind each Cargo.toml. A manifest without a
// [package] (a virtual workspace manifest) and one at the workspace root (a
// single-crate repository) are not projects of the monorepo.
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	for _, manifest := range manifests {
		root := path.Dir(manifest)
		if root == "." {
			continue
		}
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(root))
		src, err := os.ReadFile(filepath.Join(dir, "Cargo.toml"))
		if err != nil {
			return result, err
		}
		crate, err := ParseCargo(src)
		if err != nil {
			return result, fmt.Errorf("%s: %w", manifest, err)
		}
		if crate.Name == "" {
			continue
		}
		fallback := protocol.ProjectLib
		if _, err := os.Stat(filepath.Join(dir, "src", "main.rs")); err == nil {
			fallback = protocol.ProjectApp
		}
		kind := projectkind.Resolve(crate.Kind, root, fallback)
		result.Projects = append(result.Projects, protocol.Project{
			Name:     path.Base(root),
			Root:     root,
			Language: "rust",
			Kind:     kind,
			Exports:  []string{crate.Name},
			Imports:  crate.PathDependencies,
			Version:  crate.Version,
			Tasks:    Tasks(crate.Name, kind),
			Stack:    crate.Stack,
		})
	}
	result.SetToolchain(func() string { return toolrun.Toolchain("rustc", "rustc", "--version") })
	return result, nil
}
