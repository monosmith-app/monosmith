package pluginmarket

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/semver"
)

// Pick chooses the release of a plugin: the version asked for (a leading v is
// ignored), or the newest one when version is empty. The newest ignores
// pre-releases (1.0.0-beta.1) unless the plugin lists nothing else.
func Pick(idx Index, name, version string) (Plugin, Release, error) {
	p, ok := idx.Plugins[name]
	if !ok {
		return Plugin{}, Release{}, fmt.Errorf("no plugin %q in the marketplace (try `smith plugin search`)", name)
	}
	if version != "" {
		r, ok := p.Versions[strings.TrimPrefix(version, "v")]
		if !ok {
			return Plugin{}, Release{}, fmt.Errorf("%s has no version %s; it has %s", name, version, strings.Join(Versions(p), ", "))
		}
		return p, r, nil
	}
	return p, p.Versions[Newest(p)], nil
}

// Newest is the version Pick chooses when none is asked for: the highest that
// is not a pre-release, or the highest of all when every version is one.
func Newest(p Plugin) string {
	all := Versions(p)
	for _, v := range all {
		if !strings.Contains(v, "-") {
			return v
		}
	}
	return all[0]
}

// Versions lists the versions of a plugin, newest first.
func Versions(p Plugin) []string {
	out := make([]string, 0, len(p.Versions))
	for v := range p.Versions {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return semver.Compare(out[i], out[j]) > 0 })
	return out
}
