package guide

// RemoveTarget is a project a package can be removed from: the label the form
// shows and the value smith uninstall -w takes.
type RemoveTarget struct {
	Label, Value string
}

// RemovePackageTitle is the title of the form that removes one package.
func RemovePackageTitle(name string) string { return "Remove " + name }

// RemovePackageArgv turns the projects picked into `smith uninstall -w <project>
// ... <package>`. It returns nil while no project is picked, so nothing runs.
func RemovePackageArgv(name string, answers map[string][]string) []string {
	argv := []string{"uninstall"}
	for _, v := range answers["projects"] {
		if v != "" {
			argv = append(argv, "-w", v)
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
	step := Step{Key: "projects", Prompt: name + " is declared in more than one project. Remove it from which? (space toggles, enter confirms)", Multi: true}
	for _, t := range targets {
		step.Options = append(step.Options, Option{Label: t.Label, Value: t.Value})
	}
	return Action{
		Title: RemovePackageTitle(name), Hint: "smith uninstall -w <project> " + name, NeedsWorkspace: true, Presents: "uninstall",
		Steps: []Step{step},
		Argv:  func(answers map[string][]string) []string { return RemovePackageArgv(name, answers) },
	}
}
