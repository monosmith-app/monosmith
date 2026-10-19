package affected

import (
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// WatchPaths lists, for each app, the directories a deployment platform
// should watch (Railway and Render's watch paths, a GitOps path filter): the
// app's own and those of every project it depends on, transitively. A change
// anywhere else cannot change what the app deploys. Directories are
// workspace-relative with forward slashes, sorted; the map holds apps only.
func WatchPaths(g *projectgraph.Graph) map[string][]string {
	out := map[string][]string{}
	for _, p := range g.Projects {
		if p.Kind != protocol.ProjectApp {
			continue
		}
		dirs := map[string]bool{p.Root: true}
		for _, dep := range g.TransitiveDeps(p.Name) {
			if d := g.Get(dep); d != nil {
				dirs[d.Root] = true
			}
		}
		list := make([]string, 0, len(dirs))
		for d := range dirs {
			list = append(list, d)
		}
		sort.Strings(list)
		out[p.Name] = list
	}
	return out
}
