package guide

import (
	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"

	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// RemoveTarget is a project a package can be removed from: the label the form
// shows and the value smith uninstall -p takes.
type RemoveTarget struct {
	Label, Value string
}

// RemovePackageTitle is the title of the form that removes one package.
func RemovePackageTitle(name string) string {
	return words.Say("guide.remove_package.title", "name", name)
}

// RemovePackageArgv turns the projects picked into `smith uninstall -p <project>
// ... <package>`. It returns nil while no project is picked, so nothing runs.
func RemovePackageArgv(name string, answers map[string][]string) []string {
	argv := []string{"uninstall"}
	for _, v := range answers["projects"] {
		if v != "" {
			argv = append(argv, flagvocabulary.Arg("project"), v)
		}
	}
	if len(argv) == 1 {
		return nil
	}
	return append(argv, name)
}

// RemovePackageAction is the form that asks which of the projects that declare
// a package it is removed from. Several can be picked (space toggles); nothing is
// picked for the person.
//
// Parameters: name is the package; targets are the projects that declare it.
func RemovePackageAction(name string, targets []RemoveTarget) Action {
	step := Step{Key: "projects", Prompt: words.Say("guide.remove_package.which", "name", name), Multi: true}
	for _, t := range targets {
		step.Options = append(step.Options, Option{Label: t.Label, Value: t.Value})
	}
	return Action{
		Title: RemovePackageTitle(name), Hint: words.Say("guide.remove_package.hint", "name", name), NeedsWorkspace: true, Presents: "uninstall",
		Steps: []Step{step},
		Argv:  func(answers map[string][]string) []string { return RemovePackageArgv(name, answers) },
	}
}
