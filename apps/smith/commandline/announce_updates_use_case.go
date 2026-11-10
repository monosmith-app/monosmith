package commandline

import (
	"fmt"
	"os"
	"strings"
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
	o := updatenotice.ForThisProcess(version, 400*time.Millisecond)
	o.Always = asksAboutVersion(args)
	pending := updatenotice.Start(o)
	return func() {
		for _, line := range pending.Lines() {
			fmt.Fprintln(os.Stderr, line)
		}
	}
}

// asksAboutVersion is true for `smith --version`, `smith version` and
// `smith doctor`: the person looks at the version or the health of the setup,
// so a newer smith is worth saying every time, not once a day.
func asksAboutVersion(args []string) bool {
	for _, a := range args {
		switch a {
		case "--version", "version", "doctor":
			return true
		}
		if !strings.HasPrefix(a, "-") {
			return false
		}
	}
	return false
}
