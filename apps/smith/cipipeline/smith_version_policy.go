package cipipeline

import "github.com/monosmith-app/monosmith/apps/smith/workspace"

// SmithVersionFor is the smith version the pipelines install: the one the
// workspace names in [ci] smith_version, or else the smith that generates them.
func SmithVersionFor(cfg workspace.Config, running string) string {
	if cfg.CI.SmithVersion != "" {
		return cfg.CI.SmithVersion
	}
	return running
}
