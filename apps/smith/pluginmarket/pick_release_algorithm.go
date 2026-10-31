package pluginmarket

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
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
	sort.Slice(out, func(i, j int) bool { return compareVersions(out[i], out[j]) > 0 })
	return out
}

// compareVersions orders dotted numbers; a pre-release sorts before the
// release it leads to, and a part that is not a number sorts by text.
func compareVersions(a, b string) int {
	ca, pa, _ := strings.Cut(strings.SplitN(a, "+", 2)[0], "-")
	cb, pb, _ := strings.Cut(strings.SplitN(b, "+", 2)[0], "-")
	xa, xb := strings.Split(ca, "."), strings.Split(cb, ".")
	for i := 0; i < len(xa) || i < len(xb); i++ {
		var na, nb int
		var ea, eb error
		if i < len(xa) {
			na, ea = strconv.Atoi(xa[i])
		}
		if i < len(xb) {
			nb, eb = strconv.Atoi(xb[i])
		}
		if ea != nil || eb != nil {
			return strings.Compare(a, b)
		}
		if na != nb {
			if na < nb {
				return -1
			}
			return 1
		}
	}
	switch {
	case pa == pb:
		return 0
	case pa == "":
		return 1
	case pb == "":
		return -1
	}
	return strings.Compare(pa, pb)
}
