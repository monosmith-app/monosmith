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
	for _, p := range []struct{ id, name, builds string }{
		{"circleci", "CircleCI", "CircleCI builds every branch"},
		{"bitbucket", "Bitbucket", "Bitbucket builds pull requests and pushes to " + cfg.Release.Branch + " only"},
	} {
		if !slices.Contains(cfg.CI.Providers, p.id) {
			continue
		}
		if len(cfg.CI.Branches) > 0 {
			out = append(out, p.name+": [ci] branches does not apply; "+p.builds+", and a release still runs only from "+cfg.Release.Branch)
		}
		if m.Native {
			var other []string
			for _, r := range m.NativeRunners {
				if !isLinuxRunner(r) {
					other = append(other, r)
				}
			}
			if len(other) > 0 {
				out = append(out, p.name+": native projects build on Linux only, not on "+strings.Join(other, ", ")+"; add a macOS or Windows job that runs `smith ci native` yourself")
			}
		}
	}
	if len(m.AssetPlatforms) > 0 {
		for _, p := range cfg.CI.Providers {
			if p != "github" {
				out = append(out, p+": the release files a [release] build command makes are built and uploaded by the GitHub Actions workflow only; add github to [ci] providers, or run `smith release assets` on each platform yourself")
			}
		}
	}
	return out
}
