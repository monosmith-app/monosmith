package discovery

import (
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Discover describes the recipe behind each recipe file that is in a folder
// of its own. A recipe at the workspace root, one sharing a folder with
// another recipe or with other files, and one with no name are left alone:
// they are inputs of smith.toml's `recipe`, not projects with releases.
//
// A recipe is a package: it is versioned and released. A recipe that extends
// another in the workspace imports that recipe's folder, so a change in the
// parent affects the child.
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	for _, manifest := range manifests {
		root := path.Dir(manifest)
		if root == "." {
			continue
		}
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(root))
		file, ok := OwnFolder(dir)
		if !ok || file != path.Base(manifest) {
			continue
		}
		f, err := Read(filepath.Join(dir, file))
		if err != nil || f.Name == "" {
			continue
		}
		p := protocol.Project{
			Name:     f.Name,
			Root:     root,
			Language: "recipe",
			Kind:     protocol.ProjectPackage,
			Exports:  []string{root},
			Version:  f.Version,
			Tasks:    Tasks(file),
		}
		if parent := parentFolder(root, f.Extends); parent != "" {
			p.Imports = []string{parent}
		}
		result.Projects = append(result.Projects, p)
	}
	return result, nil
}

// parentFolder is the workspace folder of the recipe a path in extends names,
// or empty for a URL, a path outside the workspace or none.
func parentFolder(root, extends string) string {
	if extends == "" || strings.Contains(extends, "://") || filepath.IsAbs(extends) {
		return ""
	}
	if u, err := url.Parse(extends); err != nil || u.Scheme != "" {
		return ""
	}
	dir := path.Dir(path.Join(root, strings.ReplaceAll(extends, "\\", "/")))
	if dir == "." || dir == root || strings.HasPrefix(dir, "..") {
		return ""
	}
	return dir
}
