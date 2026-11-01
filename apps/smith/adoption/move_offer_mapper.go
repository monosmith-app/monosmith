package adoption

import (
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/adoptmove"
)

// printMoveOffer tells a dry run about the project at the root: why it has to
// move, the question --init asks, and the git mv commands for the folder in the
// report (the suggested one, or the one --move named).
func printMoveOffer(w io.Writer, r Report) {
	c := *r.RootProject
	fmt.Fprintf(w, "! %s\n\n%s\n", adoptmove.Explanation(c), adoptmove.QuestionText(c))
	if r.MovePlan == nil {
		return
	}
	fmt.Fprintf(w, "Moving it into %s would run:\n", r.MovePlan.Dir)
	for _, line := range r.MovePlan.Lines() {
		fmt.Fprintf(w, "  %s\n", line)
	}
	if len(r.MovePlan.Stays) > 0 {
		fmt.Fprintf(w, "and leave at the root: %s\n", joinNames(r.MovePlan.Stays))
	}
	for _, note := range r.MovePlan.Notes {
		fmt.Fprintf(w, "  note: %s\n", note)
	}
	fmt.Fprintf(w, "\nTo move it, choose the folder yourself: %s\n", adoptmove.Command(r.MovePlan.Dir))
}

func joinNames(names []string) string {
	out := ""
	for i, n := range names {
		if i > 0 {
			out += ", "
		}
		out += n
	}
	return out
}
