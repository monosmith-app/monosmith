package doctor

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// stableOnlyChecks reports, for a workspace with release channels, which apps
// publish every pre-release to PyPI, NuGet or RubyGems (a warning: those
// registries limit a project's total size, and a platform package set per
// pre-release fills it), and which apps keep those registries to stable
// versions with stable_only. A workspace with no channel publishes no
// pre-release, so nothing is reported.
func stableOnlyChecks(s *session.Session) []protocol.Finding {
	return stableOnlyFindings(s.Graph.Projects, len(s.Config.Release.Channels) > 0)
}

// stableOnlyFindings is stableOnlyChecks with its inputs given.
func stableOnlyFindings(all []*projectgraph.Project, channels bool) []protocol.Finding {
	if !channels {
		return nil
	}
	var out []protocol.Finding
	for _, p := range all {
		r := p.Release
		if r == nil {
			continue
		}
		var every, stable []string
		for key, name := range map[string]string{"pypi": r.PyPI, "nuget": r.NuGet, "rubygems": r.RubyGems} {
			switch {
			case name == "":
			case slices.Contains(r.StableOnly, key):
				stable = append(stable, key)
			default:
				every = append(every, key)
			}
		}
		sort.Strings(every)
		sort.Strings(stable)
		if len(every) > 0 {
			out = append(out, protocol.Finding{Check: "release: " + p.Name + " pre-releases and registry size", Warning: true, Project: p.Name,
				Problem: fmt.Sprintf("publishes every pre-release to %s", strings.Join(every, ", ")),
				Detail:  fmt.Sprintf("every pre-release of %s is published to %s, one package per platform each; these registries limit the total size of a project and refuse uploads when it is full", p.Name, strings.Join(every, ", ")),
				Remedy:  fmt.Sprintf("set stable_only = [%s] in [projects.%s.release] to send them stable versions only", quotedList(every), p.Name)})
		}
		if len(stable) > 0 {
			out = append(out, protocol.Finding{Check: "release: " + p.Name + " pre-releases and registry size", OK: true, Project: p.Name,
				Detail: fmt.Sprintf("%s receive stable versions only; a pre-release is reported as skipped: pre-release", strings.Join(stable, ", "))})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Check < out[j].Check })
	return out
}

// quotedList renders names as "a", "b".
func quotedList(names []string) string {
	q := make([]string, len(names))
	for i, n := range names {
		q[i] = `"` + n + `"`
	}
	return strings.Join(q, ", ")
}
