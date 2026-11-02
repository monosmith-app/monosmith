package dependencies

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// confirmPackages shows which project manifests a named change touches and
// asks for the go-ahead: a terminal is asked, --yes passes, and without a
// terminal and without --yes it refuses and names the flag. Nothing is asked
// when no package is named (a plain smith install restores what the workspace
// already declares and edits no declaration), or when a target does not
// resolve (the command then reports that itself).
//
// Parameters: cmd is the command; s the session; verb is "add" or "remove";
// targets are the -w projects; packages the names; yes is --yes.
// Returns: whether to go on, and the refusal or ErrDeclined as the error.
func confirmPackages(cmd *cobra.Command, s *session.Session, verb string, targets, packages []string, yes bool) (bool, error) {
	if len(packages) == 0 {
		return true, nil
	}
	projects, err := resolve(s.Graph, targets)
	if err != nil || len(projects) == 0 {
		return true, nil
	}
	var lines []operationplan.Line
	for _, p := range projects {
		lines = append(lines, operationplan.Line{Action: operationplan.Change, Path: p.Root + "/", Note: verb + " " + strings.Join(packages, ", ") + " in its dependencies"})
	}
	return operationplan.Gate(cmd.OutOrStdout(), cmd.InOrStdin(), verb+" "+strings.Join(packages, ", "), "Dependencies of", lines, yes, false)
}
