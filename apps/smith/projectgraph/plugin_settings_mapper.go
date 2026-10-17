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
	if conv.Push != nil {
		s["push"] = *conv.Push
	}
	if d := po.Docker; d != nil {
		if d.Image != "" {
			s["image"] = d.Image
		}
		if d.Push != nil {
			s["push"] = *d.Push
		}
		if d.Context != "" {
			s["context"] = strings.TrimSuffix(d.Context, "/")
		}
	}
	if len(s) == 0 {
		return nil
	}
	return s
}
