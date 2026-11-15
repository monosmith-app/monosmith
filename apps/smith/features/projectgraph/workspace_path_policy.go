package projectgraph

import (
	"fmt"
	"path"
	"strings"
)

// CleanRelative puts a relative path from a plugin or smith.toml in smith's
// form: forward slashes, no "." or ".." segments, "." for the base itself.
// Paths stay POSIX inside smith (cache keys, graphs, messages) and keep the
// case they were given, which for discovered projects is the case git reports.
// An absolute path, or one that climbs out of its base, is an error.
func CleanRelative(p string) (string, error) {
	slashed := strings.ReplaceAll(p, `\`, "/")
	if strings.HasPrefix(slashed, "/") || (len(slashed) >= 2 && slashed[1] == ':') {
		return "", fmt.Errorf("%q is absolute; give a path relative to the workspace", p)
	}
	clean := path.Clean(slashed)
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("%q leaves the workspace", p)
	}
	return clean, nil
}

// normalisePaths cleans a project's root and its tasks' working directories,
// and turns backslashes in input and output globs into slashes.
func normalisePaths(p *Project) error {
	root, err := CleanRelative(p.Root)
	if err != nil {
		return fmt.Errorf("project %s: root %w", p.Name, err)
	}
	p.Root = root
	for name, t := range p.Tasks {
		if t.Cwd != "" {
			if t.Cwd, err = CleanRelative(t.Cwd); err != nil {
				return fmt.Errorf("%s:%s: cwd %w", p.Name, name, err)
			}
		}
		t.Inputs = slashed(t.Inputs)
		t.Outputs = slashed(t.Outputs)
		p.Tasks[name] = t
	}
	return nil
}

func slashed(globs []string) []string {
	if globs == nil {
		return nil
	}
	out := make([]string, len(globs))
	for i, g := range globs {
		out[i] = strings.ReplaceAll(g, `\`, "/")
	}
	return out
}
