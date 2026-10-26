package convert

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
)

// Highlights are what a person reads first after a conversion: what behaves
// differently, and what needs them. Each line is one message; a message that
// applies to several projects is one line with the count.
type Highlights struct {
	// Changed are the messages of "behaves differently".
	Changed []string
	// Checks are the things to look at, decisions first.
	Checks []string
	// Projects reads "23 projects (22 packages, 1 app)".
	Projects string
}

// HighlightsOf reads the plan's findings into the lines of the result screen.
func HighlightsOf(plan *convertplan.Plan) Highlights {
	line := func(g group) string {
		l := g.line()
		if n := len(g.projects); n > 1 {
			l += fmt.Sprintf(" (%d projects)", n)
		}
		return l
	}
	h := Highlights{Projects: projectCount(plan)}
	var changed []convertplan.Finding
	for _, f := range withoutShadowed(plan.Findings) {
		if f.Severity == convertplan.Changed && !f.Decide {
			changed = append(changed, f)
		}
	}
	for _, g := range groupFindings(changed, true) {
		h.Changed = append(h.Changed, line(g))
	}
	_, checks := attention(plan)
	for _, g := range groupFindings(checks, false) {
		h.Checks = append(h.Checks, strings.TrimSpace(line(g)))
	}
	return h
}
