package operation

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// commitAnywayFlag is the flag that commits an operation whose check failed.
const commitAnywayFlag = "commit-anyway"

// verifyChoice is what the person decides when --commit met a failed check.
type verifyChoice string

const (
	// verifyCommit commits the files of the operation although its check failed.
	verifyCommit verifyChoice = "commit"
	// verifyLeave leaves the files uncommitted for smith commit.
	verifyLeave verifyChoice = "leave"
	// verifyCancel commits nothing.
	verifyCancel verifyChoice = "cancel"
)

// askVerifyChoice asks what to do with an operation that changed files and then
// failed its check. The default leaves the files uncommitted.
//
// Parameters: in and out are the terminal; name is the operation; n the number of files.
// Returns: the choice; an empty answer leaves the files uncommitted.
func askVerifyChoice(in *bufio.Reader, out io.Writer, name string, n int) verifyChoice {
	fmt.Fprintf(out, "%s changed %d file(s), but its check failed. What about the commit you asked for?\n", name, n)
	fmt.Fprint(out, "  1) leave it uncommitted; smith commit --only-operation commits it later (default)\n"+
		"  2) commit anyway: the files exist, only a tool is missing\n"+
		"  3) cancel: commit nothing\nChoose [1]: ")
	answer, _ := in.ReadString('\n')
	switch strings.TrimSpace(answer) {
	case "", "1":
		return verifyLeave
	case "2":
		return verifyCommit
	}
	return verifyCancel
}

// settleAfterFailedCheck runs when an operation changed files and then failed
// its own check. It always writes the operation's record (so smith commit knows
// the files and the name), and it always says what happened to the commit: with
// --commit it asks in a terminal, commits with --commit-anyway, and otherwise
// refuses and names that flag. Nothing is committed unasked.
//
// Parameters: c is the command; before is the changed set when it started;
// requested is true when the run had --commit; command is the operation's name.
func settleAfterFailedCheck(c *cobra.Command, before map[string]bool, requested bool, command string) {
	root, err := workspace.FindRoot(".")
	if err != nil {
		return
	}
	files := newSince(before, changedSet(root))
	if len(files) == 0 {
		return
	}
	_ = WritePending(root, Uncommitted{Command: command, Files: files})
	out := c.OutOrStdout()
	var choice verifyChoice
	switch f := c.Flags().Lookup(commitAnywayFlag); {
	case !requested:
		fmt.Fprintln(out, notCommittedAfterFailure(false, command))
		return
	case f != nil && f.Value.String() == "true":
		choice = verifyCommit
	case offersChoice(c):
		choice = askVerifyChoice(bufio.NewReader(c.InOrStdin()), out, command, len(files))
	default:
		fmt.Fprintln(out, notCommittedAfterFailure(true, command))
		return
	}
	switch choice {
	case verifyCommit:
		pending := Always(root)
		if pending == nil {
			fmt.Fprintln(out, "Not committed: the workspace is not in a git repository.")
			return
		}
		hash, err := pending.FinishOnly(command, files)
		switch {
		case err != nil:
			fmt.Fprintf(c.ErrOrStderr(), "smith: the changes are in the work tree, not committed: %v\n", err)
		case hash != "":
			fmt.Fprintf(out, "Committed %s: chore: %s (smith undo reverts it)\n", short(hash), command)
		}
	case verifyLeave:
		fmt.Fprintf(out, "Left uncommitted. smith commit --only-operation commits %s later.\n", command)
	default:
		fmt.Fprintln(out, "Cancelled: nothing was committed.")
	}
}

// notCommittedAfterFailure is the sentence that says what happened to the
// commit of an operation whose check failed. It always names the way to commit.
//
// Parameters: requested is true when the run had --commit; command the operation.
// Returns: the sentence, starting with "Not committed."
func notCommittedAfterFailure(requested bool, command string) string {
	if requested {
		return fmt.Sprintf("Not committed: %s was written, but its check failed, and nobody could choose. "+
			"Run smith commit --only-operation to commit it as it is, or fix the cause first; "+
			"next time add --commit-anyway to commit even when the check fails.", command)
	}
	return fmt.Sprintf("Not committed. %s was written, but its check failed. "+
		"smith commit --only-operation commits it as it is, once you decide.", command)
}
