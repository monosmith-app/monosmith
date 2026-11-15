package pluginenv

import (
	"sort"
	"strings"
)

// Build returns the environment for an external plugin, as KEY=value entries:
// the baseline, the variables its manifest declares (never smith's own
// secrets, even by a pattern) and the variables smith.toml grants (anything).
// A declared name is exact or a prefix pattern ending in *. Names compare
// without regard to case on Windows, and the baseline always does.
func Build(parent []string, manifest, granted []string, goos string) []string {
	baseline := Baseline(goos)
	fold := goos == "windows"
	var out []string
	for _, entry := range parent {
		name, _, ok := strings.Cut(entry, "=")
		if !ok || name == "" {
			continue
		}
		switch {
		case matchesAny(baseline, name, true):
		case matchesAny(granted, name, fold):
		case !IsSmithSecret(name) && matchesAny(manifest, name, fold):
		default:
			continue
		}
		out = append(out, entry)
	}
	return out
}

// Visible lists, sorted and without values, the names and patterns a plugin
// may see beyond the baseline: what its manifest declares (smith's own secrets
// excepted, which a manifest cannot ask for) and what smith.toml grants.
func Visible(manifest, granted []string) []string {
	seen := map[string]bool{}
	for _, n := range manifest {
		if n != "" && !IsSmithSecret(n) {
			seen[n] = true
		}
	}
	for _, n := range granted {
		if n != "" {
			seen[n] = true
		}
	}
	out := make([]string, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func matchesAny(patterns []string, name string, fold bool) bool {
	for _, p := range patterns {
		if matches(p, name, fold) {
			return true
		}
	}
	return false
}

// matches is an exact name, or a prefix when the pattern ends in *.
func matches(pattern, name string, fold bool) bool {
	if fold {
		pattern, name = strings.ToUpper(pattern), strings.ToUpper(name)
	}
	if prefix, ok := strings.CutSuffix(pattern, "*"); ok {
		return strings.HasPrefix(name, prefix)
	}
	return pattern == name
}
