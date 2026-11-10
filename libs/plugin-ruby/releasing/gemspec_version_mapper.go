package releasing

import (
	"regexp"
	"strings"
)

var versionField = regexp.MustCompile(`(\.version\s*=\s*)(['"])[^'"]*['"]`)

// GemVersion is version as RubyGems spells it: a semver prerelease
// 2.0.0-beta.1 is 2.0.0.pre.beta.1 (RubyGems reads a dash as ".pre."), the
// spelling that orders it below 2.0.0 and that the built gem's file name has.
// A release stays as it is, and build metadata is dropped, which RubyGems
// does not allow.
func GemVersion(version string) string {
	version, _, _ = strings.Cut(version, "+")
	core, pre, ok := strings.Cut(version, "-")
	if !ok {
		return version
	}
	return core + ".pre." + pre
}

// SetVersion writes version, spelled the way RubyGems does (see GemVersion),
// into the gemspec's version assignment, keeping its quotes; a gemspec that
// computes its version is left alone.
func SetVersion(src, version string) (string, bool) {
	version = GemVersion(version)
	loc := versionField.FindStringSubmatchIndex(src)
	if loc == nil {
		return src, false
	}
	quote := src[loc[4]:loc[5]]
	out := src[:loc[0]] + src[loc[2]:loc[3]] + quote + version + quote + src[loc[1]:]
	return out, out != src
}
