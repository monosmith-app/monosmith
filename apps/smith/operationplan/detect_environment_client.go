package operationplan

import (
	"os"

	"golang.org/x/term"
)

// DetectEnvironment reads the terminal state of the process and the CI
// variables (CI, GITHUB_ACTIONS, TF_BUILD).
func DetectEnvironment() Environment {
	ci := os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") == "true" || os.Getenv("TF_BUILD") != ""
	return Environment{
		Terminal: term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())),
		CI:       ci,
	}
}
