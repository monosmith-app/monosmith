package guide

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// convertOutputsPrefix starts the answer key of a build task's outputs
// question: "outputs:<project>:<task>".
const convertOutputsPrefix = "outputs:" // words:literal

// ConvertTitle is the title of the form that converts an Nx or Turborepo
// workspace.
func ConvertTitle() string { return words.Say("guide.actions.convert_an_nx_or_turborepo_workspace") }

// ConvertOutputs is what the convert form needs to know of the plan: the tool
// converted from and the build tasks whose declared outputs differ from what
// their build writes.
type ConvertOutputs struct {
	// Source is the plan's source (nx, turborepo).
	Source string `json:"source"`
	// Choices are the plan's outputChoices.
	Choices []convertplan.OutputChoice `json:"outputChoices"`
}

// ParseConvertOutputs reads the output of `smith convert --dry-run --json`:
// the plan, after the lines that say what convert reads.
//
// Parameters: text is the captured output.
// Returns: the tool and the output choices, or an error when no plan is in
// the text.
func ParseConvertOutputs(text string) (ConvertOutputs, error) {
	at := strings.Index(text, "\n{")
	switch {
	case strings.HasPrefix(text, "{"):
		at = 0
	case at < 0:
		return ConvertOutputs{}, fmt.Errorf("%s", words.Say("guide.convert_outputs.unreadable"))
	default:
		at++
	}
	var out ConvertOutputs
	if err := json.NewDecoder(strings.NewReader(text[at:])).Decode(&out); err != nil {
		return ConvertOutputs{}, fmt.Errorf("%s: %w", words.Say("guide.convert_outputs.unreadable"), err) // words:literal
	}
	return out, nil
}

// ConvertOutputsAction adds to the convert form one question for each build
// task whose declared outputs differ from what its build writes, before the
// commit question: what the build writes (the suggestion, first) or what the
// old tool declared. The answers reach the command as
// --outputs <project>:<task>=detected|declared.
//
// Parameters: base is the plain convert form; o is what the plan holds.
// Returns: the form with the questions; base itself when there are none.
func ConvertOutputsAction(base Action, o ConvertOutputs) Action {
	if len(o.Choices) == 0 {
		return base
	}
	tool := "Nx" // words:literal
	if o.Source == convertplan.SourceTurborepo {
		tool = "Turborepo" // words:literal
	}
	var steps []Step
	for _, c := range o.Choices {
		task, detected, declared := c.Project+":"+c.Task, strings.Join(c.Detected, ", "), strings.Join(c.Declared, ", ")
		steps = append(steps, Step{
			Key:    convertOutputsPrefix + task,
			Prompt: words.Say("guide.convert_outputs.prompt", "task", task),
			Note:   words.Say("guide.convert_outputs.note", "detected", detected, "tool", tool, "declared", declared),
			Options: []Option{
				{Label: words.Say("guide.convert_outputs.detected", "detected", detected), Value: "detected", Hint: words.Say("guide.convert_outputs.detected_hint")},
				{Label: words.Say("guide.convert_outputs.declared", "declared", declared, "tool", tool), Value: "declared", Hint: words.Say("guide.convert_outputs.declared_hint")},
			},
		})
	}
	a := base
	at := slices.IndexFunc(base.Steps, func(s Step) bool { return s.Key == "commit" })
	if at < 0 {
		at = len(base.Steps)
	}
	a.Steps = slices.Concat(base.Steps[:at], steps, base.Steps[at:])
	a.Argv = func(answers map[string][]string) []string {
		argv := base.Argv(answers)
		for _, c := range o.Choices {
			task := c.Project + ":" + c.Task
			if v := first(answers, convertOutputsPrefix+task); v != "" {
				argv = append(argv, "--outputs", task+"="+v)
			}
		}
		return argv
	}
	return a
}
