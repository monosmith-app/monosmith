package guide

import (
	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
	"github.com/monosmith-app/monosmith/libs/words/words"
	"strings"
)

// InstallTarget is a project a package can be installed into: the label the form
// shows and the value smith install -p takes.
type InstallTarget struct {
	Label, Value string
}

// RootInstallTarget is the target that stands for the workspace root: the value
// smith install -p takes for the package.json at the top of the workspace.
var RootInstallTarget = InstallTarget{
	Label: words.Say("guide.install_package.root"),
	Value: ".",
}

// InstallPackageTitle is the title of the form that installs packages.
var InstallPackageTitle = words.Say("guide.install_package.title")

// InstallPackageArgv turns the answers into `smith install -p <project> ... [-D]
// <package> ...`. It returns nil while no package or no project is given, so
// nothing runs.
//
// Parameters: answers are the answers of the form: names, projects and kind.
func InstallPackageArgv(answers map[string][]string) []string {
	names := strings.Fields(first(answers, "names"))
	argv := []string{"install"}
	for _, v := range answers["projects"] {
		if v != "" {
			argv = append(argv, flagvocabulary.Arg("project"), v)
		}
	}
	if len(names) == 0 || len(argv) == 1 {
		return nil
	}
	if first(answers, "kind") == "dev" {
		argv = append(argv, "-D")
	}
	return append(argv, names...)
}

// installNamesProblem says what is wrong with the names typed ("" when fine).
func installNamesProblem(_ map[string][]string, value string) string {
	for _, n := range strings.Fields(value) {
		if strings.HasPrefix(n, "-") {
			return words.Say("guide.install_package.option", "name", n)
		}
	}
	return ""
}

// InstallPackageAction is the form that installs packages: which names (a name
// or name@version, one or several), into which projects (several can be ticked;
// preset ones start ticked and are never final: enter confirms), and as which
// kind. smith install knows two kinds, production and development; peer and
// optional dependencies have no flag, so they are not offered. A machine-wide install
// has no command, and the first question says so.
//
// Parameters: targets are every project, with its language in the label so a
// project with several parts (a Tauri app has a Rust part and a web part) shows
// which part receives the package; the caller puts RootInstallTarget first only
// when the workspace root can take packages; preset are the values ticked at the
// start.
func InstallPackageAction(targets []InstallTarget, preset []string) Action {
	projects := Step{Key: "projects", Prompt: words.Say("guide.install_package_form.install_into_which_projects_space_ticks"), Multi: true,
		Ticked: func(map[string][]string) []string { return preset }}
	for _, t := range targets {
		projects.Options = append(projects.Options, Option{Label: t.Label, Value: t.Value})
	}
	return Action{
		Title: InstallPackageTitle, Hint: words.Say("guide.install_package_form.smith_install_p_project_package"), NeedsWorkspace: true, Presents: "install",
		Steps: []Step{
			{Key: "names", Prompt: words.Say("guide.install_package_form.which_packages_type_a_name_or"), Placeholder: words.Say("guide.install_package_form.zod_react_19"),
				Note:  words.Say("guide.install_package_form.installs_into_projects_or_the_workspace"),
				Check: installNamesProblem},
			projects,
			{Key: "kind", Prompt: words.Say("guide.install_package_form.as_which_kind_of_dependency"), Options: []Option{
				{Label: words.Say("guide.install_package_form.production"), Value: "prod", Hint: words.Say("guide.install_package_form.the_project_needs_it_when_it")},
				{Label: words.Say("guide.install_package_form.development"), Value: "dev", Hint: words.Say("guide.install_package_form.only_for_building_and_testing_smith")},
			}},
		},
		Argv: InstallPackageArgv,
	}
}
