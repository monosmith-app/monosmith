package environments

import (
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vite/webenv"
)

// Change is a change to the environments of a web app that exists: the new
// settings, the files and scripts it creates, and the files it leaves alone.
type Change struct {
	// Project is the app and Root its workspace-relative folder.
	Project, Root string
	// Before and After are the settings as they were and as they will be.
	Before, After webenv.Settings
	// Created are the files to write, by path relative to the app. A file that
	// exists is never replaced.
	Created map[string]string
	// ScriptsAdded are the package.json scripts to add (name to command);
	// ScriptsRemoved are the names to remove.
	ScriptsAdded   map[string]string
	ScriptsRemoved []string
	// Kept are the env files of removed environments: they hold the person's
	// values, so smith leaves them.
	Kept []string
}

// PlanChange works out what adding and removing environments does to an app.
// It writes nothing. The default environment cannot be removed: it is the task
// build, which CI and releases use.
//
// Parameters: p is the app, with the settings smith.toml records; add and
// remove are environment names.
// Returns: the change; an error for an app that has no environments (see the
// manual for an app that exists), a name that is not valid, a name that is
// already there or is not, the default environment, and an empty list.
func PlanChange(p *projectgraph.Project, add, remove []string) (Change, error) {
	before, ok := webenv.FromMap(p.Settings[webenv.PluginName])
	if !ok {
		return Change{}, fmt.Errorf("%s has no environments: it was made with a single build. See docs/manual/environments.md, \"An app that exists\", to move it by hand", p.Name)
	}
	if len(add) == 0 && len(remove) == 0 {
		return Change{}, fmt.Errorf("name the environments to add or to remove")
	}
	after := before
	after.Names = slices.Clone(before.Names)
	for _, name := range add {
		if slices.Contains(after.Names, name) {
			return Change{}, fmt.Errorf("%s has the environment %q already", p.Name, name)
		}
		after.Names = append(after.Names, name)
	}
	var kept []string
	for _, name := range remove {
		switch {
		case !slices.Contains(after.Names, name):
			return Change{}, fmt.Errorf("%s has no environment %q (it has %s)", p.Name, name, strings.Join(before.Names, ", "))
		case name == before.Default:
			return Change{}, fmt.Errorf("%q is the default environment of %s, the task build that CI and releases use: make another one the default first (smith.toml, [projects.%s.settings.vite] default_environment) and edit its script", name, p.Name, p.Name)
		}
		after.Names = slices.DeleteFunc(after.Names, func(n string) bool { return n == name })
		kept = append(kept, p.Root+"/"+webenv.EnvFileName(name))
	}
	if err := after.Validate(); err != nil {
		return Change{}, err
	}
	c := Change{Project: p.Name, Root: p.Root, Before: before, After: after, Created: map[string]string{}, ScriptsAdded: map[string]string{}, Kept: kept}
	scripts := webenv.Scripts(after, p.Root)
	for _, name := range add {
		c.Created[webenv.EnvFileName(name)] = webenv.EnvFile(name, after)
		task := webenv.TaskName(after, name)
		c.ScriptsAdded[task] = scripts[task]
	}
	for _, name := range remove {
		c.ScriptsRemoved = append(c.ScriptsRemoved, webenv.TaskName(before, name))
	}
	return c, nil
}

// Plan is what the change will do, as the plan the commands show before they
// write.
func (c Change) Plan() operationplan.Plan {
	plan := operationplan.Plan{Title: "environments of " + c.Project}
	const files = "Files"
	for _, name := range slices.Sorted(mapKeys(c.Created)) {
		plan.Add(files, operationplan.Line{Action: operationplan.Create, Path: c.Root + "/" + name, Note: "starter VITE_ values for the new environment"})
	}
	for _, task := range slices.Sorted(mapKeys(c.ScriptsAdded)) {
		plan.Add("Tasks", operationplan.Line{Action: operationplan.Change, Path: c.Root + "/package.json", Note: fmt.Sprintf("adds the script %s: %s", task, c.ScriptsAdded[task])})
	}
	for _, task := range c.ScriptsRemoved {
		plan.Add("Tasks", operationplan.Line{Action: operationplan.Change, Path: c.Root + "/package.json", Note: "removes the script " + task})
	}
	plan.Add("Settings", operationplan.Line{Action: operationplan.Change, Path: "smith.toml", Note: fmt.Sprintf("[projects.%s.settings.vite] environments: %s -> %s", c.Project, strings.Join(c.Before.Names, ", "), strings.Join(c.After.Names, ", "))})
	return plan
}

func mapKeys[V any](m map[string]V) func(yield func(string) bool) {
	return func(yield func(string) bool) {
		for k := range m {
			if !yield(k) {
				return
			}
		}
	}
}
