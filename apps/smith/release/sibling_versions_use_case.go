package release

import (
	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// SiblingVersions maps every exported identifier of the workspace's releasable
// projects to the version a dependent's range must name: the version this run
// releases, or for a package that is not releasing, its last released version.
// Without the second half a package released on its own would publish with the
// placeholder range its workspace manifest carries (^0.0.1) for a sibling that
// is already on the registry at a higher version, and fail to install.
func SiblingVersions(git *gitclient.Git, graph *projectgraph.Graph, plan Plan) (map[string]string, error) {
	order, err := graph.Order()
	if err != nil {
		return nil, err
	}
	versions := map[string]string{}
	for _, name := range order {
		p := graph.Get(name)
		if !p.Releases() {
			continue
		}
		latest, _, released := LatestTag(git, p)
		if !released {
			continue
		}
		for _, e := range p.Exports {
			versions[e] = latest.String()
		}
	}
	for _, item := range plan.Items {
		for _, e := range graph.Get(item.Project).Exports {
			versions[e] = item.Next
		}
	}
	return versions, nil
}
