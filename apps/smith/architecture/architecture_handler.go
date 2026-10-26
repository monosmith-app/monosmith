package architecture

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// Command is `smith architecture`.
func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "architecture [name]",
		Short: "Show the architecture of the workspace, or choose another",
		Long: `The architecture is how the workspace's code is laid out. Without a name this
lists the choices and marks the one in use; with a name it switches to that
one by editing smith.toml (the Settings tab of smith tui does the same).
"none" turns the checks off. Only the lines that turn an architecture on are
added or removed; the rest of the file stays as it is.`,
		Example: `  smith architecture
  smith architecture vertical-slices
  smith architecture none`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := workspace.FindRoot(".")
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(args) == 1 {
				if err := Set(root, args[0]); err != nil {
					return err
				}
				fmt.Fprintf(out, "architecture: %s\n", args[0])
				return nil
			}
			current, err := Current(root)
			if err != nil {
				return err
			}
			for _, a := range Catalog() {
				mark := " "
				if a.Name == current.Name {
					mark = "*"
				}
				fmt.Fprintf(out, "%s %-16s %s\n", mark, a.Name, a.Meaning)
			}
			return nil
		},
	}
}
