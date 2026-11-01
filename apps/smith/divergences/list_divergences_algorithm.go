// Package divergences lists, calmly, where a workspace differs from what smith
// would do by itself: task overrides, files a person owns, files edited by
// hand and files that differ from their siblings. Each entry says why it is
// listed and how to undo it (ADR 0064). It is a report; it never fails a run.
package divergences

import (
	"fmt"
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/derive"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// Kind groups divergences.
type Kind string

// The kinds, in the order they are printed.
const (
	// Override is a setting in smith.toml that replaces what a plugin infers.
	Override Kind = "override"
	// Ejected is a derived file a person owns.
	Ejected Kind = "ejected"
	// Drift is a derived file that was edited by hand, or that smith cannot edit.
	Drift Kind = "drift"
	// Differs is a file that differs from the shape of its sibling projects.
	Differs Kind = "differs"
)

// Divergence is one place where the workspace differs from the default.
type Divergence struct {
	Kind Kind `json:"kind"`
	// Subject is what differs: "web test", "packages/web/tsconfig.json".
	Subject string `json:"subject"`
	// Why says what makes it differ, in a sentence.
	Why string `json:"why"`
	// Undo says how to go back to the default, in a sentence.
	Undo string `json:"undo"`
}

// From lists the divergences of a workspace.
//
// Parameters: cfg is smith.toml; items and diverging are the derived-file plan
// (nil when derived files are off).
// Returns the divergences, grouped by kind, then by subject.
func From(cfg workspace.Config, items []derive.Item, diverging []derive.Divergence) []Divergence {
	var out []Divergence
	for task, o := range cfg.Tasks {
		for _, field := range fieldsSet(o) {
			out = append(out, Divergence{
				Kind:    Override,
				Subject: "all projects, " + task + " " + field,
				Why:     fmt.Sprintf("[tasks.%s] sets %s for every project that has the task, instead of what its plugin infers", task, field),
				Undo:    fmt.Sprintf("remove %s from [tasks.%s] in smith.toml", field, task),
			})
		}
	}
	for project, po := range cfg.Projects {
		for task, o := range po.Tasks {
			for _, field := range fieldsSet(o) {
				out = append(out, Divergence{
					Kind:    Override,
					Subject: project + " " + task + " " + field,
					Why:     fmt.Sprintf("[projects.%s.tasks.%s] sets %s instead of what the plugin or the workspace gives", project, task, field),
					Undo:    fmt.Sprintf("remove %s from [projects.%s.tasks.%s] in smith.toml", field, project, task),
				})
			}
		}
		if po.Stack != "" {
			out = append(out, Divergence{
				Kind:    Override,
				Subject: project + " stack",
				Why:     fmt.Sprintf("[projects.%s] names the stack %s, not the one its manifest records", project, po.Stack),
				Undo:    fmt.Sprintf("remove stack from [projects.%s] in smith.toml", project),
			})
		}
	}
	for _, it := range items {
		path := it.File.Path
		switch it.Status {
		case derive.StatusEjected:
			out = append(out, Divergence{
				Kind: Ejected, Subject: path,
				Why:  "you own this file; smith does not touch it",
				Undo: fmt.Sprintf("smith derive --reclaim %s gives it back to smith (it shows the diff and asks first)", path),
			})
		case derive.StatusDrift, derive.StatusBroken:
			out = append(out, Divergence{
				Kind: Drift, Subject: path,
				Why:  it.Detail,
				Undo: fmt.Sprintf("smith derive --reclaim %s puts back what smith derives (it shows the diff and asks first); --adopt or --eject keeps your edit", path),
			})
		case derive.StatusDiverged:
			out = append(out, Divergence{
				Kind: Drift, Subject: path,
				Why:  "the file exists, differs from what smith derives, and smith.lock does not list it",
				Undo: fmt.Sprintf("smith derive --reclaim %s puts back what smith derives (it shows the diff and asks first)", path),
			})
		}
	}
	for _, d := range diverging {
		out = append(out, Divergence{
			Kind: Differs, Subject: d.Path,
			Why:  d.Detail,
			Undo: "make it like its sibling projects, or leave it: the project owns it",
		})
	}
	order := map[Kind]int{Override: 0, Ejected: 1, Drift: 2, Differs: 3}
	sort.Slice(out, func(i, j int) bool {
		if order[out[i].Kind] != order[out[j].Kind] {
			return order[out[i].Kind] < order[out[j].Kind]
		}
		return out[i].Subject < out[j].Subject
	})
	return out
}

// fieldsSet names the fields of a task override that are set, sorted.
func fieldsSet(o workspace.TaskOverride) []string {
	var f []string
	add := func(set bool, name string) {
		if set {
			f = append(f, name)
		}
	}
	add(len(o.Command) > 0, "command")
	add(o.Cwd != "", "cwd")
	add(len(o.Inputs) > 0, "inputs")
	add(len(o.Outputs) > 0, "outputs")
	add(len(o.DependsOn) > 0, "depends_on")
	add(len(o.Env) > 0, "env")
	add(o.Cache != nil, "cache")
	add(o.Verify != nil, "verify")
	add(o.Serve != "" || o.ServeURL != "" || o.ServeTimeout != "", "serve")
	sort.Strings(f)
	return f
}
