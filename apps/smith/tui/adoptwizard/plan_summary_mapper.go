package adoptwizard

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
	"github.com/monosmith-app/monosmith/apps/smith/convert"
)

// planSummaryText is the top of the review of a conversion, before the files:
// what it comes to in numbers, then the things to check (all of them), then
// what behaves differently afterwards (the first few; the rest are in the
// report that is saved once it is converted). Lines start with a bullet, never
// with + or -.
func planSummaryText(c adoptflow.Changes, r convert.Highlights) string {
	var b strings.Builder
	b.WriteString("● In short\n")
	if c.Total() > 0 {
		fmt.Fprintf(&b, "  Files: smith creates %d, edits %d and removes %d.\n", len(c.Created), len(c.Edited), len(c.Removed))
	}
	fmt.Fprintf(&b, "  Things to check: %s\n", countOrNone(len(r.Checks), "listed below"))
	fmt.Fprintf(&b, "  Things that behave differently afterwards: %s\n", countOrNone(len(r.Changed), "listed below"))
	if len(r.Checks) > 0 {
		fmt.Fprintf(&b, "\n! Check these before you go ahead (%d)\n", len(r.Checks))
		for _, x := range r.Checks {
			fmt.Fprintf(&b, "  • %s\n", x)
		}
	}
	if len(r.Changed) > 0 {
		fmt.Fprintf(&b, "\n● Behaves differently afterwards (%d)\n", len(r.Changed))
		for i, x := range r.Changed {
			if i == shownChanged {
				fmt.Fprintf(&b, "  • and %d more: the full list is saved in %s once it is converted\n", len(r.Changed)-shownChanged, convert.ReportFile)
				break
			}
			fmt.Fprintf(&b, "  • %s\n", x)
		}
	}
	return b.String()
}

func countOrNone(n int, where string) string {
	if n == 0 {
		return "none"
	}
	return fmt.Sprintf("%d (%s)", n, where)
}
