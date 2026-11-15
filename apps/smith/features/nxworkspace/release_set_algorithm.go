package nxworkspace

import (
	"encoding/json"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// ReleaseSet returns the names of the projects nx release publishes: the
// release.projects patterns (names, name globs, directory globs and
// "tag:<tag>", with "!" excluding), or, when there are none, every package
// whose package.json is not private.
func ReleaseSet(w *Workspace) map[string]bool {
	set := map[string]bool{}
	var patterns []string
	if w.Nx.Release != nil {
		patterns = w.Nx.Release.Projects
		for _, raw := range w.Nx.Release.Groups {
			var g struct {
				Projects StringList `json:"projects"`
			}
			if json.Unmarshal(raw, &g) == nil {
				patterns = append(patterns, g.Projects...)
			}
		}
	}
	if len(patterns) == 0 {
		// Nx versions every project then, but publishes only packages that
		// are not private.
		for _, p := range w.Projects {
			if p.Package != nil && !p.Package.Private {
				set[p.Name] = true
			}
		}
		return set
	}
	for _, include := range []bool{true, false} {
		for _, pattern := range patterns {
			negate := strings.HasPrefix(pattern, "!")
			if negate == include {
				continue
			}
			pattern = strings.TrimPrefix(pattern, "!")
			for _, p := range w.Projects {
				if Matches(p, pattern) {
					set[p.Name] = include
				}
			}
		}
	}
	for name, in := range set {
		if !in {
			delete(set, name)
		}
	}
	return set
}

// Matches applies one Nx project pattern to a project.
func Matches(p *Project, pattern string) bool {
	if tag, ok := strings.CutPrefix(pattern, "tag:"); ok {
		for _, t := range p.Tags {
			if ok, _ := doublestar.Match(tag, t); ok {
				return true
			}
		}
		return false
	}
	if ok, _ := doublestar.Match(pattern, p.Name); ok {
		return true
	}
	if strings.Contains(pattern, "/") {
		ok, _ := doublestar.Match(strings.TrimPrefix(pattern, "./"), p.Root)
		return ok
	}
	return false
}
