package doctor

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// OutsideBuildDir lists the task outputs that fall outside the workspace's
// build directory, each with where the output was set: every build goes to
// dist/, so nothing is written into a source tree.
func OutsideBuildDir(projects []*projectgraph.Project) []string {
	var out []string
	for _, p := range projects {
		names := make([]string, 0, len(p.Tasks))
		for n := range p.Tasks {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			for _, o := range p.Tasks[n].Outputs {
				resolved := path.Clean(path.Join(p.Root, o))
				from := p.Sources[n+".outputs"]
				// A package whose smith.toml declares the dist/ beside its
				// package.json (main and types point there) has chosen it.
				ownDist := resolved == path.Join(p.Root, "dist") || strings.HasPrefix(resolved, path.Join(p.Root, "dist")+"/")
				if protocol.InBuildDir(resolved) || (ownDist && strings.HasPrefix(from, "smith.toml")) {
					continue
				}
				if from == "" {
					from = "plugin " + p.Plugin
				}
				out = append(out, fmt.Sprintf("%s:%s writes %s (from %s), not %s/", p.Name, n, resolved, from, protocol.ProjectBuildDir(p.Root)))
			}
		}
	}
	return out
}
