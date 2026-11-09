package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/internallink"
	"github.com/monosmith-app/monosmith/apps/smith/tui/packageprompt"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
)

// linkableLibraries lists the libraries of the workspace that project may
// depend on (not itself, not an app, the same language, no cycle). It reads the
// projects it is given and nothing of the Shell.
func linkableLibraries(projects []projectlist.Project, project string) []packageprompt.Library {
	nodes := make([]internallink.Node, len(projects))
	versions := map[string]string{}
	for i, pr := range projects {
		nodes[i] = internallink.Node{Name: pr.Name, Root: pr.Root, Kind: pr.Kind, Language: pr.Language, Deps: pr.Deps}
		versions[pr.Name] = pr.Version
	}
	var out []packageprompt.Library
	for _, n := range internallink.ValidLibraries(nodes, project) {
		out = append(out, packageprompt.Library{Name: n.Name, Kind: n.Kind, Language: n.Language, Version: versions[n.Name]})
	}
	return out
}
