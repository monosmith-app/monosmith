package doctor

import (
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// ReportCommandBugs runs the checks the plugins hold on a workspace that a
// command has just written (smith adopt, smith convert), the way smith add does
// for a new project. A check that fails there is a bug in the command, not in
// the repository: the command made the files, so it says what it got wrong and
// asks for a report.
//
// Parameters: s is the open session of the written workspace; command is what
// made the files ("smith convert"); out is where it reports.
// Returns: the failed checks as printed. It prints nothing and returns nil when
// every check passes.
func ReportCommandBugs(s *session.Session, command string, out io.Writer) []Finding {
	var failed []Finding
	for _, f := range PluginChecks(s, nil) {
		if !f.OK && !f.Warning {
			failed = append(failed, f)
		}
	}
	if len(failed) == 0 {
		return nil
	}
	fmt.Fprintf(out, "\n%s wrote files that smith doctor rejects. That is a bug in %s: please report it with this list.\n", command, command)
	for _, f := range failed {
		fmt.Fprintf(out, "  ✖ %s\n", strings.TrimSpace(f.Check+": "+f.Detail))
		if f.Remedy != "" {
			fmt.Fprintf(out, "      → %s\n", strings.ReplaceAll(f.Remedy, "\n", "\n        "))
		}
	}
	return failed
}
