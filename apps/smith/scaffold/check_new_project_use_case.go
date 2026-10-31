package scaffold

import (
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// CheckNewProject runs the doctor checks the plugins hold for the new project
// (npm workspaces, scripts, entry files) so smith add never says "Done" for a
// project that smith doctor then rejects. It prints each failed check with its
// remedy and returns an error that says the project fails smith doctor; it
// returns nil when every check passes.
func CheckNewProject(s *session.Session, p *projectgraph.Project, out io.Writer) error {
	var failed []doctor.Finding
	for _, f := range doctor.PluginChecks(s, map[string]bool{p.Name: true}) {
		about := f.Check + " " + f.Detail
		if !f.OK && !f.Warning && (strings.Contains(about, p.Root) || strings.Contains(about, p.Name)) {
			failed = append(failed, f)
		}
	}
	if len(failed) == 0 {
		return nil
	}
	fmt.Fprintf(out, "The new project %s fails smith doctor:\n", p.Name)
	for _, f := range failed {
		fmt.Fprintf(out, "  ✖ %s\n", f.Detail)
		for i, line := range strings.Split(f.Remedy, "\n") {
			if i == 0 {
				line = "→ " + line
			} else {
				line = "  " + line
			}
			fmt.Fprintf(out, "      %s\n", line)
		}
	}
	return fmt.Errorf("the new project %s was written but fails smith doctor; follow the steps above, then run smith doctor", p.Name)
}
