package doctor

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/buildoutputs"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// OutsideBuildDir lists the task outputs that fall outside the workspace's
// build directory, each with where the output was set: every build goes to
// dist/, so nothing is written into a source tree. A path in allowed
// ([workspace] allowed_outputs, workspace-relative) and anything under it is
// accepted: a folder a tool needs the output in, such as a go:embed folder.
func OutsideBuildDir(projects []*projectgraph.Project, allowed ...string) []string {
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
				if buildoutputs.InsideAny(resolved, allowed) || protocol.InBuildDir(resolved) || (ownDist && (strings.HasPrefix(from, "smith.toml") || keepsBuildsInProject(p))) {
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

// SharedWorkOutputs lists the task outputs that are, or hold, a folder
// another project ships from (its Task.Ship): a tool's work folder that the
// tool shares among projects, such as dist/cargo or dist/dotnet. A work folder
// under dist/<tool> is accepted as a build folder (OutsideBuildDir does not
// report it), but never as an output of a task when another project builds
// there: a cache restore empties each output folder first, which would delete
// the other project's build state (ADR 0092).
//
// Parameters: projects are the workspace's projects. Returns one line for
// each such output, in project and task order; none when every task caches
// only its own folders.
func SharedWorkOutputs(projects []*projectgraph.Project) []string {
	type source struct{ dir, project string }
	var shipped []source
	for _, p := range projects {
		for _, t := range p.Tasks {
			for _, g := range t.Ship {
				shipped = append(shipped, source{shipFolder(p.Root, g), p.Name})
			}
		}
	}
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
				for _, s := range shipped {
					if s.project != p.Name && (s.dir == resolved || strings.HasPrefix(s.dir, resolved+"/")) {
						out = append(out, fmt.Sprintf("%s:%s caches %s, where %s ships from", p.Name, n, resolved, s.project))
						break
					}
				}
			}
		}
	}
	return out
}

// shipFolder is the folder a ship glob of the project at root reads from,
// workspace-relative: the path before its first wildcard, without the file
// name.
func shipFolder(root, glob string) string {
	full := path.Join(root, glob)
	if i := strings.IndexAny(full, "*?["); i >= 0 {
		full = full[:i]
	}
	return path.Dir(path.Clean(full))
}

// keepsBuildsInProject reports whether a web app built per environment was
// given dist/<env> inside its own folder when it was added (ADR 0081).
func keepsBuildsInProject(p *projectgraph.Project) bool {
	v, _ := p.Settings["vite"]["output"].(string)
	return v == "project"
}
