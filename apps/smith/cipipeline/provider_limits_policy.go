package cipipeline

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// Limits are what a configured provider cannot express of the workspace's
// settings, in plain words, so `smith ci generate` says so instead of
// leaving the file to differ silently from what the team asked for.
func Limits(cfg workspace.Config, m Model) []string {
	var out []string
	if slices.Contains(cfg.CI.Providers, "circleci") {
		if len(cfg.CI.Branches) > 0 {
			out = append(out, "CircleCI: [ci] branches does not apply; CircleCI builds every branch, and a release still runs only from "+cfg.Release.Branch)
		}
		if m.Native {
			var other []string
			for _, r := range m.NativeRunners {
				if !isLinuxRunner(r) {
					other = append(other, r)
				}
			}
			if len(other) > 0 {
				out = append(out, "CircleCI: native projects build on Linux only, not on "+strings.Join(other, ", ")+"; add a macOS or Windows job that runs `smith ci native` yourself")
			}
		}
	}
	return out
}
