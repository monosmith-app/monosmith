package doctorfix

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/marks"
)

// FindingText is what a closing report says of an advisory step that failed:
// the step, each failure it printed with its remedies, and the command that
// checks again.
//
// Parameters: f is the finding. Returns the lines, each ending in a newline.
func FindingText(f Finding) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s\n", marks.Warning, stepName(f.Step))
	for _, line := range failureLines(f.Output) {
		b.WriteString("  " + line + "\n")
	}
	fmt.Fprintf(&b, "  → check again: %s\n", displayCommand(f.Step.Argv))
	return b.String()
}

// failureLines picks from what a smith check printed the result lines that
// failed or warn (× or ▲ at the start of a line) and the → remedies under
// each. Passed checks, "!" notes, tables and detail lines are left out.
func failureLines(output string) []string {
	var out []string
	keep := false
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimRight(raw, "\r")
		text := strings.TrimSpace(line)
		if text == "" {
			continue
		}
		remedy := strings.HasPrefix(text, "→ ")
		if !remedy && line[0] != ' ' && line[0] != '\t' {
			keep = strings.HasPrefix(text, marks.Failed+" ") || strings.HasPrefix(text, marks.Warning+" ")
			if keep {
				out = append(out, text)
			}
			continue
		}
		if keep && remedy {
			out = append(out, "  "+text)
		}
	}
	return out
}
