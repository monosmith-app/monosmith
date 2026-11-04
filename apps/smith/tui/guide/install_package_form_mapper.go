package guide

import (
	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
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
	Label: "Workspace root: shared by every project, in the root package.json",
	Value: ".",
}

// InstallPackageTitle is the title of the form that installs packages.
const InstallPackageTitle = "Install a package"

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
			return n + " is an option, not a package: the next question sets the kind"
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
	projects := Step{Key: "projects", Prompt: "Install into which projects? (space ticks, enter confirms). Each project uses its own package manager, so tick projects of one language when the name is the same in every registry.", Multi: true,
		Ticked: func(map[string][]string) []string { return preset }}
	for _, t := range targets {
		projects.Options = append(projects.Options, Option{Label: t.Label, Value: t.Value})
	}
	return Action{
		Title: InstallPackageTitle, Hint: "smith install -p <project> <package>", NeedsWorkspace: true, Presents: "install",
		Steps: []Step{
			{Key: "names", Prompt: "Which packages? Type a name or name@version; separate several with spaces.", Placeholder: "zod  react@19",
				Note:  "Installs into projects, or the workspace root where it is offered. A tool installed on the machine (npm -g) has no smith command yet.",
				Check: installNamesProblem},
			projects,
			{Key: "kind", Prompt: "As which kind of dependency?", Options: []Option{
				{Label: "Production", Value: "prod", Hint: "the project needs it when it runs (the default)"},
				{Label: "Development", Value: "dev", Hint: "only for building and testing (smith install -D). Peer and optional have no flag yet."},
			}},
		},
		Argv: InstallPackageArgv,
	}
}
