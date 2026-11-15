package release

import (
	"fmt"
	"strings"
)

// RenderJournalStatus is what `smith release --status` prints: for each
// project, the steps its releases got through, oldest first.
func RenderJournalStatus(journals []Journal) string {
	if len(journals) == 0 {
		return "No release journal yet: nothing has been published from here.\n"
	}
	var b strings.Builder
	for _, j := range journals {
		fmt.Fprintf(&b, "%s\n", j.Project)
		for _, e := range j.Entries {
			fmt.Fprintf(&b, "  %s  %-8s %-14s %-9s %s", e.Time.Format("2006-01-02 15:04"), e.Version, e.Step, e.Outcome, e.Target)
			if e.Error != "" {
				fmt.Fprintf(&b, " (%s)", e.Error)
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}
