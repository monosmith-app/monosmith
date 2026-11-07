package commands

import (
	"strings"

	"github.com/monosmith-app/monosmith/libs/words/words"
)

// RunProject is the project the run keys b, t, l and c act on: the one selected
// in the list, on every view, so one key has one meaning. The capital keys B, T,
// L and C always cover every project, and every label names its scope.
func (s State) RunProject() string { return s.Project }

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
		return verb + words.Say("commands.scope.affected")
	case projects == 1:
		return verb + words.Say("commands.scope.only")
	case projects > 0:
		return words.Say("commands.scope.count", "verb", verb, "count", projects)
	}
	return verb + words.Say("commands.scope.all")
}
