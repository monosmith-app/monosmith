package adoption

import (
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/features/agentfiles"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/features/session"
	"github.com/monosmith-app/monosmith/apps/smith/shared/marks"
)

// DescribeOutcome checks the adopted workspace the way smith doctor does and
// says what is left, so "adopted" is never followed by a surprise: each
// remaining problem with its first step, then the exact next command. It is
// the one "left to do" list adopt prints: the steps there are expected after
// a conservative adoption (install, a lint config), not faults of adopt.
func DescribeOutcome(root, version string, out io.Writer) error {
	s, err := session.Open(session.Options{Dir: root, Version: version, Log: io.Discard})
	if err != nil {
		return err
	}
	defer s.Close()
	report := doctor.Run(s)
	var left []doctor.Finding
	for _, f := range report.Findings {
		if !f.OK && !f.Warning {
			left = append(left, f)
		}
	}
	if len(s.Graph.Projects) == 0 {
		fmt.Fprintln(out, "! workspace created, but it has no projects: nothing in this repository was adopted as one")
		fmt.Fprintln(out, "  next: move your code into packages/<name> (smith finds it there), or run smith add app <name> --lang <language>")
		return nil
	}
	if len(left) == 0 {
		fmt.Fprintln(out, marks.Passed+" adopted, and smith doctor finds nothing to fix")
		fmt.Fprintln(out, "  next: smith run build")
		offerAgent(s.Root, s.Config.AgentEnabled(), out)
		return nil
	}
	fmt.Fprintf(out, marks.Passed+" adopted. Left to do (%d):\n", len(left))
	for _, f := range left {
		fmt.Fprintf(out, "  × %s\n", f.Detail)
		for i, line := range strings.Split(f.Remedy, "\n") {
			if i == 0 {
				line = "→ " + line
			} else {
				line = "  " + line
			}
			fmt.Fprintf(out, "    %s\n", line)
		}
	}
	fmt.Fprintln(out, "  next: follow the steps above, then run smith doctor")
	return nil
}

// offerAgent tells a workspace without the MonoSmith agent files how to add
// them. It never writes: smith agent install shows the files and asks.
func offerAgent(root string, enabled bool, out io.Writer) {
	if enabled && !agentfiles.Finding(root).OK {
		fmt.Fprintln(out, "  agent: smith agent install adds the MonoSmith agent, so Claude Code and VS Code work through smith (optional)")
	}
}
