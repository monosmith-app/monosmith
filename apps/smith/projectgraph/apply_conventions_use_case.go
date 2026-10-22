package projectgraph

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ApplyConventions asks every convention smith.toml configures for its
// additions to the projects it applies to, and merges them: a new task is
// added unless the project already has one by that name, and extra
// dependencies are added to existing tasks. It returns the projects each
// convention applies to, by convention name, including conventions that add
// no tasks.
func ApplyConventions(projects []*Project, plugins *pluginhost.PluginSet, cfg workspace.Config) (map[string][]string, error) {
	applied := map[string][]string{}
	for _, c := range plugins.Clients {
		name := c.Info.Plugin.Name
		conv, configured := cfg.Conventions[name]
		if !configured {
			continue
		}
		if c.Info.Kind != protocol.KindConvention {
			return nil, fmt.Errorf("smith.toml [conventions.%s]: %s is not a convention plugin", name, name)
		}
		targets, err := ConventionTargets(name, projects, c.Info.Languages, conv)
		if err != nil {
			return nil, err
		}
		for _, p := range targets {
			applied[name] = append(applied[name], p.Name)
			if settings := withProjectSettings(PluginSettings(name, conv, cfg.Projects[p.Name]), name, cfg.Projects[p.Name]); settings != nil {
				if p.Settings == nil {
					p.Settings = map[string]map[string]any{}
				}
				p.Settings[name] = settings
			}
		}
		// A convention that only sets projects up when they are generated
		// (eslint, jest) adds no tasks.
		if len(targets) == 0 || !c.Supports(protocol.MethodContribute) {
			continue
		}
		params := protocol.ContributeParams{Projects: make([]protocol.Project, len(targets))}
		for i, p := range targets {
			params.Projects[i] = p.Project
		}
		var res protocol.ContributeResult
		if err := c.Call(protocol.MethodContribute, params, &res, time.Minute); err != nil {
			return nil, fmt.Errorf("convention %s: %w", name, err)
		}
		for _, p := range targets {
			merge(p, res.Projects[p.Name], "convention "+name)
			p.addDeployer(name, c.Info, res.Projects[p.Name])
		}
	}
	return applied, nil
}

func merge(p *Project, c protocol.Contribution, source string) {
	if p.Tasks == nil {
		p.Tasks = map[string]protocol.Task{}
	}
	if p.contributed == nil {
		p.contributed = map[string]string{}
	}
	for _, name := range sortedTaskNames(c.Tasks) {
		if _, exists := p.Tasks[name]; exists {
			continue
		}
		p.Tasks[name] = c.Tasks[name]
		for _, f := range TaskFields {
			p.contributed[name+"."+f] = source
		}
	}
	for _, name := range sortedTaskNames(c.Replace) {
		p.Tasks[name] = c.Replace[name]
		for _, f := range TaskFields {
			p.contributed[name+"."+f] = source
		}
	}
	p.contributedConsumes = append(p.contributedConsumes, c.Consumes...)
	added := maps.Clone(c.Tasks)
	if added == nil {
		added = map[string]protocol.Task{}
	}
	maps.Copy(added, c.Replace)
	for _, name := range sortedTaskNames(added) {
		if !IsGenerateTask(name) {
			continue
		}
		for _, in := range added[name].Inputs {
			if file, ok := strings.CutPrefix(in, "{workspaceRoot}/"); ok && SchemaKind(file) != "" {
				p.contributedConsumes = append(p.contributedConsumes, file)
			}
		}
	}
	for _, name := range sortedTaskNames(c.DependsOn) {
		t, exists := p.Tasks[name]
		if !exists {
			continue
		}
		for _, d := range c.DependsOn[name] {
			if !slices.Contains(t.DependsOn, d) {
				t.DependsOn = append(t.DependsOn, d)
				p.contributed[name+".dependsOn"] = source
			}
		}
		p.Tasks[name] = t
	}
}

func sortedTaskNames[V any](m map[string]V) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
