package guide

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/scaffold"
)

// The answers of the first question after the name of a new workspace.
const (
	answerSetup   = "setup"
	setupDefaults = "defaults"
	setupCustom   = "custom"
)

// defaultedQuestions are the questions the defaults settle, with the words a
// person knows them by and how to change the answer later.
var defaultedQuestions = []struct{ key, name, later string }{
	{scaffold.AnswerCI, "CI", "[ci] providers in smith.toml, then smith ci generate"},
	{scaffold.AnswerLanguage, "language", "smith add --lang <language> for each project"},
	{scaffold.AnswerArchitecture, "code layout", "smith architecture"},
	{scaffold.AnswerAgent, "agent files", "smith agent install adds them"},
	{scaffold.AnswerDevcontainer, "dev container", "[ci] devcontainer in smith.toml"},
}

// setupStep is the first question after the name: take smith's defaults and
// answer nothing more, or go through the questions. The defaults are read from
// the questions smith new asks, so the summary never differs from what is
// chosen. The two things that carry risk (what the first commit holds, secrets
// in the folder) are still asked in a folder that calls for them.
func setupStep() Step {
	return Step{
		Key: answerSetup, Prompt: "How do you want to set it up?",
		Note: "CI = where your tests run on every push.\n" + defaultsSummary() +
			"\nEvery one can be changed later. If the folder has files and no git repository, smith still asks what the first commit holds.",
		Options: []Option{
			{Label: "Start with the defaults", Value: setupDefaults, Hint: "the name is the only answer; the list above is what you get (default)"},
			{Label: "Customise", Value: setupCustom, Hint: "answer the questions one at a time: CI, language, code layout, agent files, dev container, more options"},
		},
	}
}

// defaultsSummary lists what the defaults choose, one line each, with how to
// change it later.
func defaultsSummary() string {
	var lines []string
	for _, d := range defaultedQuestions {
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
			lines = append(lines, "  "+d.name+": "+label+" (change it later: "+d.later+")")
		}
	}
	return strings.Join(lines, "\n")
}
