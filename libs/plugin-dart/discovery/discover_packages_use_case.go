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

// Discover describes the package behind each pubspec.yaml. The workspace
// root's pubspec only lists the members, so it is not a project.
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	for _, manifest := range manifests {
		root := path.Dir(manifest)
		if root == "." {
			continue
		}
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(root))
		src, err := os.ReadFile(filepath.Join(dir, "pubspec.yaml"))
		if err != nil {
			return result, err
		}
		pub, err := ParsePubspec(src)
		if err != nil {
			return result, fmt.Errorf("%s: %w", manifest, err)
		}
		if len(pub.Workspace) > 0 {
			continue
		}
		kind := projectkind.Resolve(pub.Kind, root, protocol.ProjectLib)
		_, testErr := os.Stat(filepath.Join(dir, "test"))
		result.Projects = append(result.Projects, protocol.Project{
			Name:     path.Base(root),
			Root:     root,
			Language: "dart",
			Kind:     kind,
			Exports:  []string{pub.Name},
			Imports:  pub.Dependencies,
			Version:  pub.Version,
			Tasks:    Tasks(Shape{Root: root, Name: pub.Name, Kind: kind, Flutter: pub.Flutter, Tests: testErr == nil}),
			Stack:    pub.Stack,
		})
	}
	result.SetToolchain(func() string { return toolrun.Toolchain("dart", "dart", "--version") })
	return result, nil
}
