// Package overridelearn reads the per-project task overrides in smith.toml and
// reports the ones that say something about the workspace defaults: the same
// value in several projects, or a default that nearly every project overrides
// (ADR 0056). It reports; it never edits.
package overridelearn

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

const (
	// DefaultThreshold is how many projects must set the same value before smith
	// suggests a workspace default.
	DefaultThreshold = 3
	// replaceShare is the share of all projects that must set a value before it
	// is a default, or before a default is called wrong. Fewer is a job for
	// [[task_groups]] (smith doctor says so already).
	replaceShare = 0.8
)

// Analyse finds the suggestions in the per-project task overrides.
//
// Parameters: cfg is smith.toml; total is the number of projects in the
// workspace; threshold is the number of projects that makes a pattern (use
// DefaultThreshold).
// Returns: suggestions, Lift before Replace, then by task and field.
func Analyse(cfg workspace.Config, total, threshold int) []Suggestion {
	type key struct{ task, field, value string }
	setBy := map[key][]string{}
	overridden := map[[2]string][]string{} // task, field -> projects
	for project, po := range cfg.Projects {
		for task, o := range po.Tasks {
			for field, value := range fieldsOf(o) {
				setBy[key{task, field, value}] = append(setBy[key{task, field, value}], project)
				if def, ok := fieldsOf(cfg.Tasks[task])[field]; ok && def != value {
					f := [2]string{task, field}
					overridden[f] = append(overridden[f], project)
				}
			}
		}
	}
	var out []Suggestion
	for k, projects := range setBy {
		if len(projects) < threshold || float64(len(projects)) < replaceShare*float64(total) {
			continue
		}
		if def, ok := fieldsOf(cfg.Tasks[k.task])[k.field]; ok && def == k.value {
			continue
		}
		sort.Strings(projects)
		out = append(out, Suggestion{Kind: Lift, Task: k.task, Field: k.field, Value: k.value, Projects: projects, Total: total})
	}
	for f, projects := range overridden {
		def, ok := fieldsOf(cfg.Tasks[f[0]])[f[1]]
		if !ok || len(projects) < threshold || float64(len(projects)) < replaceShare*float64(total) {
			continue
		}
		sort.Strings(projects)
		out = append(out, Suggestion{Kind: Replace, Task: f[0], Field: f[1], Value: def, Projects: projects, Total: total})
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Task != b.Task {
			return a.Task < b.Task
		}
		if a.Field != b.Field {
			return a.Field < b.Field
		}
		return a.Value < b.Value
	})
	return out
}

// fieldsOf lists the fields an override sets, each as written in smith.toml;
// an environment variable is its own field ("env.CI").
func fieldsOf(o workspace.TaskOverride) map[string]string {
	f := map[string]string{}
	list := func(name string, v []string) {
		if len(v) > 0 {
			f[name] = fmt.Sprintf("[%s]", strings.Join(quoteAll(v), ", "))
		}
	}
	list("command", o.Command)
	list("inputs", o.Inputs)
	list("outputs", o.Outputs)
	list("depends_on", o.DependsOn)
	text := func(name, v string) {
		if v != "" {
			f[name] = fmt.Sprintf("%q", v)
		}
	}
	text("cwd", o.Cwd)
	text("serve", o.Serve)
	text("serve_url", o.ServeURL)
	text("serve_timeout", o.ServeTimeout)
	flag := func(name string, v *bool) {
		if v != nil {
			f[name] = fmt.Sprint(*v)
		}
	}
	flag("cache", o.Cache)
	flag("verify", o.Verify)
	for k, v := range o.Env {
		f["env."+k] = fmt.Sprintf("%q", v)
	}
	return f
}

func quoteAll(v []string) []string {
	out := make([]string, len(v))
	for i, s := range v {
		out[i] = fmt.Sprintf("%q", s)
	}
	return out
}
