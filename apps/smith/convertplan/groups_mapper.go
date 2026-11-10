package convertplan

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
	"github.com/monosmith-app/monosmith/apps/smith/taskfolding"
)

// planGroups writes the task overrides that several projects share once,
// as [[groups]], instead of once per project. A task counts as one a
// project has when Nx had a target of that name or the conversion made it a
// script; the rest of the plan keeps its per-project overrides.
func planGroups(plan *Plan, w *nxworkspace.Workspace) {
	targets := map[string]map[string]bool{}
	for _, p := range w.Projects {
		has := map[string]bool{}
		for name := range p.Targets {
			has[name] = true
		}
		targets[p.Name] = has
	}
	var projects []taskfolding.Project
	for _, pp := range plan.Projects {
		if !pp.Supported || pp.Language == "" {
			continue
		}
		has := map[string]bool{}
		for name := range targets[pp.Nx] {
			has[name] = true
		}
		for name := range pp.Scripts {
			has[name] = true
		}
		projects = append(projects, taskfolding.Project{
			Name: pp.Name, Language: pp.Language, Kind: string(pp.Kind),
			Has: has, Overrides: plan.Override(pp.Name).Tasks,
		})
	}
	folded := taskfolding.Fold(projects)
	if len(folded.Groups) == 0 {
		return
	}
	for _, p := range projects {
		o := plan.Override(p.Name)
		o.Tasks = folded.Overrides[p.Name]
		plan.SetOverride(p.Name, o)
	}
	plan.Config.Groups = append(plan.Config.Groups, folded.Groups...)
	for _, f := range folded.Folded {
		plan.Add(Mapped, "", f.Task, fmt.Sprintf("%d projects share the same %s settings, written once as a group for %s (%s)",
			len(f.Projects), f.Task, strings.Join(f.Select, " and "), strings.Join(f.Projects, ", ")))
	}
}
