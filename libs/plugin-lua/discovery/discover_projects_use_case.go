package discovery

import (
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Discover describes the project behind each rockspec. A project exports its
// rock's name and imports the rocks its dependencies name; the host keeps the
// ones another workspace project exports. A directory is one project: the
// first rockspec in it counts, and a file without a package is not one. A
// rock is a lib unless its comment or directory says otherwise.
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	seen := map[string]bool{}
	for _, manifest := range manifests {
		root := path.Dir(manifest)
		if root == "." || seen[root] {
			continue
		}
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(root))
		src, err := os.ReadFile(filepath.Join(dir, path.Base(manifest)))
		if err != nil {
			return result, err
		}
		spec := ParseRockspec(string(src))
		if spec.Package == "" {
			continue
		}
		seen[root] = true
		info, statErr := os.Stat(filepath.Join(dir, "spec"))
		result.Projects = append(result.Projects, protocol.Project{
			Name:     spec.Package,
			Root:     root,
			Language: "lua",
			Kind:     projectkind.Resolve(spec.Kind, root, protocol.ProjectLib),
			Exports:  []string{spec.Package},
			Imports:  spec.Dependencies,
			Version:  ReleaseVersion(spec.Version),
			Tasks:    Tasks(root, path.Base(manifest), statErr == nil && info.IsDir()),
			Stack:    spec.Stack,
		})
	}
	return result, nil
}
