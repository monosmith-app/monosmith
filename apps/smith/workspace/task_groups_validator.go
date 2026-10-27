package workspace

import (
	"fmt"
	"strings"
)

// validateTaskGroups checks [[task_groups]]: a group names the projects it is
// about, with selectors that mean something, and sets at least one task.
func validateTaskGroups(cfg *Config) error {
	for i, g := range cfg.TaskGroups {
		at := fmt.Sprintf("%s: [[task_groups]] number %d", FileName, i+1)
		if len(g.Select) == 0 {
			return fmt.Errorf("%s has no select; name the projects it is about (a name, a glob, kind:lib, lang:node or tag:ui)", at)
		}
		for _, selector := range g.Select {
			if err := validateSelector(selector); err != nil {
				return fmt.Errorf("%s: %w", at, err)
			}
		}
		if len(g.Tasks) == 0 {
			return fmt.Errorf("%s (select %s) sets no task; add [task_groups.tasks.<task>]", at, strings.Join(g.Select, ", "))
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
