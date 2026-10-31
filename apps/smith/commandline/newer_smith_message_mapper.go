package commandline

import (
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// newerSmithMessage is what smith prints when smith.toml needs a newer smith:
// one sentence and the one command that fixes it. With verbose, the parse
// errors follow. It returns false for any other error.
func newerSmithMessage(err error, version string, verbose bool) (string, bool) {
	newer, ok := workspace.AsNewerSmith(err)
	if !ok {
		return "", false
	}
	message := newer.Summary(version)
	if verbose {
		message += "\n\n" + newer.Detail
	}
	return message, true
}

// wantsVerbose says whether the command line asks for --verbose.
func wantsVerbose(args []string) bool {
	return slices.Contains(args, "--verbose") || slices.Contains(args, "-v")
}
