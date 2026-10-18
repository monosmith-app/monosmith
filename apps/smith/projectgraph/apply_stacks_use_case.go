package projectgraph

import (
	"fmt"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// StackOf is the name of the stack a project uses: smith.toml
// [projects.<name>] stack, else the one its manifest records, else the
// workspace default for its language plugin. Empty means none: the language
// plugin alone.
func StackOf(p *Project, cfg workspace.Config) string {
	if o, ok := cfg.Projects[p.Name]; ok && o.Stack != "" {
		return o.Stack
	}
	if p.Stack != "" {
		return p.Stack
	}
	if s, ok := cfg.DefaultStacks[p.Plugin]; ok {
		return s
	}
	return cfg.DefaultStacks[p.Language]
}

// ApplyStacks resolves each project's stack and lets the stack's toolchain,
// structure and deploy plugins contribute, in the stack's order: a plugin
// adds tasks the project lacks and may replace ones it has. Each plugin is
// asked once, about every project whose stack names it.
func ApplyStacks(projects []*Project, plugins *pluginhost.PluginSet, catalog *pluginhost.Catalog, cfg workspace.Config) error {
	stacks := map[string]protocol.Stack{}
	users := map[string][]*Project{}
	for _, p := range projects {
		name := StackOf(p, cfg)
		if name == "" || name == protocol.NoStack {
			p.Stack = name
			continue
		}
		st, err := catalog.Get(name)
		if err != nil {
			return fmt.Errorf("project %s: %w", p.Name, err)
		}
		if st.Language != p.Plugin {
			return fmt.Errorf("project %s is discovered by the %s plugin, but stack %s is for %s projects", p.Name, p.Plugin, name, st.Language)
		}
		p.Stack = name
		p.stackPlugins = st.Plugins()
		if st.Scaffold != "" {
			p.stackPlugins = append(p.stackPlugins, st.Scaffold)
		}
		stacks[p.Name] = st
		for _, n := range st.Plugins() {
			users[n] = append(users[n], p)
		}
		// Every plugin of the stack, its language and scaffold included, reads
		// the settings smith.toml holds for it, which a recipe wrote when it
		// created the project.
		for _, n := range append(st.Plugins(), p.Plugin, st.Scaffold) {
			if settings := ProjectSettings(n, cfg.Projects[p.Name]); n != "" && settings != nil {
				if p.Settings == nil {
					p.Settings = map[string]map[string]any{}
				}
				p.Settings[n] = settings
			}
		}
	}
	contributed := map[string]protocol.ContributeResult{}
	for _, n := range sortedTaskNames(users) {
		c := plugins.Get(n)
		if c == nil || !c.Supports(protocol.MethodContribute) {
			continue
		}
		params := protocol.ContributeParams{}
		for _, p := range users[n] {
			params.Projects = append(params.Projects, p.Project)
		}
		var res protocol.ContributeResult
		if err := c.Call(protocol.MethodContribute, params, &res, time.Minute); err != nil {
			return fmt.Errorf("plugin %s: %w", n, err)
		}
		contributed[n] = res
	}
	for _, p := range projects {
		st, ok := stacks[p.Name]
		if !ok {
			continue
		}
		for _, n := range st.Plugins() {
			merge(p, contributed[n].Projects[p.Name], fmt.Sprintf("plugin %s (stack %s)", n, st.Name))
			if c := plugins.Get(n); c != nil {
				p.addDeployer(n, c.Info, contributed[n].Projects[p.Name])
			}
		}
	}
	return nil
}
