package buildoutputs

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// Overlap is a pair of tasks whose outputs overlap.
type Overlap struct {
	// A and B are the tasks as "project:task", A before B in sorted order.
	A, B string
	// DirA and DirB are the workspace-relative output folders that overlap;
	// they are equal when both tasks write the same folder.
	DirA, DirB string
}

// Where is the folder the pair writes, or both folders when one holds the other.
func (o Overlap) Where() string {
	if o.DirA == o.DirB {
		return o.DirA
	}
	return fmt.Sprintf("%s and %s", o.DirA, o.DirB)
}

// Line is the one-line form of the overlap: "a and b both write dir".
func (o Overlap) Line() string {
	return fmt.Sprintf("%s and %s both write %s", o.A, o.B, o.Where())
}

// longRunningWords are the words of a task name that mark a task that runs
// until it is stopped (a watcher, a server): it is never run in the same run
// as a build, so its folder is not compared with the build's.
var longRunningWords = []string{"watch", "serve", "start"}

// LongRunning reports whether a task name is one of a task that runs until it
// is stopped: "dev", or a name with watch, serve or start as a word
// ("watch", "build:watch", "serve-docs"). The name "build-dev" is a build and
// is not long running.
func LongRunning(task string) bool {
	if task == "dev" {
		return true
	}
	words := strings.FieldsFunc(task, func(r rune) bool { return r == ':' || r == '-' || r == '_' || r == '.' })
	for _, w := range words {
		for _, l := range longRunningWords {
			if w == l {
				return true
			}
		}
	}
	return false
}

// Overlaps lists the pairs of tasks whose outputs overlap: one directory is
// the other, or holds it. A cache hit replaces a task's output directories
// wholesale, so restoring one task deletes what the other wrote, and two
// package tasks that both declare dist/drop fail each other mid-write.
//
// Three things are not reported: a directory in allowed ([workspace]
// shared_outputs), which several tasks fill on purpose, each into its own
// files; a pair that one group of alternatives ([workspace]
// alternative_outputs) lists together, because the tasks never run together;
// and a task that runs until it is stopped (see LongRunning).
func Overlaps(projects []*projectgraph.Project, allowed []string, alternatives [][]string) []Overlap {
	type output struct{ task, dir string }
	var all []output
	for _, p := range projects {
		for n, t := range p.Tasks {
			if LongRunning(n) {
				continue
			}
			for _, o := range t.Outputs {
				all = append(all, output{p.Name + ":" + n, path.Clean(path.Join(p.Root, o))})
			}
		}
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].task != all[j].task {
			return all[i].task < all[j].task
		}
		return all[i].dir < all[j].dir
	})
	var out []Overlap
	seen := map[string]bool{}
	for i, a := range all {
		for _, b := range all[i+1:] {
			overlap := a.dir == b.dir || strings.HasPrefix(b.dir, a.dir+"/") || strings.HasPrefix(a.dir, b.dir+"/")
			shorter := a.dir
			if len(b.dir) < len(shorter) {
				shorter = b.dir
			}
			if a.task == b.task || !overlap || InsideAny(shorter, allowed) || Alternative(alternatives, a.task, b.task) {
				continue
			}
			o := Overlap{A: a.task, B: b.task, DirA: a.dir, DirB: b.dir}
			if !seen[o.Line()] {
				seen[o.Line()] = true
				out = append(out, o)
			}
		}
	}
	return out
}

// Alternative reports whether one group of alternatives lists both tasks.
func Alternative(groups [][]string, a, b string) bool {
	for _, g := range groups {
		var hasA, hasB bool
		for _, t := range g {
			hasA = hasA || t == a
			hasB = hasB || t == b
		}
		if hasA && hasB {
			return true
		}
	}
	return false
}

// SharedOutputs lists the pairs of tasks whose outputs overlap, one line each
// ("a and b both write dir"). A directory in allowed ([workspace]
// shared_outputs) may be written by several tasks, each into its own files:
// overlap inside it is not reported. See Overlaps for the rest.
func SharedOutputs(projects []*projectgraph.Project, allowed ...string) []string {
	var out []string
	for _, o := range Overlaps(projects, allowed, nil) {
		out = append(out, o.Line())
	}
	return out
}

// OverlapWarnings turns SharedOutputs into the sentences smith ci generate and
// the assets run print: each names both tasks, says what goes wrong and gives
// the fixes. It returns nothing when no two tasks overlap.
func OverlapWarnings(projects []*projectgraph.Project, allowed ...string) []string {
	return OverlapWarningsFor(projects, allowed, nil)
}

// OverlapWarningsFor is OverlapWarnings with the groups of alternatives
// ([workspace] alternative_outputs), whose pairs are not reported.
func OverlapWarningsFor(projects []*projectgraph.Project, allowed []string, alternatives [][]string) []string {
	var out []string
	for _, o := range Overlaps(projects, allowed, alternatives) {
		out = append(out, "warning: "+o.Line()+". smith keeps each task's output in its cache, and a task clears its output folder before it runs, so one task deletes the files of the other. "+
			"Fix: give each task its own folder; list a folder that tasks fill on purpose, with different files, under [workspace] shared_outputs; "+
			"or list tasks that never run together under [workspace] alternative_outputs (smith doctor shows the exact lines).")
	}
	return out
}
