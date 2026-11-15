package doctorfix

import (
	"fmt"
	"strings"
)

// Summary says, before the fixes are shown, how many there are and how each
// one is applied. Without it a `doctor --fix` run in a script or CI looks like
// plain `doctor`: the person cannot tell which fixes `--yes` applies and which
// wait for a decision.
//
// Parameters: plan is the fixes doctor offers.
// Returns: the text to print, ending in a newline; empty for an empty plan.
func Summary(plan []Planned) string {
	if len(plan) == 0 {
		return ""
	}
	var safe, asked []Planned
	for _, p := range plan {
		if p.Ask {
			asked = append(asked, p)
		} else {
			safe = append(safe, p)
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s. Nothing is changed until you apply one.\n", Headline(plan))
	if len(safe) > 0 {
		fmt.Fprintf(&b, "  --yes applies the %d safe fix(es), which only restore files smith owns: %s\n", len(safe), names(safe))
	}
	if len(asked) > 0 {
		fmt.Fprintf(&b, "  %d need(s) your decision, and --yes alone never applies them: %s\n", len(asked), names(asked))
		fmt.Fprintf(&b, "  Each is shown below with the flags that apply it: --only \"<name>\" --yes, and --choose N when it asks a question.\n")
	}
	return b.String()
}

// Reminder is the short version of Summary for the end of the output.
//
// Parameters: plan is the fixes doctor offers.
// Returns: the headline and the command that applies the safe fixes, ending in a
// newline; empty for an empty plan.
func Reminder(plan []Planned) string {
	if len(plan) == 0 {
		return ""
	}
	text := Headline(plan) + "."
	for _, p := range plan {
		if !p.Ask {
			return text + " Apply the safe ones: `smith doctor --fix --yes`.\n"
		}
	}
	return text + "\n"
}

// Headline is the count of fixes in one line: "3 fixes available: 2 applied by
// --yes, 1 needs a choice". It opens the output of `doctor --fix` without a
// terminal and closes it again, so the count is not buried under the checks.
//
// Parameters: plan is the fixes doctor offers.
// Returns: the line without a final stop; empty for an empty plan.
func Headline(plan []Planned) string {
	if len(plan) == 0 {
		return ""
	}
	asked := 0
	for _, p := range plan {
		if p.Ask {
			asked++
		}
	}
	text := fmt.Sprintf("%d %s available", len(plan), plural(len(plan), "fix", "fixes"))
	if asked == 0 {
		return text + fmt.Sprintf(": all %s applied by --yes", plural(len(plan), "is", "are"))
	}
	text += fmt.Sprintf(": %d applied by --yes", len(plan)-asked)
	if asked == 1 {
		return text + ", 1 needs a choice"
	}
	return text + fmt.Sprintf(", %d need a choice", asked)
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

func names(plan []Planned) string {
	out := make([]string, len(plan))
	for i, p := range plan {
		out[i] = fmt.Sprintf("%q", p.Description)
	}
	return strings.Join(out, ", ")
}
