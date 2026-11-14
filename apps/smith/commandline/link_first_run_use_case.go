package commandline

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"golang.org/x/term"

	"github.com/monosmith-app/monosmith/apps/smith/directlink"
)

// noLinkCommands are the commands that serve a program or a shell on the
// terminal, or answer in a line: the first run does not link in front of them
// (a line on their output would be read as an answer), and smith link and
// smith update do their own work on the entries.
var noLinkCommands = []string{"serve", "mcp", "lsp", "prompt", "link", "update", "completion", "help", "__complete", "__completeNoDesc"}

// linkFirstRun points the smith command at the program itself the first time a
// version runs behind the launcher of its package manager, then says so once,
// when the command is done. It changes nothing in CI, without a terminal, for
// the commands in noLinkCommands, when SMITH_NO_LINK=1 is set, when smith
// already starts directly and when the version was linked before.
//
// Parameters: args are the command line; version is the running smith.
// Returns: what to call when the command is done: it prints the line, if any.
func linkFirstRun(args []string, version string) func() {
	none := func() {}
	if directlink.FirstRunOff(os.Getenv) || !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stderr.Fd())) {
		return none
	}
	if len(args) > 0 && slices.Contains(noLinkCommands, args[0]) {
		return none
	}
	o := directlink.ForThisProcess(version)
	if !directlink.Needed(o) {
		return none
	}
	report, err := directlink.Link(o)
	if err != nil || len(report.Changed) == 0 {
		return none
	}
	return func() {
		names := make([]string, 0, len(report.Changed))
		for _, e := range report.Changed {
			names = append(names, e.Path)
		}
		fmt.Fprintf(os.Stderr, "smith now starts directly, with no launcher in front: about 110 ms faster and 40 MB less memory per run (%s).\n"+
			"It takes effect from the next command. smith link --status shows how smith starts, smith link --undo gives the entries back.\n", strings.Join(names, ", "))
	}
}
