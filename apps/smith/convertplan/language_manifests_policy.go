package convertplan

import "github.com/monosmith-app/monosmith/apps/smith/nxworkspace"

// languageManifests are the files that make a folder a project of a language
// smith has a plugin for; a folder with two of them would become two projects
// with one name.
var languageManifests = []string{"package.json", "go.mod", "pyproject.toml", "Cargo.toml", "pom.xml", "pubspec.yaml"}

// manifestsAt lists the language manifests in the project's own folder.
func manifestsAt(p *nxworkspace.Project) []string {
	var found []string
	for _, m := range languageManifests {
		if p.HasFile(m) {
			found = append(found, m)
		}
	}
	return found
}
