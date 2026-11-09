package explain

import (
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/nearmatch"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// Field is one resolved value and where it came from.
type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Source string `json:"source"`
}

// Task is a resolved task.
type Task struct {
	ID          string  `json:"id"`
	Description string  `json:"description,omitempty"`
	Fields      []Field `json:"fields"`
}

// Explain resolves "project" (every task) or "project:task".
func Explain(g *projectgraph.Graph, target string) ([]Task, error) {
	name, only, _ := strings.Cut(target, ":")
	p := g.Get(name)
	if p == nil {
		return nil, fmt.Errorf("no project named %q (projects: %s).%s", name, strings.Join(g.Names(), ", "), nearmatch.Suggestion(name, g.Names()))
	}
	names := make([]string, 0, len(p.Tasks))
	for t := range p.Tasks {
		if only == "" || t == only {
			names = append(names, t)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("%s has no %q task", name, only)
	}
	sort.Strings(names)
	var out []Task
	for _, t := range names {
		def := p.Tasks[t]
		src := func(f string) string { return p.Sources[t+"."+f] }
		cwd := def.Cwd
		if cwd == "" {
			cwd = p.Root
		}
		inputs := def.Inputs
		if len(inputs) == 0 {
			inputs = []string{"**/* (every file of the project)"}
		}
		env := make([]string, 0, len(def.Env))
		for k, v := range def.Env {
			env = append(env, k+"="+v)
		}
		sort.Strings(env)
		out = append(out, Task{ID: name + ":" + t, Description: def.Description, Fields: []Field{
			{"command", strings.Join(def.Command, " "), src("command")},
			{"cwd", cwd, src("command")},
			{"dependsOn", list(def.DependsOn), src("dependsOn")},
			{"inputs", strings.Join(inputs, ", "), src("inputs")},
			{"outputs", list(def.Outputs), src("outputs")},
			{"env", list(env), src("env")},
			{"cache", fmt.Sprint(def.Cache), src("cache")},
		}})
	}
	return out, nil
}

func list(items []string) string {
	if len(items) == 0 {
		return "—"
	}
	return strings.Join(items, ", ")
}
