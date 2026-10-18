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

// Discover describes the package behind each composer.json. The root's
// only gathers the packages, so it is not a project.
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	for _, manifest := range manifests {
		root := path.Dir(manifest)
		if root == "." {
			continue
		}
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(root))
		src, err := os.ReadFile(filepath.Join(dir, "composer.json"))
		if err != nil {
			return result, err
		}
		m, err := ParseManifest(src)
		if err != nil {
			return result, fmt.Errorf("%s: %w", manifest, err)
		}
		if m.Name == "" {
			m.Name = path.Base(root)
		}
		kind := projectkind.Resolve(m.Kind, root, protocol.ProjectLib)
		_, testsErr := os.Stat(filepath.Join(dir, "tests"))
		bin := ""
		if len(m.Bin) > 0 {
			bin = m.Bin[0]
		}
		result.Projects = append(result.Projects, protocol.Project{
			Name:     path.Base(root),
			Root:     root,
			Language: "php",
			Kind:     kind,
			Exports:  []string{m.Name},
			Imports:  m.Requires,
			Version:  m.Version,
			Tasks:    Tasks(Shape{Kind: kind, Tests: testsErr == nil, Bin: bin}),
			Stack:    m.Stack,
		})
	}
	result.SetToolchain(func() string { return toolrun.Toolchain("php", "php", "--version") })
	return result, nil
}
