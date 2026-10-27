package eject

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"
)

// projectsUsing returns the roots of the projects with a task that runs the
// tool on smith's default: a task command that names the default file. A
// project whose repository configures the tool has no such task.
func projectsUsing(g *projectgraph.Graph, tool tooldefaults.Tool) []string {
	needle := "/" + tooldefaults.Dir + "/" + tool.File
	var roots []string
	for _, p := range g.Projects {
		for _, task := range p.Tasks {
			if slices.ContainsFunc(task.Command, func(arg string) bool { return strings.HasSuffix(arg, needle) }) {
				roots = append(roots, p.Root)
				break
			}
		}
	}
	slices.Sort(roots)
	return roots
}
