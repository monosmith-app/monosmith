package discovery

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-rust/formatting"
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
		if isAppShell(workspaceRoot, root) {
			continue
		}
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
		fmtArgs, _, err := formatting.Args(workspaceRoot, root)
		if err != nil {
			return result, fmt.Errorf("%s: %w", manifest, err)
		}
		var tags []string
		if crate.Native {
			tags = append(tags, NativeTag)
		}
		result.Projects = append(result.Projects, protocol.Project{
			Name:        path.Base(root),
			Root:        root,
			Language:    "rust",
			Kind:        kind,
			Exports:     []string{crate.Name},
			Imports:     crate.PathDependencies,
			Version:     crate.Version,
			Description: crate.Description,
			Tasks:       Tasks(crate.Name, kind, fmtArgs),
			Stack:       crate.Stack,
			Tags:        tags,
		})
	}
	result.SetToolchain(func() string { return toolrun.Toolchain("rustc", "rustc", "--version") })
	return result, nil
}
