package architecture

import (
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// Command is `smith architecture`.
func Command(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "architecture [name]",
		Short: "Show the architecture of the workspace, or choose another",
		Long: `The architecture is how the workspace's code is laid out. Without a name this
lists the choices, marks the one in use, and says which languages each one
applies to and which of your projects it would check or skip; with a name it
switches to that one by editing smith.toml (the Settings tab of smith tui does
the same). "none" turns the checks off. Only the lines that turn an
architecture on are added or removed; the rest of the file stays as it is.

An architecture checks only the languages it applies to. One that applies to
none of your projects would check nothing, and smith says so before it
applies the choice.`,
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
			// The projects are read for what they say about the choices; a
			// workspace that cannot be opened still lists and switches.
			if len(args) == 1 {
				chosen, err := Find(args[0])
				if err != nil {
					return err
				}
				if chosen.Table != "" {
					members, _ := WorkspaceMembers(root, version)
					if warning := chosen.Warning(chosen.Cover(members)); warning != "" {
						fmt.Fprintln(cmd.ErrOrStderr(), "warning: "+warning+" Applying it anyway; smith architecture none turns the checks off.")
					}
				}
				if err := Set(root, chosen.Name); err != nil {
					return err
				}
				fmt.Fprintf(out, "architecture: %s\n", chosen.Name)
				return nil
			}
			current, err := Current(root)
			if err != nil {
				return err
			}
			members, _ := WorkspaceMembers(root, version)
			List(out, current, members)
			return nil
		},
	}
}

// List writes the choices, the one in use marked, each with the languages it
// applies to and, when the projects are known, what it does to them.
func List(out io.Writer, current Architecture, members []Member) {
	const indent = "                   "
	for _, a := range Catalog() {
		mark := " "
		if a.Name == current.Name {
			mark = "*"
		}
		fmt.Fprintf(out, "%s %-16s %s\n", mark, a.Name, a.Meaning)
		if applies := a.AppliesText(); applies != "" {
			fmt.Fprintf(out, "%sapplies to: %s\n", indent, applies)
		}
		if here := a.Cover(members).Here(); here != "" {
			fmt.Fprintf(out, "%shere: %s\n", indent, here)
		}
	}
}
