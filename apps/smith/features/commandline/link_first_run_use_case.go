package commandline

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"golang.org/x/term"

	"github.com/monosmith-app/monosmith/apps/smith/features/directlink"
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
// already starts directly (from the link, or handed off by it) and when the
// version was linked before.
//
// Parameters: args are the command line; version is the running smith.
// Returns: what to call when the command is done (it prints the line, if any), and
// whether smith was linked by this run: the interface then offers to restart.
func linkFirstRun(args []string, version string) (done func(), linked bool) {
	none := func() {}
	// A run the link handed to another version is linked already.
	if strings.HasPrefix(os.Getenv("SMITH_LAUNCHER"), directlink.LauncherPrefix) {
		return none, false
	}
	if directlink.FirstRunOff(os.Getenv) || !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stderr.Fd())) {
		return none, false
	}
	if len(args) > 0 && slices.Contains(noLinkCommands, args[0]) {
		return none, false
	}
	o := directlink.ForThisProcess(version)
	if !directlink.Needed(o) {
		return none, false
	}
	report, err := directlink.Link(o)
	if err != nil || len(report.Changed) == 0 {
		return none, false
	}
	return func() {
		names := make([]string, 0, len(report.Changed))
		for _, e := range report.Changed {
			names = append(names, e.Path)
		}
		fmt.Fprintf(os.Stderr, "smith now starts directly, with no launcher in front: about 110 ms faster and 40 MB less memory per run (%s).\n"+
			"It takes effect from the next command. smith link --status shows how smith starts, smith link --undo gives the entries back.\n", strings.Join(names, ", "))
	}, true
}

// restartInterface opens the interface again in a new process: the first run made
// the smith command start the program directly, and only a new start uses it. It
// returns the exit code of the new interface.
func restartInterface() int {
	program, err := exec.LookPath("smith")
	if err != nil {
		program = os.Args[0]
	}
	cmd := exec.Command(program)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return exit.ExitCode()
		}
		fmt.Fprintln(os.Stderr, "smith: could not open the interface again: "+err.Error())
		return 1
	}
	return 0
}
