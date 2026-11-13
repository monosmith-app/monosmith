package convert

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
)

// The answers --outputs takes for a build task's outputs.
const (
	// OutputsDetected caches what the build writes (the suggestion).
	OutputsDetected = "detected"
	// OutputsDeclared caches the outputs the old tool declared.
	OutputsDeclared = "declared"
)

// ApplyOutputAnswers applies the answers of --outputs to the plan, so a run
// with no terminal (the TUI, CI) can choose what a terminal asks.
//
// Parameters: plan is the conversion plan, changed in place when an answer
// keeps the declared outputs; answers are "<project>:<task>=declared" or
// "<project>:<task>=detected".
// Returns: the "<project>:<task>" keys answered, which AskOutputChoices does
// not ask again, and an error when an answer is malformed, names a task
// with no such choice, or is neither declared nor detected.
func ApplyOutputAnswers(plan *convertplan.Plan, answers []string) (map[string]bool, error) {
	given := map[string]bool{}
	for _, a := range answers {
		key, value, ok := strings.Cut(a, "=")
		if !ok || !strings.Contains(key, ":") {
			return nil, fmt.Errorf("--outputs %q is not <project>:<task>=declared or <project>:<task>=detected", a)
		}
		i := outputChoiceIndex(plan, key)
		if i < 0 {
			return nil, fmt.Errorf("--outputs %q: %s has no outputs to choose (smith convert --dry-run --json lists them under outputChoices)", a, key)
		}
		switch value {
		case OutputsDetected:
		case OutputsDeclared:
			plan.KeepDeclaredOutputs(i)
		default:
			return nil, fmt.Errorf("--outputs %q: the answer is declared or detected, not %q", a, value)
		}
		given[key] = true
	}
	return given, nil
}

// outputChoiceIndex is the index of the choice of "<project>:<task>", or -1.
func outputChoiceIndex(plan *convertplan.Plan, key string) int {
	for i, c := range plan.OutputChoices {
		if c.Project+":"+c.Task == key {
			return i
		}
	}
	return -1
}
