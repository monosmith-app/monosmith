package taskoutput

import (
	"path"
	"slices"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckShippedOutput reads what a build task caches against the two folders
// of a build (ADR 0092): the shipped folder dist/<project dir>, which the task
// lists as its output, and the tool's work folder, which it never lists when
// the tool shares it among projects. A cache restore empties each output
// folder first, so a shared work folder as an output would delete the build
// state of every other project that builds there.
//
// Two things are reported. An output that is, holds or sits inside one of
// shared is a failure. A task that ships files (Task.Ship) but does not list
// the shipped folder as an output is a failure when the task is cached (a
// cache hit restores nothing), else a warning.
//
// Parameters: p is the project; task the task name; shared the work folders,
// workspace-relative, that the tool shares among projects (dist/cargo).
// Returns the finding, and false when the task is absent or follows the rule.
func CheckShippedOutput(p protocol.Project, task string, shared ...string) (protocol.Finding, bool) {
	t, ok := p.Tasks[task]
	if !ok {
		return protocol.Finding{}, false
	}
	shipped := protocol.ProjectBuildDir(p.Root)
	check := p.Name + " caches only " + shipped
	remedy := "list only " + protocol.BuildDirFromProject(p.Root) + " in the outputs of [projects." + p.Name + ".tasks." + task + "] in smith.toml"
	resolved := make([]string, len(t.Outputs))
	for i, o := range t.Outputs {
		resolved[i] = path.Clean(path.Join(p.Root, o))
		for _, s := range shared {
			if within(resolved[i], s) || within(s, resolved[i]) {
				return protocol.Finding{
					Check: check, Project: p.Name,
					Problem: "caches " + resolved[i] + ", a work folder other projects share",
					Detail:  "the " + task + " task lists " + resolved[i] + " as an output. A cache restore empties each output folder first, so it would delete the build state of every project that builds into " + s + "/",
					Remedy:  remedy,
				}, true
			}
		}
	}
	if len(t.Ship) == 0 || slices.Contains(resolved, shipped) {
		return protocol.Finding{}, false
	}
	f := protocol.Finding{
		Check: check, Project: p.Name, Warning: !t.Cache,
		Problem: "does not cache " + shipped,
		Detail:  "the " + task + " task links what it ships into " + shipped + "/ but does not list that folder as an output",
		Remedy:  remedy,
	}
	if t.Cache {
		f.Detail += ", so a cache hit restores nothing"
	}
	return f, true
}

// within reports whether the workspace-relative folder dir is base or sits
// inside it.
func within(dir, base string) bool {
	return dir == base || len(dir) > len(base) && dir[:len(base)+1] == base+"/"
}
