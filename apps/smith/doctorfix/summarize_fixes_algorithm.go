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
	fmt.Fprintf(&b, "%d fix(es) available. Nothing is changed until you apply one.\n", len(plan))
	if len(safe) > 0 {
		fmt.Fprintf(&b, "  --yes applies the %d safe fix(es), which only restore files smith owns: %s\n", len(safe), names(safe))
	}
	if len(asked) > 0 {
		fmt.Fprintf(&b, "  %d need(s) your decision, and --yes alone never applies them: %s\n", len(asked), names(asked))
		fmt.Fprintf(&b, "  Each is shown below with the flags that apply it: --only \"<name>\" --yes, and --choose N when it asks a question.\n")
	}
	return b.String()
}

func names(plan []Planned) string {
	out := make([]string, len(plan))
	for i, p := range plan {
		out[i] = fmt.Sprintf("%q", p.Description)
	}
	return strings.Join(out, ", ")
}
