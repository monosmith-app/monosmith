package commandline

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/monosmith-app/monosmith/apps/smith/updatenotice"
)

// startUpdateNotice begins the quiet update check for a command a person runs
// at a terminal, and returns what to call when the command is done: it prints
// the one line (smith was updated, or a newer smith exists) to stderr, last.
// Anywhere else (a pipe, CI, smith serve, smith mcp) it does nothing.
func startUpdateNotice(root *cobra.Command, args []string, version string) func() {
	if !canAsk(root, args) {
		return func() {}
	}
	pending := updatenotice.Start(updatenotice.ForThisProcess(version, 400*time.Millisecond))
	return func() {
		for _, line := range pending.Lines() {
			fmt.Fprintln(os.Stderr, line)
		}
	}
}
