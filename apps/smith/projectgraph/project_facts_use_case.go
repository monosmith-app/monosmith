package projectgraph

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/factmerge"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// configFactLayers are the facts smith.toml sets for a project: [facts] at
// the workspace scope and [projects.<name>.facts] at the project scope. It
// returns an error that names the table when a fact is invalid.
func configFactLayers(cfg workspace.Config, project string) ([]factmerge.Layer, error) {
	var layers []factmerge.Layer
	if len(cfg.Facts) > 0 {
		if err := protocol.ValidateFacts(cfg.Facts); err != nil {
			return nil, fmt.Errorf("smith.toml [facts]: %w", err)
		}
		layers = append(layers, factmerge.Layer{Scope: factmerge.ScopeWorkspace, Source: "smith.toml [facts]", Facts: cfg.Facts})
	}
	if own := cfg.Projects[project].Facts; len(own) > 0 {
		table := fmt.Sprintf("smith.toml [projects.%s.facts]", project)
		if err := protocol.ValidateFacts(own); err != nil {
			return nil, fmt.Errorf("%s: %w", table, err)
		}
		layers = append(layers, factmerge.Layer{Scope: factmerge.ScopeProject, Source: table, Facts: own})
	}
	return layers, nil
}

// ConfigFacts are the facts that hold for a project because smith.toml says
// so, merged: the values a plugin gets in a generate request, before the
// project exists and before any plugin has published a default. project is
// the name of the project. It returns nil when there are none and an error
// when a fact in smith.toml is invalid.
func ConfigFacts(cfg workspace.Config, project string) (map[string]any, error) {
	layers, err := configFactLayers(cfg, project)
	if err != nil {
		return nil, err
	}
	merged, err := factmerge.Merge(layers)
	if err != nil {
		return nil, err
	}
	return factmerge.Values(merged), nil
}

// SeedFacts gives every project the facts smith.toml sets, so the first
// requests to plugins (contribute) already carry them. It returns an error
// when a fact in smith.toml is invalid.
func SeedFacts(projects []*Project, cfg workspace.Config) error {
	for _, p := range projects {
		layers, err := configFactLayers(cfg, p.Name)
		if err != nil {
			return err
		}
		p.factLayers = layers
	}
	return ResolveFacts(projects)
}

// ResolveFacts merges, for each project, the smith.toml facts and the
// defaults plugins published so far into Facts and FactDetail. It runs again
// after each round of contributions, so later requests carry the defaults of
// the plugins asked earlier. It returns an error that names the project and
// both sources when two plugins publish different values for one key, or a
// plugin publishes an invalid fact.
func ResolveFacts(projects []*Project) error {
	for _, p := range projects {
		for _, l := range p.factLayers {
			if err := protocol.ValidateFacts(l.Facts); err != nil {
				return fmt.Errorf("project %s: %s: %w", p.Name, l.Source, err)
			}
		}
		merged, err := factmerge.Merge(p.factLayers)
		if err != nil {
			return fmt.Errorf("project %s: %w", p.Name, err)
		}
		p.FactDetail = merged
		p.Facts = factmerge.Values(merged)
	}
	return nil
}
