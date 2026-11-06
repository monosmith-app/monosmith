package guide

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/scaffold"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// The answers of the first question after the name of a new workspace.
const (
	answerSetup   = "setup"
	setupDefaults = "defaults"
	setupCustom   = "custom"
)

// defaultedQuestions are the questions the defaults settle, with the words a
// person knows them by and how to change the answer later.
func defaultedQuestions() []struct{ key, name, later string } {
	return []struct{ key, name, later string }{
		{scaffold.AnswerCI, words.Say("guide.new_workspace_setup.question.ci.name"), words.Say("guide.new_workspace_setup.question.ci.later")},
		{scaffold.AnswerLanguage, words.Say("guide.new_workspace_setup.question.language.name"), words.Say("guide.new_workspace_setup.question.language.later")},
		{scaffold.AnswerArchitecture, words.Say("guide.new_workspace_setup.question.layout.name"), words.Say("guide.new_workspace_setup.question.layout.later")},
		{scaffold.AnswerAgent, words.Say("guide.new_workspace_setup.question.agent.name"), words.Say("guide.new_workspace_setup.question.agent.later")},
		{scaffold.AnswerDevcontainer, words.Say("guide.new_workspace_setup.question.devcontainer.name"), words.Say("guide.new_workspace_setup.question.devcontainer.later")},
	}
}

// setupStep is the first question after the name: take smith's defaults and
// answer nothing more, or go through the questions. The defaults are read from
// the questions smith new asks, so the summary never differs from what is
// chosen. The two things that carry risk (what the first commit holds, secrets
// in the folder) are still asked in a folder that calls for them.
func setupStep() Step {
	return Step{
		Key: answerSetup, Prompt: words.Say("guide.new_workspace_setup.how_do_you_want_to_set"),
		Note: words.Say("guide.new_workspace_setup.note", "defaults", defaultsSummary()),
		Options: []Option{
			{Label: words.Say("guide.new_workspace_setup.start_with_the_defaults"), Value: setupDefaults, Hint: words.Say("guide.new_workspace_setup.the_name_is_the_only_answer")},
			{Label: words.Say("guide.new_workspace_setup.customise"), Value: setupCustom, Hint: words.Say("guide.new_workspace_setup.answer_the_questions_one_at_a")},
		},
	}
}

// defaultsSummary lists what the defaults choose, one line each, with how to
// change it later.
func defaultsSummary() string {
	var lines []string
	for _, d := range defaultedQuestions() {
		for _, q := range scaffold.NewQuestions() {
			if q.Key != d.key {
				continue
			}
			label := q.Default
			for _, o := range q.Options {
				if o.Value == q.Default {
					label = strings.ToLower(o.Label)
					if d.key == scaffold.AnswerCI {
						label = o.Label
					}
				}
			}
			lines = append(lines, words.Say("guide.new_workspace_setup.summary_line", "name", d.name, "label", label, "later", d.later))
		}
	}
	return strings.Join(lines, "\n")
}
