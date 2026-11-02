package operationplan

import (
	"os"

	"golang.org/x/term"
)

// DetectEnvironment reads the terminal state of the process and the CI
// variables (CI, GITHUB_ACTIONS, TF_BUILD).
//
// SMITH_FORCE_INTERACTIVE=1 reports a terminal and no CI, so a script can feed
// the answers on standard input.
func DetectEnvironment() Environment {
	if os.Getenv("SMITH_FORCE_INTERACTIVE") == "1" {
		return Environment{Terminal: true}
	}
	ci := os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") == "true" || os.Getenv("TF_BUILD") != ""
	return Environment{
		Terminal: term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())),
		CI:       ci,
	}
}
