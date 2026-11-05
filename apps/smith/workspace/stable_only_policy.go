package workspace

import (
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// StableOnlyRegistries are the registries stable_only can name. npm is not
// one: a pre-release goes to the channel's dist-tag there, and that tag is how
// a channel is installed.
var StableOnlyRegistries = []string{"pypi", "nuget", "rubygems"}

// SkipsPrerelease reports whether the release leaves a pre-release version
// out of the registry ("pypi", "nuget" or "rubygems") because stable_only
// names it. A stable version is never skipped.
//
// registry is the key of the [projects.<app>.release] table; version is the
// semantic version about to be published. A nil release skips nothing.
func (r *ProjectRelease) SkipsPrerelease(registry, version string) bool {
	return r != nil && strings.Contains(version, "-") && slices.Contains(r.StableOnly, registry)
}

// validateStableOnly checks stable_only of [projects.<name>.release]: each
// entry names a registry that takes it, once, and the table publishes there.
func validateStableOnly(table string, r *ProjectRelease) error {
	published := map[string]bool{"pypi": r.PyPI != "", "nuget": r.NuGet != "", "rubygems": r.RubyGems != ""}
	seen := map[string]bool{}
	for _, name := range r.StableOnly {
		switch {
		case !slices.Contains(StableOnlyRegistries, name):
			return problem.New(fmt.Sprintf("%s stable_only has %q, which cannot skip pre-releases", table, name),
				"use "+strings.Join(StableOnlyRegistries, ", ")+" (npm keeps pre-releases under the channel's tag)")
		case seen[name]:
			return problem.New(fmt.Sprintf("%s stable_only lists %q twice", table, name), "list each registry once")
		case !published[name]:
			return problem.New(fmt.Sprintf("%s stable_only names %s, but the table sets no %s package", table, name, name),
				fmt.Sprintf("add %s = \"<package name>\", or remove %s from stable_only", name, name))
		}
		seen[name] = true
	}
	return nil
}
