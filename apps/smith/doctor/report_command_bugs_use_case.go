package doctor

import (
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// ReportCommandBugs runs the checks the plugins hold on a workspace that a
// command has just written (smith adopt, smith convert), the way smith add does
// for a new project. A check that fails because of a file is a bug in the
// command: it made the files, so it says what it got wrong and asks for a
// report. A check about this computer (IsMachineCheck) is listed apart, as
// something the user does on their machine.
//
// Parameters: s is the open session of the written workspace; command is what
// made the files ("smith convert"); out is where it reports.
// Returns: the failed checks as printed. It prints nothing and returns nil when
// every check passes.
func ReportCommandBugs(s *session.Session, command string, out io.Writer) []Finding {
	var failed, bugs, machine []Finding
	for _, f := range PluginChecks(s, nil) {
		if f.OK || f.Warning {
			continue
		}
		failed = append(failed, f)
		if IsMachineCheck(f) {
			machine = append(machine, f)
		} else {
			bugs = append(bugs, f)
		}
	}
	if len(bugs) > 0 {
		fmt.Fprintf(out, "\n%s wrote files that smith doctor rejects. That is a bug in %s: please report it with this list.\n", command, command)
		printFindings(out, bugs)
	}
	if len(machine) > 0 {
		fmt.Fprintf(out, "\nNeeds your machine (no file that %s wrote causes this):\n", command)
		printFindings(out, machine)
	}
	return failed
}

// printFindings prints each finding with its remedy on an arrow line.
func printFindings(out io.Writer, list []Finding) {
	for _, f := range list {
		fmt.Fprintf(out, "  ✖ %s\n", strings.TrimSpace(f.Check+": "+f.Detail))
		if f.Remedy != "" {
			fmt.Fprintf(out, "      → %s\n", strings.ReplaceAll(f.Remedy, "\n", "\n        "))
		}
	}
}
