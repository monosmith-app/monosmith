package doctor

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// taskGroupChecks warns about a [[task_groups]] entry that names no project:
// a selector that has a typo, or a project that was removed. Nothing is
// reported when there are no groups, or every group names a project.
func taskGroupChecks(s *session.Session) []protocol.Finding {
	return TaskGroupFindings(s.Graph.Projects, s.Config.TaskGroups)
}

// TaskGroupFindings is taskGroupChecks over the projects and groups.
func TaskGroupFindings(projects []*projectgraph.Project, groups []workspace.TaskGroup) []protocol.Finding {
	var out []protocol.Finding
	for _, g := range groups {
		matched := false
		for _, p := range projects {
			if p.InTaskGroup(g) {
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		selectors := strings.TrimPrefix(projectgraph.TaskGroupSource(g), "smith.toml [[task_groups]] ")
		out = append(out, protocol.Finding{
			Check:   "every task group names a project",
			Warning: true,
			Detail:  fmt.Sprintf("the [[task_groups]] with %s matches no project, so its settings apply to nothing", selectors),
			Remedy:  "check the selectors in smith.toml against your project names, folders, kinds, languages and tags, or delete the group",
		})
	}
	return out
}
