package projectgraph

import (
	"fmt"
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// ConventionTargets picks the projects a convention applies to: those of the
// languages it declares (or the subset smith.toml names) that are not
// excluded. A convention that declares no languages applies to all of them. A
// language the convention does not declare is an error, so a typo cannot
// silently switch it off.
func ConventionTargets(name string, projects []*Project, supported []string, cfg workspace.ConventionConfig) ([]*Project, error) {
	languages := supported
	if len(cfg.Languages) > 0 {
		for _, l := range cfg.Languages {
			if !CoversLanguage(supported, l) {
				return nil, fmt.Errorf("smith.toml [conventions.%s]: it does not apply to %q; it supports %v", name, l, supported)
			}
		}
		languages = cfg.Languages
	}
	var out []*Project
	for _, p := range projects {
		if CoversLanguage(languages, p.Language) && !slices.Contains(cfg.Exclude, p.Name) {
			out = append(out, p)
		}
	}
	return out, nil
}
