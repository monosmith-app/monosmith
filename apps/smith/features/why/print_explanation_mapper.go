package why

import (
	"fmt"
	"io"
	"strings"
)

// Print writes the explanation as labelled lines.
func Print(w io.Writer, e Explanation) {
	fmt.Fprintln(w, e.Path)
	fmt.Fprintf(w, "  %-8s %s\n", "Summary", e.Summary)
	row := func(label, value string) {
		if value != "" {
			fmt.Fprintf(w, "  %-8s %s\n", label, value)
		}
	}
	if e.Owner != "" {
		row("Owner", e.Owner)
	}
	row("Project", e.Project)
	row("Because", e.Source)
	row("Rule", e.Rule)
	for _, p := range e.Parts {
		state := ""
		if p.State != "" {
			state = ", " + p.State
			if p.Detail != "" {
				state += ": " + p.Detail
			}
		}
		row("Part", fmt.Sprintf("%s v%d (%s, owner %s%s)", p.Renderer, p.Version, p.Mode, p.Owner, state))
	}
	if e.Stack != "" {
		stack := e.Stack
		if len(e.Choices) > 0 {
			stack += " (choices: " + strings.Join(e.Choices, ", ") + ")"
		}
		row("Stack", stack)
	}
	for i, c := range e.Change {
		label := ""
		if i == 0 {
			label = "Change"
		}
		fmt.Fprintf(w, "  %-8s %s\n", label, c)
	}
}
