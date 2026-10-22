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
func Discover(files []string, plugins *pluginhost.PluginSet, catalog *pluginhost.Catalog, cfg workspace.Config) (*Graph, error) {
	var projects []*Project
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
		if len(manifests) == 0 {
			continue
		}
		sort.Strings(manifests)
		var res protocol.DiscoverResult
		if err := c.Call(protocol.MethodDiscover, protocol.DiscoverParams{Manifests: manifests}, &res, time.Minute); err != nil {
			return nil, err
		}
		for _, p := range res.Projects {
			projects = append(projects, &Project{Project: p, Plugin: c.Info.Plugin.Name})
		}
	}
	for _, p := range projects {
		if err := normalisePaths(p); err != nil {
			return nil, fmt.Errorf("plugin %s: %w", p.Plugin, err)
		}
	}
	ResolveEdges(projects)
	ResolveSchemaEdges(projects, files, cfg)
	if err := ApplyStacks(projects, plugins, catalog, cfg); err != nil {
		return nil, err
	}
	applied, err := ApplyConventions(projects, plugins, cfg)
	if err != nil {
		return nil, err
	}
	// Again: a plugin's generate task names the schema it reads.
	ResolveSchemaEdges(projects, files, cfg)
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
	if _, err := g.Order(); err != nil {
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
