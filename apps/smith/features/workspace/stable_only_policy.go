package workspace

import (
	"slices"
	"strings"
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
