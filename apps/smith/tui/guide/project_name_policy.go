package guide

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/scaffold"
)

// checkProjectName is the live check of the name step of Add a project: the
// rule is the one `smith add` enforces, said in plain words.
func checkProjectName(answers map[string][]string, name string) string {
	kind := first(answers, "kind")
	if scaffold.ValidateName(kind, name) == nil {
		return ""
	}
	return fmt.Sprintf("use lowercase words joined by dashes, like %s", exampleName(kind))
}

// checkWorkspaceName is the live check of the name step of Create a workspace:
// the rule and the words are those of Add a project, so a bad name is refused
// on the spot and not after the last question.
func checkWorkspaceName(_ map[string][]string, name string) string {
	if scaffold.ValidateName("workspace", name) == nil {
		return ""
	}
	return fmt.Sprintf("use lowercase words joined by dashes, like %s", exampleName("workspace"))
}

// namePlaceholder suggests a name that suits what is being added.
func namePlaceholder(answers map[string][]string) string { return exampleName(first(answers, "kind")) }

func exampleName(kind string) string {
	switch kind {
	case "lib":
		return "shared-code"
	case "package":
		return "my-package"
	case "workspace":
		return "my-studio"
	}
	return "web"
}
