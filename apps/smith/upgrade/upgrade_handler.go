package upgrade

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// Command is `smith upgrade`.
func Command(version string) *cobra.Command {
	var check, asJSON bool
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Bring the workspace up to this smith: migrate smith.toml, regenerate the pipelines",
		Long: `After installing a newer smith, run smith upgrade: it migrates smith.toml to the
current schema (comments and layout kept) and regenerates the files smith owns,
deleting one the workspace no longer asks for (a dropped CI provider). Files
smith wrote once (README, .gitignore) and your own are left alone; steps in
smith:slot markers are kept. --check changes nothing and fails when something
would change, for CI.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			r, err := Upgrade(".", version, check, cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				if err := enc.Encode(r); err != nil {
					return err
				}
			} else {
				verb := "upgraded"
				if check {
					verb = "would change"
				}
				for _, m := range r.Migrations {
					fmt.Fprintf(out, "smith.toml %s\n", m)
				}
				for _, f := range r.Files {
					fmt.Fprintf(out, "%s: %s\n", verb, f)
				}
				for _, c := range r.Recipes {
					fmt.Fprintf(out, "recipe %s\n", c)
				}
				for _, l := range r.Lost {
					fmt.Fprintf(out, "warning: slot %s has no place in the new pipeline; its steps were not kept\n", l)
				}
				for _, h := range r.Hints {
					fmt.Fprintf(out, "hint: %s\n", h)
				}
				if !r.Changed() {
					fmt.Fprintln(out, "The workspace is up to date.")
				}
			}
			if check && r.Changed() {
				return errors.New("the workspace is not up to date: run smith upgrade")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "change nothing; fail when something would change")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the result as JSON")
	return cmd
}
