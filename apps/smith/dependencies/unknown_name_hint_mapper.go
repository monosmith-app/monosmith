package dependencies

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/nearmatch"
)

// UnknownNameHint explains a package that failed to install when its name
// looks like a mistyped project of the workspace: the name matches no project
// and the registry does not have it, and the closest project names are the
// likely meant ones. It is empty when no name is close to a project.
//
// Parameters: projects are the names of the workspace's projects; packages are
// the names the person asked for.
// Returns: the hint, one line per mistyped name, or "".
func UnknownNameHint(projects, packages []string) string {
	var lines []string
	for _, pkg := range packages {
		if slices.Contains(projects, pkg) {
			continue
		}
		if near := nearmatch.Closest(pkg, projects); len(near) > 0 {
			lines = append(lines, pkg+" matches no project of this workspace and the registry does not have it. Closest projects: "+
				strings.Join(near, ", ")+" (smith install -w <project> workspace:"+near[0]+" links one)")
		}
	}
	return strings.Join(lines, "\n")
}
