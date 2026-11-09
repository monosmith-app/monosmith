package shellcompletion

import (
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
	"io"

	"github.com/monosmith-app/monosmith/libs/words/words"
)

// PrintChanges tells what install or uninstall did to each file, then what the
// person does next.
//
// Parameters: out is where the text goes; shell is the shell; changes is what
// Install or Uninstall returned.
func PrintChanges(out io.Writer, shell shellenv.Shell, changes []Change) {
	for _, c := range changes {
		fmt.Fprintln(out, words.Say("completion.change."+c.Action, "path", c.Path))
	}
	if len(changes) == 0 {
		return
	}
	for _, c := range changes {
		if c.Action == "removed" {
			return
		}
	}
	fmt.Fprintln(out, words.Say("completion.next."+string(shell), "shell", shell.Title()))
}
