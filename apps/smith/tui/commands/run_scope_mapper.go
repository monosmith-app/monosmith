package commands

import (
	"fmt"
	"strings"
)

// RunProject is the project the run keys act on in a state: the one selected
// in the list, but only on the Projects view, where the selection is in sight.
// Everywhere else (Home first of all) a run covers the whole workspace, so a
// selection the user never made there cannot narrow it.
func (s State) RunProject() string {
	if s.View != TabProjects {
		return ""
	}
	return s.Project
}

// RunScopeLabel names what a run command line covers, to be shown with the run:
// "Build · all 23 projects", "Test · web", "Lint · affected projects".
func RunScopeLabel(argv []string, projects int) string {
	req, ok := ParseRun(argv)
	if !ok {
		return ""
	}
	verb := strings.Join(req.Tasks, " + ")
	if len(verb) > 0 {
		verb = strings.ToUpper(verb[:1]) + verb[1:]
	}
	switch {
	case len(req.Projects) > 0:
		return verb + " · " + strings.Join(req.Projects, ", ")
	case req.Affected:
		return verb + " · affected projects"
	case projects > 0:
		return fmt.Sprintf("%s · all %d projects", verb, projects)
	}
	return verb + " · all projects"
}
