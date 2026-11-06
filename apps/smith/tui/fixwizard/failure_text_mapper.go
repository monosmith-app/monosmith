package fixwizard

import (
	"errors"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// failureOf tells how a fix stopped: a headline, and what the step printed.
// name is the wizard's own name ("Bring in"): "" for the plain fix wizard.
// A failure that is not a step (a file that could not be written) has no
// output.
func failureOf(err error, name string) (headline string, output []string) {
	var step *doctorfix.StepError
	if !errors.As(err, &step) {
		if name != "" {
			return words.Say("fixwizard.failed.named", "name", name, "error", err.Error()), nil
		}
		return words.Say("fixwizard.failed.fix", "error", err.Error()), nil
	}
	label := step.Step.Description
	if label == "" {
		label = strings.Join(step.Step.Argv, " ")
	}
	headline = words.Say("fixwizard.failed.step", "number", step.Number, "total", step.Total, "label", label, "error", step.Err)
	if strings.TrimSpace(step.Output) != "" {
		output = strings.Split(strings.TrimRight(step.Output, "\r\n"), "\n")
	}
	return headline, output
}
