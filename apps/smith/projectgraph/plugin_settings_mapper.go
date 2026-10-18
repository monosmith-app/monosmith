package projectgraph

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// PluginSettings is what a convention reads for one project: its
// [conventions.<name>] settings with the project's [projects.<project>.<name>]
// on top, or nil when smith.toml sets none. Only the docker convention has
// settings so far.
func PluginSettings(name string, conv workspace.ConventionConfig, po workspace.ProjectOverride) map[string]any {
	if name != "docker" {
		return nil
	}
	s := map[string]any{}
	if conv.Registry != "" {
		s["registry"] = conv.Registry
	}
	if conv.Registries != nil {
		s["registries"] = conv.Registries
	}
	if conv.DescriptionFrom != "" {
		s["description_from"] = conv.DescriptionFrom
	}
	if conv.Push != nil {
		s["push"] = *conv.Push
	}
	if conv.Verify != nil {
		s["verify"] = *conv.Verify
	}
	if d := po.Docker; d != nil {
		if d.Image != "" {
			s["image"] = d.Image
		}
		// The app's registries replace the convention's, in either form.
		if d.Registry != "" || d.Registries != nil {
			delete(s, "registry")
			delete(s, "registries")
		}
		if d.Registry != "" {
			s["registry"] = d.Registry
		}
		if d.Registries != nil {
			s["registries"] = d.Registries
		}
		if d.DescriptionFrom != "" {
			s["description_from"] = d.DescriptionFrom
		}
		if d.Push != nil {
			s["push"] = *d.Push
		}
		if d.Context != "" {
			s["context"] = strings.TrimSuffix(d.Context, "/")
		}
		if d.Verify != nil {
			s["verify"] = *d.Verify
		}
	}
	if len(s) == 0 {
		return nil
	}
	return s
}

// ProjectSettings is a plugin's settings for one project from smith.toml's
// [projects.<project>.settings.<plugin>]: what a recipe's settings became
// when it created the project. It is a copy, nil when there are none.
func ProjectSettings(plugin string, po workspace.ProjectOverride) map[string]any {
	table := po.Settings[plugin]
	if len(table) == 0 {
		return nil
	}
	out := make(map[string]any, len(table))
	for k, v := range table {
		out[k] = v
	}
	return out
}

// withProjectSettings puts a project's own settings for a plugin under the
// settings a convention computed from its [conventions.<name>] table and the
// project's [projects.<project>.<name>] table: those are more specific, so
// they win key by key.
func withProjectSettings(computed map[string]any, plugin string, po workspace.ProjectOverride) map[string]any {
	own := ProjectSettings(plugin, po)
	if own == nil {
		return computed
	}
	for k, v := range computed {
		own[k] = v
	}
	return own
}
