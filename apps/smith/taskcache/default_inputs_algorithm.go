package taskcache

import (
	"slices"
	"strings"
)

// defaultsPrefix is where plugins materialise the defaults they pass to tools.
// Git ignores it, so the file list never holds these files.
const defaultsPrefix = ".smith/defaults/"

// DefaultInputs returns the workspace-relative files under .smith/defaults/ that
// globs name outright, which SelectInputs cannot find because they are ignored
// by git, and that selected does not already hold. Their content joins the key.
func DefaultInputs(selected, globs []string) []string {
	var out []string
	for _, g := range globs {
		if !strings.HasPrefix(g, WorkspacePrefix) {
			continue
		}
		g = strings.TrimPrefix(g, WorkspacePrefix)
		if strings.HasPrefix(g, defaultsPrefix) &&!strings.ContainsAny(g, "*?[{!") && !slices.Contains(selected, g) && !slices.Contains(out, g) {
			out = append(out, g)
		}
	}
	return out
}
