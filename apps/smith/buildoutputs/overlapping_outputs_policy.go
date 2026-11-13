package buildoutputs

import (
	"fmt"
	"path"
	"sort"
	"strings"
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

// TaskOutputs is one task and the folders it writes.
type TaskOutputs struct {
	// ID is the task as "project:task"; Name is the task's own name.
	ID, Name string
	// Root is the project's folder, workspace-relative.
	Root string
	// Outputs are the task's outputs, relative to Root.
	Outputs []string
}

// longRunningWords are the words of a task name that mark a task that runs
// until it is stopped (a watcher, a server): its folder is not compared with
// the folders of the other tasks.
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

// Overlaps lists the pairs of tasks whose outputs overlap: one folder is the
// other, or holds it. A task clears its output folders before it runs, and a
// cache hit replaces them, so of two such tasks in one run the folder keeps
// only the last one's files, and two at once fail each other mid-write.
//
// Two things are not reported: a folder in allowed ([workspace]
// shared_outputs), which several tasks fill on purpose, each with its own
// files and never cleared; and a task that runs until it is stopped (see
// LongRunning).
//
// Parameters: tasks are the tasks to compare, allowed the shared folders.
// Returns the pairs, each once, A before B.
func Overlaps(tasks []TaskOutputs, allowed []string) []Overlap {
	type output struct{ task, dir string }
	var all []output
	for _, t := range tasks {
		if LongRunning(t.Name) {
			continue
		}
		for _, o := range t.Outputs {
			all = append(all, output{t.ID, path.Clean(path.Join(t.Root, o))})
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
			if a.task == b.task || !overlap || InsideAny(shorter, allowed) {
				continue
			}
			o := Overlap{A: a.task, B: b.task, DirA: a.dir, DirB: b.dir}
			if !seen[o.A+"\x00"+o.B] {
				seen[o.A+"\x00"+o.B] = true
				out = append(out, o)
			}
		}
	}
	return out
}
