package workspace

import (
	"fmt"
	"strings"
)

// validateGroups checks [[groups]]: a group names the projects it is
// about, with selectors that mean something, and sets at least one task, tag
// or release.
func validateGroups(cfg *Config) error {
	for i, g := range cfg.Groups {
		at := fmt.Sprintf("%s: [[groups]] number %d", FileName, i+1)
		if len(g.Select) == 0 {
			return fmt.Errorf("%s has no select; name the projects it is about (a name, a glob, kind:lib, lang:node or tag:ui)", at)
		}
		for _, selector := range g.Select {
			if err := validateSelector(selector); err != nil {
				return fmt.Errorf("%s: %w", at, err)
			}
		}
		if len(g.Tasks) == 0 && len(g.Tags) == 0 && g.Release == nil {
			return fmt.Errorf("%s (select %s) sets nothing; add tags, [groups.release] or [groups.tasks.<task>]", at, strings.Join(g.Select, ", "))
		}
		for _, tag := range g.Tags {
			if strings.TrimSpace(tag) == "" {
				return fmt.Errorf("%s (select %s) lists an empty tag", at, strings.Join(g.Select, ", "))
			}
		}
		if r := g.Release; r != nil && (len(r.NPM) > 0 || r.PyPI != "" || r.NuGet != "" || r.RubyGems != "" || r.Pub != "" || r.PubTag != "") {
			return fmt.Errorf("%s (select %s): [groups.release] cannot name the package (npm, pypi, nuget, rubygems, pub); a package name belongs to one project, so set it in [projects.<name>.release]", at, strings.Join(g.Select, ", "))
		}
	}
	return nil
}

func validateSelector(selector string) error {
	s := strings.TrimSpace(selector)
	if s == "" {
		return fmt.Errorf("select lists an empty selector")
	}
	if strings.HasPrefix(s, "!") {
		return fmt.Errorf("select %q: a group lists what it matches, so leave out the !; every selector must match", selector)
	}
	if kind, ok := strings.CutPrefix(s, "kind:"); ok && kind != "app" && kind != "lib" && kind != "package" {
		return fmt.Errorf("select %q: kind must be app, lib or package", selector)
	}
	for _, prefix := range []string{"lang:", "tag:"} {
		if v, ok := strings.CutPrefix(s, prefix); ok && strings.TrimSpace(v) == "" {
			return fmt.Errorf("select %q has nothing after %s", selector, prefix)
		}
	}
	return nil
}
