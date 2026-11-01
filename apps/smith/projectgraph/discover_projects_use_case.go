package projectgraph

import (
	"fmt"
	"path"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Discover asks every plugin to describe the manifests it claims among the
// workspace's files (tracked, or untracked and not ignored, outside
// [workspace] exclude), then resolves
// edges, applies each project's stack, the configured conventions and then
// smith.toml overrides.
//
// It fails when projects import each other in a cycle, because tasks cannot be
// ordered; DiscoverAllowingCycles is for the commands that explain the cycle.
func Discover(files []string, plugins *pluginhost.PluginSet, catalog *pluginhost.Catalog, cfg workspace.Config) (*Graph, error) {
	g, err := DiscoverAllowingCycles(files, plugins, catalog, cfg)
	if err != nil {
		return nil, err
	}
	if _, err := g.Order(); err != nil {
		return nil, err
	}
	return g, nil
}

// DiscoverAllowingCycles is Discover for a workspace whose projects may import
// each other in a cycle: the graph opens, Cycles names them, and Order is the
// call that refuses. smith doctor and smith graph use it to report the cycle.
func DiscoverAllowingCycles(files []string, plugins *pluginhost.PluginSet, catalog *pluginhost.Catalog, cfg workspace.Config) (*Graph, error) {
	var projects []*Project
	asked := map[string]bool{}
	for _, c := range plugins.Clients {
		if !c.Info.Capabilities.Discover {
			continue
		}
		var manifests []string
		for _, f := range files {
			if protocol.IsManifest(c.Info.Manifests, path.Base(f)) && !cfg.Excluded(f) {
				manifests = append(manifests, f)
			}
		}
		declared := declaredFor(c.Info.Plugin.Name, c.Info.Languages, cfg)
		if len(manifests) == 0 && len(declared) == 0 {
			continue
		}
		sort.Strings(manifests)
		var res protocol.DiscoverResult
		if err := c.Call(protocol.MethodDiscover, protocol.DiscoverParams{Manifests: manifests, Declared: declared}, &res, time.Minute); err != nil {
			return nil, err
		}
		described := map[string]string{}
		for _, p := range res.Projects {
			described[p.Name] = p.Root
			projects = append(projects, &Project{Project: p, Plugin: c.Info.Plugin.Name})
		}
		for _, d := range declared {
			asked[d.Name] = true
			if root, ok := described[d.Name]; !ok || root != d.Root {
				return nil, fmt.Errorf("smith.toml [projects.%s] declares a %s project at %s, which the %s plugin did not describe: it needs a folder it can read", d.Name, cfg.Projects[d.Name].Language, d.Root, c.Info.Plugin.Name)
			}
		}
	}
	for name, po := range cfg.Projects {
		if po.Declared() && !asked[name] {
			return nil, fmt.Errorf("smith.toml [projects.%s] declares language %q, which no enabled plugin has: see smith plugin list", name, po.Language)
		}
	}
	for _, p := range projects {
		if err := normalisePaths(p); err != nil {
			return nil, fmt.Errorf("plugin %s: %w", p.Plugin, err)
		}
	}
	ResolveEdges(projects)
	ResolveSchemaEdges(projects, files, cfg)
	if err := SeedFacts(projects, cfg); err != nil {
		return nil, err
	}
	if err := ApplyStacks(projects, plugins, catalog, cfg); err != nil {
		return nil, err
	}
	// The conventions' requests carry the stacks' default facts too.
	if err := ResolveFacts(projects); err != nil {
		return nil, err
	}
	applied, err := ApplyConventions(projects, plugins, cfg)
	if err != nil {
		return nil, err
	}
	if err := ResolveFacts(projects); err != nil {
		return nil, err
	}
	for convention, names := range applied {
		for _, p := range projects {
			if slices.Contains(names, p.Name) {
				p.Conventions = append(p.Conventions, convention)
			}
		}
	}
	for _, p := range projects {
		sort.Strings(p.Conventions)
		p.toolset = toolset(p, plugins)
		if err := ApplyOverrides(p, cfg); err != nil {
			return nil, err
		}
		// Again: conventions, stacks and smith.toml may have set paths too.
		if err := normalisePaths(p); err != nil {
			return nil, err
		}
	}
	g, err := New(projects)
	if err != nil {
		return nil, err
	}
	g.SetExcluded(cfg.Excluded)
	return g, nil
}

// toolset is the language plugin, then the project's other plugins sorted,
// each with the version it reported.
func toolset(p *Project, plugins *pluginhost.PluginSet) string {
	others := map[string]bool{}
	for _, n := range append(append([]string(nil), p.stackPlugins...), p.Conventions...) {
		if n != p.Plugin {
			others[n] = true
		}
	}
	names := append([]string{p.Plugin}, sortedTaskNames(others)...)
	parts := make([]string, len(names))
	for i, n := range names {
		parts[i] = n
		if c := plugins.Get(n); c != nil {
			parts[i] += "@" + c.Info.Plugin.Version
		}
	}
	return strings.Join(parts, "+")
}
