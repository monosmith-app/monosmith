package doctor

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// groupChecks warns about a [[groups]] entry that names no project:
// a selector that has a typo, or a project that was removed. Nothing is
// reported when there are no groups, or every group names a project.
func groupChecks(s *session.Session) []protocol.Finding {
	return GroupFindings(s.Graph.Projects, s.Config.Groups)
}

// GroupFindings is groupChecks over the projects and groups.
func GroupFindings(projects []*projectgraph.Project, groups []workspace.Group) []protocol.Finding {
	var out []protocol.Finding
	for _, g := range groups {
		matched := false
		for _, p := range projects {
			if p.InGroup(g) {
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		selectors := strings.TrimPrefix(projectgraph.GroupSource(g), "smith.toml [[groups]] ")
		out = append(out, protocol.Finding{
			Check:   "every group names a project",
			Warning: true,
			Detail:  fmt.Sprintf("the [[groups]] with %s matches no project, so its settings apply to nothing", selectors),
			Remedy:  "check the selectors in smith.toml against your project names, folders, kinds, languages and tags, or delete the group",
		})
	}
	return out
}
