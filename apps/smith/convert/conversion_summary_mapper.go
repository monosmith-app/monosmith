package convert

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
)

// ReportFile is where an applied conversion saves its full report, inside the
// smith folder that .gitignore already skips.
const ReportFile = ".smith/convert-report.txt"

// keepSection lists the files convert writes that look like clutter but that
// builds or tests load: deleting one breaks every package that uses it.
func keepSection(plan *convertplan.Plan) string {
	if len(plan.Keep) == 0 {
		return ""
	}
	names := make([]string, 0, len(plan.Keep))
	for name := range plan.Keep {
		names = append(names, name)
	}
	sort.Strings(names)
	var b strings.Builder
	b.WriteString("\nFiles convert adds that the workspace needs (keep them and commit them):\n")
	for _, name := range names {
		fmt.Fprintf(&b, "  - %s: %s\n", name, plan.Keep[name])
	}
	return b.String()
}

// Summary is the last thing an applied conversion prints: what was done, the
// files to keep, and what to do next. It comes after everything else so it is
// the part still on screen.
func Summary(plan *convertplan.Plan, lostSlots []string, staged []string) string {
	var b strings.Builder
	for _, s := range lostSlots {
		fmt.Fprintf(&b, "! pipeline slot %s holds steps the new pipeline has no place for: move them into another slot\n", s)
	}
	fmt.Fprintf(&b, "\nConverted %s. The full report is in %s.\n", projectCount(plan), ReportFile)
	b.WriteString(keepSection(plan))
	if len(staged) > 0 {
		fmt.Fprintf(&b, "\nStaged with git add so they cannot be lost: %s\n", strings.Join(staged, ", "))
	}
	b.WriteString("\n" + NextSteps(plan, false))
	return b.String()
}
