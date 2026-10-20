package tui

import (
	"slices"
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// loadPlugins lists the built-in plugins (on or off in smith.toml) and the
// installed ones, with the projects each discovered or deploys, as
// `smith plugin list` does.
func loadPlugins(open func() (*session.Session, error), version string) pluginsLoaded {
	s, err := open()
	if err != nil {
		return pluginsLoaded{err: err}
	}
	defer s.Close()
	enabled := s.Config.Plugins.Enabled
	if len(enabled) == 0 {
		enabled = pluginhost.BuiltinNames()
	}
	used := map[string][]string{}
	for _, name := range s.Graph.Names() {
		p := s.Graph.Get(name)
		for _, plugin := range append([]string{p.Plugin}, p.Deployers...) {
			if plugin != "" && !slices.Contains(used[plugin], name) {
				used[plugin] = append(used[plugin], name)
			}
		}
	}
	var rows []pluginRow
	names := append(pluginhost.BuiltinNames(), sortedAddonNames()...)
	for _, name := range names {
		ctor, ok := pluginhost.Builtin(name)
		if !ok {
			continue
		}
		info, err := ctor(version).Describe(protocol.InitializeParams{})
		if err != nil {
			return pluginsLoaded{err: err}
		}
		row := pluginRow{
			Name: name, Source: "built-in", Builtin: true, UsedBy: used[name],
			Switchable: slices.Contains(pluginhost.BuiltinNames(), name), Enabled: slices.Contains(enabled, name),
		}
		if !row.Switchable {
			row.Enabled = true
		}
		for _, r := range info.EffectiveRoles() {
			row.Roles = append(row.Roles, string(r))
		}
		row.Roles = append(row.Roles, info.Provides...)
		for _, ref := range info.Requires {
			row.Requires = append(row.Requires, ref.String())
		}
		rows = append(rows, row)
	}
	installed, err := pluginhost.FindInstalled(s.Root)
	if err != nil {
		return pluginsLoaded{err: err}
	}
	for _, p := range installed {
		rows = append(rows, pluginRow{Name: p.Name, Version: p.Version, Source: p.Dir, Enabled: true, UsedBy: used[p.Name]})
	}
	return pluginsLoaded{rows: rows}
}

func sortedAddonNames() []string {
	var out []string
	for name := range pluginhost.BuiltinAddons {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
