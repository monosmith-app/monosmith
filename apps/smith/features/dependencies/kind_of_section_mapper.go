package dependencies

import (
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// propertySectionPrefix starts the section a JVM plugin reports for a
// dependency whose version lives in a pom property.
const propertySectionPrefix = "property:"

// KindOf names, in one plain word, why a manifest declares a dependency.
// It maps the section a plugin reports (devDependencies, dev_dependencies,
// dev-dependencies, development_dependencies, require-dev, dev) to "dev",
// peer and optional sections to "peer" and "optional", build-dependencies to
// "build", and version tables that only pin a version for other projects
// (workspace.dependencies, dependencyManagement, a pom property) to
// "managed". Every other section (dependencies, require, a PackageReference,
// a Gradle line) is "prod".
//
// d is the declaration to classify. It returns the kind and never fails.
func KindOf(d protocol.Declaration) string {
	s := strings.ToLower(d.Section)
	switch {
	case d.Peer || strings.HasPrefix(s, "peer"):
		return "peer"
	case strings.HasPrefix(s, "optional"):
		return "optional"
	case strings.HasPrefix(s, "build"):
		return "build"
	case strings.Contains(s, "dev"):
		return "dev"
	case s == "workspace.dependencies", s == "dependencymanagement", strings.HasPrefix(s, propertySectionPrefix):
		return "managed"
	}
	return "prod"
}

// KindsOf lists the distinct kinds of the declarations of one package,
// sorted, so a package declared as dev in one project and prod in another
// reads "dev, prod".
//
// decls are the declarations of one package. It returns nil for none.
func KindsOf(decls []protocol.Declaration) []string {
	seen := map[string]bool{}
	var out []string
	for _, d := range decls {
		if k := KindOf(d); !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
