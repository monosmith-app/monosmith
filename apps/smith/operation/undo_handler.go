package operation

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// UndoCommand is `smith undo`.
func UndoCommand() *cobra.Command {
	var yes, dryRun bool
	cmd := &cobra.Command{
		Use:   "undo",
		Short: "Revert the last operation smith committed (add, remove, upgrade, derive, sync)",
		Long: `Revert the last operation smith made. With commit_operations = true under
[workspace] in smith.toml, each mutating command commits its work as one
commit; smith undo shows what that commit changed, asks (in a terminal), and
reverts it with git revert: a new commit, history untouched.

It refuses, and says why, when a later commit or an uncommitted change touches
the same files, or when the commit is no longer in the history. --dry-run shows
what it would revert. --yes skips the question; without a terminal, or in CI,
smith undo refuses unless --yes is given.`,
		Example: "  smith undo --dry-run\n  smith undo",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := workspace.FindRoot(".")
			if err != nil {
				return err
			}
			c, err := Prepare(root)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			ok, err := operationplan.Present(out, cmd.InOrStdin(), c.Plan(), operationplan.Options{
				Yes: yes, DryRun: dryRun, Ask: true, Env: operationplan.DetectEnvironment(),
			})
			if err != nil || !ok {
				return err
			}
			hash, err := Undo(root, c)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "Reverted %s in commit %s.\n", c.Record.Command, short(hash))
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "do not ask for confirmation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be reverted and change nothing")
	return cmd
}
