package buildoutputs

import (
	"path"
	"slices"
	"strings"
)

// Task is the build task whose outputs are checked.
const Task = "build"

// Gap is a project whose build writes directories its build task does not
// declare as outputs.
type Gap struct {
	// Project is the smith project name; Root its directory.
	Project string `json:"project"`
	Root    string `json:"root"`
	// Written are the directories the build script writes, relative to the
	// project, outermost only.
	Written []string `json:"written"`
	// Declared are the outputs the task has now.
	Declared []string `json:"declared"`
	// Missing are the written directories no declared output restores.
	Missing []string `json:"missing"`
	// Outputs are the outputs to declare: what the build writes, plus the
	// outputs smith.toml already declares by hand.
	Outputs []string `json:"outputs"`
}

// GapOf finds the gap of one project: script is its build script, declared
// the outputs of its build task, and fromConfig tells whether smith.toml
// declared them (they are kept) or a plugin's default did (they are replaced
// by what the script writes). ok is false when the script's output is all
// declared.
func GapOf(project, root, script string, declared []string, fromConfig bool) (Gap, bool) {
	written := Collapse(WrittenDirs(script))
	missing := Uncovered(written, declared)
	if len(missing) == 0 {
		return Gap{}, false
	}
	outputs := slices.Clone(written)
	if fromConfig {
		for _, d := range declared {
			if d = path.Clean(d); !slices.Contains(outputs, d) {
				outputs = append(outputs, d)
			}
		}
	}
	return Gap{Project: project, Root: root, Written: written, Declared: declared, Missing: missing, Outputs: Collapse(outputs)}, true
}

// Problem says what is wrong in plain words.
func (g Gap) Problem() string {
	return g.Project + ":" + Task + " writes " + g.paths(g.Missing) + " but its task caches " + g.declared() + ": a cache hit would leave " + strings.Join(g.Missing, ", ") + " empty"
}

// Remedy says how to declare the outputs.
func (g Gap) Remedy() string {
	return "declare outputs = [" + quoteAll(g.Outputs) + "] under [projects." + tableKey(g.Project) + ".tasks." + Task + "] in smith.toml"
}

func (g Gap) paths(dirs []string) string {
	out := make([]string, len(dirs))
	for i, d := range dirs {
		out[i] = path.Join(g.Root, d)
	}
	return strings.Join(out, ", ")
}

func (g Gap) declared() string {
	if len(g.Declared) == 0 {
		return "nothing"
	}
	return g.paths(g.Declared)
}
