package fixwizard

import (
	"errors"
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
)

// failureOf tells how a fix stopped: a headline, and what the step printed.
// name is the wizard's own name ("Bring in"): "" for the plain fix wizard.
// A failure that is not a step (a file that could not be written) has no
// output.
func failureOf(err error, name string) (headline string, output []string) {
	var step *doctorfix.StepError
	if !errors.As(err, &step) {
		if name != "" {
			return name + " failed: " + err.Error(), nil
		}
		return "Could not apply the fix: " + err.Error(), nil
	}
	label := step.Step.Description
	if label == "" {
		label = strings.Join(step.Step.Argv, " ")
	}
	headline = fmt.Sprintf("Step %d of %d failed: %s (%v)", step.Number, step.Total, label, step.Err)
	if strings.TrimSpace(step.Output) != "" {
		output = strings.Split(strings.TrimRight(step.Output, "\r\n"), "\n")
	}
	return headline, output
}
