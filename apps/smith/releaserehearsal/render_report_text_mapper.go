package releaserehearsal

import (
	"fmt"
	"github.com/monosmith-app/monosmith/libs/words/words"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/marks"
)

// RenderText prints a report in doctor's style: one line per check, failures
// and warnings with the reason and the fix under them, and a closing count.
// Passed checks show only with verbose.
func RenderText(r Report, targets int, verbose bool) string {
	var b strings.Builder
	if targets == 0 && len(r.Results) == 0 {
		return "Release rehearsal: no package is published to a registry smith can ask (npm, PyPI, NuGet, crates.io), so there is nothing to check.\n"
	}
	if targets > 0 {
		fmt.Fprintf(&b, "Release rehearsal: asking %d registr%s with read-only requests; nothing is published.\n", targets, plural(targets))
	} else {
		b.WriteString("Release rehearsal: no registry to ask; the packages attach to GitHub Releases. Read-only requests; nothing is created.\n")
	}
	for _, res := range r.Results {
		if res.Status == Pass && !verbose {
			continue
		}
		mark := map[string]string{Pass: marks.Passed, Warn: "!", Fail: marks.Failed}[res.Status]
		fmt.Fprintf(&b, "%s %s: %s on %s\n    %s\n", mark, res.Project, res.Check, res.Registry, res.Message)
		if res.Fix != "" {
			fmt.Fprintf(&b, "    → %s\n", res.Fix)
		}
	}
	fmt.Fprintf(&b, "\n%d passed, %d failed, %s\n", r.Passed, r.Failed, words.Count("doctor.summary.warnings", r.Warned))
	return b.String()
}

func plural(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}
