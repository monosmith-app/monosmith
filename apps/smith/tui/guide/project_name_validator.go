package guide

import (
	"github.com/monosmith-app/monosmith/apps/smith/scaffold"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// checkProjectName is the live check of the name step of Add a project: the
// rule is the one `smith add` enforces, said in plain words.
func checkProjectName(answers map[string][]string, name string) string {
	kind := first(answers, "kind")
	if scaffold.ValidateName(kind, name) == nil {
		return ""
	}
	return words.Say("guide.project_name.example", "example", exampleName(kind))
}

// checkWorkspaceName is the live check of the name step of Create a workspace:
// the rule and the words are those of Add a project, so a bad name is refused
// on the spot and not after the last question.
func checkWorkspaceName(_ map[string][]string, name string) string {
	if scaffold.ValidateName("workspace", name) == nil {
		return ""
	}
	return words.Say("guide.project_name.example", "example", exampleName("workspace"))
}

// nameRule explains why a name has the shape smith asks for.
func nameRule() string { return words.Say("guide.project_name.rule") }
