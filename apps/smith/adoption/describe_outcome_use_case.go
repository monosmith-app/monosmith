package adoption

import (
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// DescribeOutcome checks the adopted workspace the way smith doctor does and
// says what is left, so "adopted" is never followed by a surprise: each
// remaining problem with its first step, then the exact next command.
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
	if len(left) == 0 {
		fmt.Fprintln(out, "✔ adopted, and smith doctor finds nothing to fix")
		fmt.Fprintln(out, "  next: smith run build")
		return nil
	}
	fmt.Fprintf(out, "✔ adopted. smith doctor still reports %d thing(s) for you:\n", len(left))
	for _, f := range left {
		fmt.Fprintf(out, "  ✖ %s\n", f.Detail)
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
