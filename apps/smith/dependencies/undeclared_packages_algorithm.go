package dependencies

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// undeclared returns the packages that a project's manifest does not declare,
// and the names it does declare.
func undeclared(declarations []protocol.Declaration, project string, packages []string) (missing, declared []string) {
	have := map[string]bool{}
	for _, d := range declarations {
		if d.Project == project {
			have[strings.ToLower(d.Name)] = true
			declared = append(declared, d.Name)
		}
	}
	for _, p := range packages {
		if !have[strings.ToLower(p)] {
			missing = append(missing, p)
		}
	}
	return missing, declared
}
