package guide

import (
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// addSetupKey is the key of the first question after the name of a new project.
const addSetupKey = "add-setup"

// addSetupStep is the first question after the name of a project: take the
// defaults and get to the plan, or answer each question. It uses the same two
// answers as the setup of a new workspace (setupDefaults, setupCustom). The
// defaults are what the stack suggests, so the plan can show every one of
// them as a flag.
func addSetupStep() Step {
	return Step{
		Key: addSetupKey, Prompt: words.Say("guide.add_project_setup.how_do_you_want_to_set"),
		Note: words.Say("guide.add_project_setup.note"),
		Options: []Option{
			{Label: words.Say("guide.add_project_setup.start_with_the_defaults"), Value: setupDefaults, Hint: words.Say("guide.add_project_setup.go_to_the_plan_now_each")},
			{Label: words.Say("guide.add_project_setup.customise"), Value: setupCustom, Hint: words.Say("guide.add_project_setup.answer_the_questions_one_at_a")},
		},
	}
}
