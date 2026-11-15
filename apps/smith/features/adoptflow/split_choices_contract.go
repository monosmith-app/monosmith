package adoptflow

import "strings"

// Split is what the person decided about the split of workspace and project
// when the project at the root moves (ADR 0090): which development
// dependencies smith could not place go to the workspace, and what happens to
// the project's own ESLint setup. The zero value decides nothing, and the
// step then takes its safe defaults.
type Split struct {
	// Hoist are the development dependencies sent to the workspace manifest.
	Hoist []string
	// ESLint is "shared", "keep" or "" when the question was not asked.
	ESLint string
}

// Argv is the Split as the flags of smith adopt --init.
func (s Split) Argv() []string {
	var out []string
	if len(s.Hoist) > 0 {
		out = append(out, "--hoist", strings.Join(s.Hoist, ","))
	}
	if s.ESLint != "" {
		out = append(out, "--eslint-setup", s.ESLint)
	}
	return out
}
