package discovery

import (
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Discover describes the project of each directory that holds exported VBA
// modules. It is named for the directory, exports it and imports nothing: a
// module's references are to the Office host's libraries, not to files. The
// main module carries the kind, stack and version in comments. A directory
// whose files are not VBA modules (a LaTeX .cls, say) is left out, and so are
// modules at the workspace root. The project has no tasks: VBA has no
// command-line compiler, linter or test runner, and none is invented.
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	var roots []string
	names := map[string][]string{}
	for _, manifest := range manifests {
		root := path.Dir(manifest)
		if root == "." {
			continue
		}
		if _, ok := names[root]; !ok {
			roots = append(roots, root)
		}
		names[root] = append(names[root], path.Base(manifest))
	}
	for _, root := range roots {
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(root))
		main, src := FindMain(names[root], func(name string) (string, bool) {
			raw, err := os.ReadFile(filepath.Join(dir, name))
			return string(raw), err == nil
		})
		if main == "" {
			continue
		}
		h := ParseHeader(src)
		result.Projects = append(result.Projects, protocol.Project{
			Name:     path.Base(root),
			Root:     root,
			Language: "vba",
			Kind:     projectkind.Resolve(h.Kind, root, protocol.ProjectLib),
			Exports:  []string{root},
			Version:  h.Version,
			Tasks:    map[string]protocol.Task{},
			Stack:    h.Stack,
		})
	}
	return result, nil
}
