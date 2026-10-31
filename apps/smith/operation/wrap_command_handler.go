package operation

import (
	"fmt"

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
	run := cmd.RunE
	cmd.RunE = func(c *cobra.Command, args []string) error {
		var pending *Pending
		if !readsOnly(c) {
			if root, err := workspace.FindRoot("."); err == nil {
				var note string
				if pending, note = Begin(root); note != "" {
					fmt.Fprintln(c.ErrOrStderr(), "smith: "+note)
				}
			}
		}
		if err := run(c, args); err != nil || pending == nil {
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

// readsOnly is true for a run that writes nothing: --dry-run or --check.
func readsOnly(c *cobra.Command) bool {
	for _, name := range []string{"dry-run", "check"} {
		if f := c.Flags().Lookup(name); f != nil && f.Value.String() == "true" {
			return true
		}
	}
	return false
}
