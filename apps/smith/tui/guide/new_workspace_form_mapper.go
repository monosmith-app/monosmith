package guide

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/scaffold"
)

// NewWorkspaceSteps are the questions of "Create a workspace": the name, then
// the questions smith new asks at a terminal (scaffold.NewQuestions), so the
// form and the command ask the same. The first option of each is the default
// and a text question left empty takes its default.
func NewWorkspaceSteps() []Step {
	steps := []Step{{Key: "name", Prompt: "What should the new workspace be called? It becomes a folder (lowercase words joined by dashes)", Placeholder: "my-studio"}}
	for _, q := range scaffold.NewQuestions() {
		step := Step{Key: q.Key, Prompt: q.Prompt}
		if q.Advanced {
			step.Prompt += " Enter keeps " + q.Default + "."
		}
		if q.Options != nil {
			for _, o := range q.Options {
				step.Options = append(step.Options, Option{Label: o.Label, Value: o.Value, Hint: o.Hint})
			}
		} else {
			step.Placeholder, step.Optional = q.Default, true
			if check := q.Check; check != nil {
				step.Check = func(_ map[string][]string, value string) string {
					if err := check(value); err != nil {
						return err.Error()
					}
					return ""
				}
			}
		}
		if applies := q.Applies; applies != nil {
			step.Skip = func(a map[string][]string) bool { return !applies(singleAnswers(a)) }
		}
		steps = append(steps, step)
	}
	return steps
}

// singleAnswers is the first answer of each question.
func singleAnswers(a map[string][]string) map[string]string {
	out := make(map[string]string, len(a))
	for k := range a {
		out[k] = first(a, k)
	}
	return out
}

// NewWorkspaceArgv is the command the answers build. It passes --yes, so the
// command asks nothing the form already asked, and a flag for every answer
// that was given.
func NewWorkspaceArgv(a map[string][]string) []string {
	argv := []string{"new", first(a, "name"), "--yes"}
	for _, q := range scaffold.NewQuestions() {
		value := strings.TrimSpace(first(a, q.Key))
		if q.Flag == "" || value == "" || value == q.Default || q.Key == scaffold.AnswerRecipeName {
			continue
		}
		switch q.Key {
		case scaffold.AnswerNodeStack:
			argv = append(argv, "--"+q.Flag, "node="+value)
		case scaffold.AnswerGitHubReleases:
			argv = append(argv, "--"+q.Flag)
		default:
			argv = append(argv, "--"+q.Flag, value)
		}
	}
	if first(a, scaffold.AnswerSaveRecipe) == "yes" {
		name := strings.TrimSpace(first(a, scaffold.AnswerRecipeName))
		if name == "" {
			name = first(a, "name")
		}
		argv = append(argv, "--save-recipe", name)
	}
	return argv
}
