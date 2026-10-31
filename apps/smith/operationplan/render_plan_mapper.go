package operationplan

import (
	"fmt"
	"io"
)

// Render writes the plan to w: the title, then each group with one line per
// path. It writes nothing for a plan without lines.
//
// Parameters: w receives the text; plan is what to show.
func Render(w io.Writer, plan Plan) {
	if plan.Empty() {
		return
	}
	fmt.Fprintf(w, "Plan: %s\n", plan.Title)
	for _, g := range plan.Groups {
		if len(g.Lines) == 0 {
			continue
		}
		fmt.Fprintf(w, "  %s\n", g.Heading)
		for _, l := range g.Lines {
			fmt.Fprintf(w, "    %-7s %s", l.Action, l.Path)
			if l.Note != "" {
				fmt.Fprintf(w, "  (%s)", l.Note)
			}
			if l.Owner != "" {
				fmt.Fprintf(w, "  [owner %s]", l.Owner)
			}
			fmt.Fprintln(w)
		}
	}
}
