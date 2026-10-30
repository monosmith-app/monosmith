package discovery

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Discover describes the project behind each Package.swift. A project is named
// for its directory and exports it; its path dependencies on directories
// outside it, inside the workspace, are its imports. Its kind is the
// smithKind comment, else the top directory, else an app when the package
// declares an executable and a lib when it does not. The manifest at the
// workspace root is skipped.
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	for _, manifest := range manifests {
		root := path.Dir(manifest)
		if root == "." {
			continue
		}
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(root))
		src, err := os.ReadFile(filepath.Join(dir, Manifest))
		if err != nil {
			return result, err
		}
		pkg := ParsePackage(string(src))
		var imports []string
		for _, dep := range pkg.Paths {
			if r := PackageRoot(root, dep); r != "" && r != root && !strings.HasPrefix(r, root+"/") {
				imports = append(imports, r)
			}
		}
		fallback := protocol.ProjectLib
		if pkg.Executable {
			fallback = protocol.ProjectApp
		}
		result.Projects = append(result.Projects, protocol.Project{
			Name:     path.Base(root),
			Root:     root,
			Language: "swift",
			Kind:     projectkind.Resolve(pkg.Kind, root, fallback),
			Exports:  []string{root},
			Imports:  imports,
			Version:  pkg.Version,
			Tasks:    Tasks(root),
			Stack:    pkg.Stack,
		})
	}
	return result, nil
}
