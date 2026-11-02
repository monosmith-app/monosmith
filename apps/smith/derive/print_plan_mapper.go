package derive

import (
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/textdiff"
)

// PrintPlan writes one line per derived file that needs action, then a count.
// Files that are current are only counted.
func PrintPlan(w io.Writer, plan Plan) {
	current := 0
	for _, i := range plan.Items {
		if !i.NeedsAction() {
			current++
			continue
		}
		fmt.Fprintf(w, "  %-10s %s  (%s)  %s\n", strings.ToUpper(string(i.Status[:1]))+string(i.Status[1:]), i.File.Path, i.File.Renderer, i.Detail)
	}
	fmt.Fprintf(w, "%d derived files, %d need action, %d up to date or owned by you.\n", len(plan.Items), len(plan.Pending()), current)
}

// PrintDiffs shows, for each item that holds a decision, what smith would
// change in the content it owns.
func PrintDiffs(w io.Writer, items []Item) {
	for _, i := range items {
		if i.Present == nil || i.Wanted == nil {
			continue
		}
		fmt.Fprintf(w, "\n%s (owner %s, %s):\n%s", i.File.Path, i.Owner, i.File.Renderer, textdiff.UnifiedDiff(i.File.Path, string(i.Present), string(i.Wanted)))
	}
}

// PrintExplanation says everything smith knows about one derived file.
func PrintExplanation(w io.Writer, items []Item) {
	for _, i := range items {
		fmt.Fprintf(w, "%s\n  owner     %s\n  renderer  %s v%d (plugin %s)\n  mode      %s\n  project   %s\n  state     %s: %s\n", i.File.Path, i.Owner, i.File.Renderer, i.File.Version, i.Plugin, i.File.Mode, i.File.Project, i.Status, i.Detail)
		if i.Entry != nil {
			fmt.Fprintf(w, "  recorded  %s\n", i.Entry.Hash)
		}
		PrintDiffs(w, filterDiffable(i))
	}
}

func filterDiffable(i Item) []Item {
	if i.Status == StatusDrift || i.Status == StatusDiverged || i.Status == StatusStale {
		return []Item{i}
	}
	return nil
}
