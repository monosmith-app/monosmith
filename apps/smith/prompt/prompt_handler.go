package prompt

import (
	"encoding/json"
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// Command is `smith prompt`. It never opens a session: it finds the workspace
// root, reads the saved summary and asks git for the changed files.
func Command(version string) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "prompt",
		Short: "Print a one-line status for a shell prompt, from the last doctor run",
		Long: `Print a one-line status for a shell prompt (Oh My Posh, Starship, powerlevel10k):
the smith version, the number of projects, the failed checks and warnings of the
last smith doctor run, and the projects the working tree changes touch.

It answers in about 50 ms because it only reads what a normal smith doctor run
saved in .smith/prompt.json and asks git which files changed. It never runs doctor
and never starts a plugin, so the numbers are as old as the last doctor run. With
no saved run it prints the version alone and says nothing is wrong. Outside a
smith workspace it prints nothing and exits 0. --json prints the same facts for
a custom segment.`,
		Example: "  smith prompt\n  smith prompt --json",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := workspace.FindRoot(".")
			if err != nil {
				return nil
			}
			summary, ok := Load(root)
			var changed []string
			if ok {
				changed = ChangedFiles(root)
			}
			view := NewView(version, summary, ok, changed)
			out := cmd.OutOrStdout()
			if asJSON {
				return json.NewEncoder(out).Encode(view)
			}
			_, err = fmt.Fprintln(out, view.Line())
			return err
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the facts as JSON for a custom prompt segment")
	return cmd
}
