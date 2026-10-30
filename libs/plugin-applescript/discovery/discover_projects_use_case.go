package discovery

import (
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Discover describes the project of each directory that holds a
// .applescript source. The project is named for the directory, exports it and
// imports nothing: AppleScript has no declared dependencies smith could read.
// Its main script (main.applescript, else the first by name) carries the
// kind, stack and version in comments. A script is a lib unless its header or
// directory says otherwise. A script at the workspace root is skipped.
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
		main := PickMain(names[root])
		src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(root), main))
		if err != nil {
			return result, err
		}
		h := ParseHeader(string(src))
		name := path.Base(root)
		result.Projects = append(result.Projects, protocol.Project{
			Name:     name,
			Root:     root,
			Language: "applescript",
			Kind:     projectkind.Resolve(h.Kind, root, protocol.ProjectLib),
			Exports:  []string{root},
			Version:  h.Version,
			Tasks:    Tasks(root, name, main),
			Stack:    h.Stack,
			Tags:     []string{NativeTag},
		})
	}
	return result, nil
}
