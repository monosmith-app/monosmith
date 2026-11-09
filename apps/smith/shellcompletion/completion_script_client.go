package shellcompletion

import (
	"bytes"
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"

	"github.com/spf13/cobra"
)

// ScriptFor generates the completion script of a shell from the command tree.
// The script only calls "smith __complete", so it does not go stale when
// commands change.
//
// Parameters: root is the smith command; shell is the shell; noDescriptions
// leaves the descriptions of the commands out of the script.
// Returns: the script. The error says that cobra could not write it, or that
// the shell is not supported.
func ScriptFor(root *cobra.Command, shell shellenv.Shell, noDescriptions bool) (string, error) {
	var b bytes.Buffer
	var err error
	switch shell {
	case shellenv.Bash:
		err = root.GenBashCompletionV2(&b, !noDescriptions)
	case shellenv.Zsh:
		if noDescriptions {
			err = root.GenZshCompletionNoDesc(&b)
		} else {
			err = root.GenZshCompletion(&b)
		}
	case shellenv.Fish:
		err = root.GenFishCompletion(&b, !noDescriptions)
	case shellenv.PowerShell:
		if noDescriptions {
			err = root.GenPowerShellCompletion(&b)
		} else {
			err = root.GenPowerShellCompletionWithDesc(&b)
		}
	default:
		err = fmt.Errorf("unsupported shell %q", shell)
	}
	return b.String(), err
}
