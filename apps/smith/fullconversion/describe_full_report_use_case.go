package fullconversion

import (
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
)

// maxKeptShown is how many kept files and packages are named before "and N
// more".
const maxKeptShown = 12

// Describe writes the plan for people: the choices, each group of removals
// (as diffs and the packages when detail is set, else as one line), what
// stays and why, and the checks that follow. root is where the diffs read the
// files from.
func Describe(w io.Writer, root string, r Report, detail bool) error {
	fmt.Fprintf(w, "\nFull conversion: smith becomes the only tool\n")
	fmt.Fprintf(w, "  lint: %s   tests: %s   build: %s   CI: %s\n", r.Choices.Lint, r.Choices.Test, r.Choices.Build, r.Choices.CI)
	if len(r.Groups) == 0 {
		fmt.Fprintln(w, "\nNothing of the old tooling is left to remove.")
	}
	for _, g := range r.Groups {
		fmt.Fprintln(w)
		if detail {
			if err := doctorfix.ShowFix(w, root, g.Fix); err != nil {
				return err
			}
			if len(g.Removes) > 0 {
				fmt.Fprintf(w, "  packages removed: %s\n", strings.Join(g.Removes, ", "))
			}
			continue
		}
		fmt.Fprintf(w, "● %s\n", g.Reason)
		if len(g.Removes) > 0 {
			fmt.Fprintf(w, "  packages: %s\n", strings.Join(g.Removes, ", "))
		}
	}
	if len(r.Kept) > 0 {
		fmt.Fprintf(w, "\nLooks like the old tooling but stays (something still uses it):\n")
		for i, k := range r.Kept {
			if i == maxKeptShown {
				fmt.Fprintf(w, "  and %d more (smith convert --level full --json lists them)\n", len(r.Kept)-maxKeptShown)
				break
			}
			fmt.Fprintf(w, "  - %s: %s\n", k.What, k.Why)
		}
	}
	for _, n := range r.Notes {
		fmt.Fprintf(w, "\n! %s\n", n)
	}
	fmt.Fprintln(w)
	return doctorfix.ShowFix(w, root, r.Finish)
}
