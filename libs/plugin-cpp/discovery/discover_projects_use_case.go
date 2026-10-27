package discovery

import (
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Discover describes the project behind each CMakeLists.txt that calls
// project(): one without is a subdirectory of a project. A project exports
// its directory, and the workspace directories it adds with
// add_subdirectory are its imports.
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	for _, manifest := range manifests {
		root := path.Dir(manifest)
		if root == "." {
			continue
		}
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(root))
		src, err := os.ReadFile(filepath.Join(dir, "CMakeLists.txt"))
		if err != nil {
			return result, err
		}
		lists := ParseLists(string(src))
		if !lists.Project {
			continue
		}
		var imports []string
		for _, s := range lists.Subdirectories {
			if r := SubdirectoryRoot(root, s); r != "" && r != root && !isWithin(r, root) {
				imports = append(imports, r)
			}
		}
		var lint []string
		for _, d := range []string{"src", "include", "tests"} {
			if _, err := os.Stat(filepath.Join(dir, d)); err == nil {
				lint = append(lint, d)
			}
		}
		kind := projectkind.Resolve(lists.Kind, root, protocol.ProjectLib)
		var tags []string
		if lists.Native {
			tags = append(tags, NativeTag)
		}
		result.Projects = append(result.Projects, protocol.Project{
			Name:     path.Base(root),
			Root:     root,
			Language: "cpp",
			Kind:     kind,
			Exports:  []string{root},
			Imports:  imports,
			Version:  lists.Version,
			Tasks:    Tasks(Shape{Root: root, Kind: kind, Executable: lists.Executable, Lint: lint}),
			Stack:    lists.Stack,
			Tags:     tags,
		})
	}
	result.SetToolchain(func() string { return toolrun.Toolchain("cmake", "cmake", "--version") })
	return result, nil
}

func isWithin(p, root string) bool {
	return len(p) > len(root) && p[:len(root)+1] == root+"/"
}
