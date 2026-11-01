package tui

import (
	"slices"
	"strings"
)

// siblingsFollowedFrom is how many projects a workspace needs before smith add
// follows the ones that are like the new project.
const siblingsFollowedFrom = 3

// addPlanWhy is the sentence above the plan of an add: what the project is
// built from (the choices named with --with, or the stack's defaults) and
// whether it follows the other projects. projects is how many the workspace has.
func addPlanWhy(argv []string, projects int) string {
	var b strings.Builder
	b.WriteString("Nothing is written yet. ")
	if i := slices.Index(argv, "--with"); i >= 0 && i+1 < len(argv) {
		b.WriteString("It is built from your choices: " + strings.ReplaceAll(argv[i+1], ",", ", ") + ". ")
	} else {
		b.WriteString("It takes the stack's own defaults. ")
	}
	switch {
	case slices.Contains(argv, "--fresh"):
		b.WriteString("It does not follow the other projects. ")
	case projects >= siblingsFollowedFrom:
		b.WriteString("It follows the other projects of its kind: the plan names what it takes from them. ")
	}
	b.WriteString("Run makes the project and Esc leaves everything as it is. Tick the box to commit it, so it can be undone.")
	return b.String()
}
