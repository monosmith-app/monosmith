package operation

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// dirtyChoice is what the person decides when --commit met a work tree that
// already holds changes of their own.
type dirtyChoice string

const (
	// chooseOperationFiles commits only the files the operation changed.
	chooseOperationFiles dirtyChoice = "operation"
	// chooseLeave stops here and leaves everything uncommitted.
	chooseLeave dirtyChoice = "leave"
	// chooseCancel commits nothing and says so.
	chooseCancel dirtyChoice = "cancel"
)

// notCommittedText is the line printed when an operation left its changes
// uncommitted. After a run with --commit it must not suggest --commit again: the
// flag was given, and the work tree is why nothing was committed.
//
// Parameters: requested is true when the run had --commit; command is the text
// of the operation; atCommandLine is the advice for the next run without the flag.
// Returns: the sentence, starting with "Not committed."
func notCommittedText(requested bool, command, atCommandLine string) string {
	if requested {
		return fmt.Sprintf("Not committed. Your work tree had other changes. Commit or stash them, then run smith commit %s --only-operation, or review the changes with git status.", command)
	}
	return fmt.Sprintf("Not committed. Review it with git status; %s next time to commit it as one operation that smith undo can revert.", CommitNextTime(atCommandLine))
}

// askDirtyChoice asks what to do with an operation whose commit was refused
// because the work tree holds other changes. The default is to commit only the
// files the operation changed.
//
// Parameters: in and out are the terminal; files are the files the operation changed.
// Returns: the choice; an empty answer picks the operation's files.
func askDirtyChoice(in *bufio.Reader, out io.Writer, files []string) dirtyChoice {
	fmt.Fprintf(out, "Your work tree has changes of its own, so smith did not commit everything. The operation changed %d file(s).\n", len(files))
	fmt.Fprintf(out, "  1) commit only the files this operation changed (default)\n  2) stop here, leave everything uncommitted\n  3) cancel\nChoose [1]: ")
	answer, _ := in.ReadString('\n')
	switch strings.TrimSpace(answer) {
	case "", "1":
		return chooseOperationFiles
	case "2":
		return chooseLeave
	}
	return chooseCancel
}
