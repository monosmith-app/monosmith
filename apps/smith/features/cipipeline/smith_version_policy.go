package cipipeline

import "github.com/monosmith-app/monosmith/apps/smith/features/workspace"

// SmithVersionFor is the smith version the pipelines install: the one the
// workspace names in [workspace] smith, or else the smith that generates them.
func SmithVersionFor(cfg workspace.Config, running string) string {
	if cfg.Workspace.Smith != "" {
		return cfg.Workspace.Smith
	}
	return running
}
