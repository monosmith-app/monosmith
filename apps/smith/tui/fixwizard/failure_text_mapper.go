package fixwizard

import (
	"errors"
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
)

// failureOf tells how a fix stopped: a headline, and what the step printed.
// A failure that is not a step (a file that could not be written) has no
// output.
func failureOf(err error) (headline string, output []string) {
	var step *doctorfix.StepError
	if !errors.As(err, &step) {
		return "Could not apply the fix: " + err.Error(), nil
	}
	name := step.Step.Description
	if name == "" {
		name = strings.Join(step.Step.Argv, " ")
	}
	headline = fmt.Sprintf("Step %d of %d failed: %s (%v)", step.Number, step.Total, name, step.Err)
	if strings.TrimSpace(step.Output) != "" {
		output = strings.Split(strings.TrimRight(step.Output, "\r\n"), "\n")
	}
	return headline, output
}
