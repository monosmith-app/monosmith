package projectgraph

import (
	"slices"
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// declaredFor lists the projects smith.toml declares ([projects.<name>] with a
// root and a language) for the plugin that owns their language: the plugin's
// name or one of the languages it answers to. The list is sorted by name, so a
// run is repeatable.
func declaredFor(plugin string, languages []string, cfg workspace.Config) []protocol.DeclaredProject {
	var out []protocol.DeclaredProject
	for name, po := range cfg.Projects {
		if !po.Declared() {
			continue
		}
		if po.Language == plugin || slices.Contains(languages, po.Language) {
			out = append(out, protocol.DeclaredProject{Name: name, Root: po.Root, Kind: protocol.ProjectKind(po.Kind)})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// discoverSettings collects, by project name, the settings smith.toml gives
// the plugin ([projects.<name>.settings.<plugin>]), for the plugin to read
// while it describes the projects. Nil when no project has any.
func discoverSettings(plugin string, cfg workspace.Config) map[string]map[string]any {
	var out map[string]map[string]any
	for name, po := range cfg.Projects {
		if s := ProjectSettings(plugin, po); s != nil {
			if out == nil {
				out = map[string]map[string]any{}
			}
			out[name] = s
		}
	}
	return out
}
