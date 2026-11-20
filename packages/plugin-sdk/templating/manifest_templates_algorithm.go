package templating

import (
	"path"
	"strings"
)

// DirPrefix starts the path of every template that goes into the project's
// folder.
const DirPrefix = "[[.Dir]]/"

// ManifestOnly keeps, of a project's templates, the ones a folder that holds
// the person's sources needs (GenerateParams.ManifestOnly): the files directly
// at the top of the project's folder that are not sources and not documents
// (package.json and its tsconfig, a pyproject.toml, a .csproj). Starter
// sources, files in sub-folders, README and LICENSE, and workspace files
// (paths that do not start with DirPrefix) are left out.
//
// Parameters: templates are the project's templates, keyed by their path
// templates; sources are the plugin's source patterns (*.ts, *.py).
// Returns: the templates kept, a new map.
func ManifestOnly(templates map[string]string, sources []string) map[string]string {
	out := map[string]string{}
	for target, body := range templates {
		rest, ok := strings.CutPrefix(target, DirPrefix)
		if !ok || strings.Contains(rest, "/") || isDocument(rest) || matchesAny(sources, rest) {
			continue
		}
		out[target] = body
	}
	return out
}

func isDocument(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".md") || strings.HasPrefix(lower, "license")
}

func matchesAny(patterns []string, name string) bool {
	for _, p := range patterns {
		if ok, _ := path.Match(p, name); ok {
			return true
		}
	}
	return false
}
