package operation

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// Spec says how a mutating command joins the operation log.
type Spec struct {
	// Describe is the text of the commit: "add package billing". It receives
	// the command and its arguments.
	Describe func(cmd *cobra.Command, args []string) string
	// AddYes adds --yes (-y) when the command has no such flag.
	AddYes bool
	// AddDryRun adds --dry-run when the command has no such flag.
	AddDryRun bool
	// ChecksResult says the command checks its own result after it wrote files
	// and returns AfterWork(err) when that check fails. It adds --commit-anyway.
	ChecksResult bool
}

// Wrap makes cmd commit its work as one operation when [workspace]
// commit_operations is true and the work tree was clean (ADR 0047). A run with
// --dry-run or --check, a failed run, and a run that changed no file make no
// commit. A commit that git refuses does not fail the command: the changes stay
// in the work tree and smith says why.
//
// Parameters: cmd is the command to wrap; spec describes the operation.
// Returns: cmd, for chaining.
func Wrap(cmd *cobra.Command, spec Spec) *cobra.Command {
	if spec.AddYes && cmd.Flags().Lookup("yes") == nil {
		cmd.Flags().BoolP("yes", "y", false, "do not ask for confirmation")
	}
	if spec.AddDryRun && cmd.Flags().Lookup("dry-run") == nil {
		cmd.Flags().Bool("dry-run", false, "show the plan and change nothing")
	}
	if cmd.Flags().Lookup("commit") == nil {
		cmd.Flags().Bool("commit", false, "commit the change as one operation so smith undo can revert it (nothing is committed without it)")
	}
	if spec.ChecksResult && cmd.Flags().Lookup(commitAnywayFlag) == nil {
		cmd.Flags().Bool(commitAnywayFlag, false, "with --commit, commit even when the check of the new project fails (the files exist, a tool is missing)")
	}
	run := cmd.RunE
	cmd.RunE = func(c *cobra.Command, args []string) error {
		var pending *Pending
		var before map[string]bool
		requested, dirty := false, false
		if !readsOnly(c) {
			if root, err := workspace.FindRoot("."); err == nil {
				before = changedSet(root)
				var note string
				begin := Begin
				if f := c.Flags().Lookup("commit"); f != nil && f.Value.String() == "true" {
					begin, requested = BeginRequested, true
				}
				if pending, note = begin(root); note != "" {
					fmt.Fprintln(c.ErrOrStderr(), "smith: "+note)
					dirty = strings.HasSuffix(note, "uncommitted changes")
				}
			}
		}
		err := run(c, args)
		if err != nil && isVerifyFailure(err) && !readsOnly(c) {
			settleAfterFailedCheck(c, before, requested, spec.Describe(c, args))
			return err
		}
		if err == nil && pending == nil && !readsOnly(c) && !dryRunning(c) {
			if root, rootErr := workspace.FindRoot("."); rootErr == nil && hasChanges(root) {
				command := spec.Describe(c, args)
				files := newSince(before, changedSet(root))
				fmt.Fprintln(c.OutOrStdout(), notCommittedText(requested, command, "add --commit"))
				// Remember which files this operation changed, so smith commit can leave the others out.
				_ = WritePending(root, Uncommitted{Command: command, Files: files})
				if requested && dirty && len(files) > 0 && offersChoice(c) {
					offerOperationCommit(c, root, command, files)
				}
			}
		}
		if err != nil || pending == nil {
			return err
		}
		command := spec.Describe(c, args)
		hash, err := pending.Finish(command)
		switch {
		case err != nil:
			fmt.Fprintf(c.ErrOrStderr(), "smith: the changes are in the work tree, not committed: %v\n", err)
		case hash != "":
			fmt.Fprintf(c.OutOrStdout(), "Committed %s: chore: %s (smith undo reverts it)\n", short(hash), command)
		}
		return nil
	}
	return cmd
}

// offersChoice is true when a person can answer: a terminal, no CI, and no --yes.
func offersChoice(c *cobra.Command) bool {
	if f := c.Flags().Lookup("yes"); f != nil && f.Value.String() == "true" {
		return false
	}
	env := operationplan.DetectEnvironment()
	return env.Terminal && !env.CI
}

// offerOperationCommit lets the person commit only the files the operation
// changed after --commit met a work tree with changes of its own.
func offerOperationCommit(c *cobra.Command, root, command string, files []string) {
	out := c.OutOrStdout()
	switch askDirtyChoice(bufio.NewReader(c.InOrStdin()), out, files) {
	case chooseOperationFiles:
		hash, err := Always(root).FinishOnly(command, files)
		switch {
		case err != nil:
			fmt.Fprintf(c.ErrOrStderr(), "smith: the changes are in the work tree, not committed: %v\n", err)
		case hash != "":
			fmt.Fprintf(out, "Committed %s: chore: %s (smith undo reverts it)\n", short(hash), command)
		}
	case chooseLeave:
		fmt.Fprintln(out, "Left uncommitted. smith commit --only-operation commits the files of this operation later.")
	default:
		fmt.Fprintln(out, "Cancelled: nothing was committed.")
	}
}

// readsOnly is true for a run that writes nothing: --dry-run or --check.
func readsOnly(c *cobra.Command) bool {
	for _, name := range []string{"dry-run", "check"} {
		if f := c.Flags().Lookup(name); f != nil && f.Value.String() == "true" {
			return true
		}
	}
	return false
}
