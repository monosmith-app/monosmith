package pluginsview

import (
	"slices"
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Load lists the built-in plugins (on or off in smith.toml) and the
// installed ones, with the projects each discovered or deploys, as
// `smith plugin list` does.
func Load(open func() (*session.Session, error), version string) Result {
	s, err := open()
	if err != nil {
		return Result{Err: err}
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
	var rows []Row
	names := append(pluginhost.BuiltinNames(), sortedAddonNames()...)
	for _, name := range names {
		ctor, ok := pluginhost.Builtin(name)
		if !ok {
			continue
		}
		info, err := ctor(version).Describe(protocol.InitializeParams{})
		if err != nil {
			return Result{Err: err}
		}
		row := Row{
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
		return Result{Err: err}
	}
	for _, p := range installed {
		rows = append(rows, Row{Name: p.Name, Version: p.Version, Source: p.Dir, Enabled: true, UsedBy: used[p.Name]})
	}
	// Plugins a project relies on come first, so a newcomer sees what matters.
	sort.SliceStable(rows, func(i, j int) bool { return len(rows[i].UsedBy) > 0 && len(rows[j].UsedBy) == 0 })
	return Result{Rows: rows}
}

func sortedAddonNames() []string {
	var out []string
	for name := range pluginhost.BuiltinAddons {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
