package lockfilesync

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// Command is `smith npm-lockfile`: install the packages and leave
// package-lock.json in step with package.json. Hidden: the closing step of a
// full conversion runs it, people run npm install.
func Command() *cobra.Command {
	return &cobra.Command{
		Use:    "npm-lockfile",
		Short:  "Install the packages and settle package-lock.json, rebuilding it when git can take that back",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := workspace.FindRoot(".")
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			res, err := Sync(Options{Root: root, Out: out, Create: true})
			if err != nil {
				return err
			}
			if res.Skipped != "" {
				_, err = fmt.Fprintf(out, "package-lock.json left alone: %s\n", res.Skipped)
				return err
			}
			for _, n := range res.Notes {
				fmt.Fprintln(out, "  "+n)
			}
			_, err = fmt.Fprintln(out, "package-lock.json is in step with package.json")
			return err
		},
	}
}
