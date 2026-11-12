package operation

import (
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/commandflags"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// UndoCommand is `smith undo`.
func UndoCommand() *cobra.Command {
	var yes, dryRun, clean bool
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
smith undo refuses unless --yes is given.

git revert cannot remove files git ignores. When the undone project left generated
folders behind (node_modules, dist, a venv), smith undo lists them. In a terminal it asks
whether to delete them; --clean deletes them without asking; otherwise it only says what is left.`,
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
			var removed []string
			for _, l := range c.Changes {
				if l.Action == operationplan.Remove {
					removed = append(removed, l.Path)
				}
			}
			return settleLeftovers(cmd, root, FindLeftovers(root, removed), clean, yes)
		},
	}
	commandflags.AddYesFlag(cmd, &yes, "do not ask for confirmation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be reverted and change nothing")
	cmd.Flags().BoolVar(&clean, "clean", false, "also delete the generated folders the project left behind (node_modules, dist), which git ignores")
	return cmd
}
