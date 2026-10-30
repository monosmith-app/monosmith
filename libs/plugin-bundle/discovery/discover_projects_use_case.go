package discovery

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ManifestFile is the file that makes a directory a bundle, together with an
// app.json beside it.
const ManifestFile = "flow.json"

// AppFile is the file the bundle's name, version and dependencies are read from.
const AppFile = "app.json"

// otherManifests are files a language plugin owns: a directory that has one is
// a project of that language, not a bundle.
var otherManifests = []string{"package.json", "Cargo.toml", "go.mod", "pyproject.toml", "pom.xml", "pubspec.yaml"}

// Discover describes the bundle behind each flow.json. A directory is a bundle
// when it has an app.json beside the flow.json and no manifest of a language
// plugin; the root is skipped. The project is named for its directory, exports
// the app.json name (when it differs) and imports what app.json's dependsOn
// lists, of which the host keeps the ones another project exports. It is
// always an app: it has a version and a release, and no sources of its own, so
// it has no tasks until smith.toml gives it some.
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	for _, manifest := range manifests {
		root := path.Dir(manifest)
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(root))
		if root == "." || ownedByLanguage(dir) {
			continue
		}
		src, err := os.ReadFile(filepath.Join(dir, AppFile))
		if err != nil {
			continue
		}
		app, err := ParseAppManifest(src)
		if err != nil {
			return result, fmt.Errorf("%s/%s: %w", root, AppFile, err)
		}
		project := protocol.Project{
			Name:     path.Base(root),
			Root:     root,
			Language: "bundle",
			Kind:     protocol.ProjectApp,
			Imports:  app.DependsOn,
			Version:  app.Version,
			Stack:    app.Stack,
			Tasks:    map[string]protocol.Task{},
		}
		if app.Name != "" && app.Name != project.Name {
			project.Exports = []string{app.Name}
		}
		result.Projects = append(result.Projects, project)
	}
	return result, nil
}

func ownedByLanguage(dir string) bool {
	for _, m := range otherManifests {
		if _, err := os.Stat(filepath.Join(dir, m)); err == nil {
			return true
		}
	}
	return false
}
