package commandline

import (
	"bufio"
	"fmt"
	"os"
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/completioncommand"
	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
	"github.com/monosmith-app/monosmith/apps/smith/shellcompletion"
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
	"golang.org/x/term"
)

// offersCompletionAfter is true for the commands whose run is a first look at
// smith that a person reads in a terminal: smith doctor, without a
// machine-readable format. The interface offers it too (see Execute).
func offersCompletionAfter(args []string) bool {
	return len(args) > 0 && args[0] == "doctor" && !slices.Contains(args, "--json")
}

// offerCompletion asks once whether to set up tab completion. It never asks in
// CI, when a pipe replaces the keyboard or the screen, or after the person
// said no. The answer is kept in tui.json in the smith home; see
// completioncommand.PreferencesMemory.
func offerCompletion(version string) {
	if plugintrust.InCI(os.Getenv) || !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return
	}
	root := Root(version)
	err := completioncommand.Offer(completioncommand.OfferOptions{
		Env:      shellenv.OSEnvironment{},
		Memory:   completioncommand.PreferencesMemory{Home: plugintrust.Home()},
		Terminal: completioncommand.Terminal{In: bufio.NewReader(os.Stdin), Out: os.Stdout},
		Script:   func(s shellenv.Shell) (string, error) { return shellcompletion.ScriptFor(root, s, false) },
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "smith: "+err.Error())
	}
}
