package divergences

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/derive"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// Command is `smith divergences`.
func Command(version string) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "divergences",
		Short: "List where the workspace differs from smith's defaults, each with why and how to undo it",
		Long: `List the places where this workspace made a different choice from smith's
defaults, each with the reason it is listed and how to go back:

  override  a task setting in smith.toml that replaces what a plugin infers
  ejected   a derived file you own
  drift     a derived file edited by hand, or one smith cannot edit
  differs   a file that differs from its sibling projects

A divergence is a choice, not an error: the list changes nothing and never
fails. A short list means a tidy workspace. Derived files are listed only when
[derive] enabled is true.`,
		Example: "  smith divergences\n  smith divergences --json",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr(), AllowCycles: true})
			if err != nil {
				return err
			}
			defer s.Close()
			var items []derive.Item
			var diverging []derive.Divergence
			if s.Config.Derive.Enabled {
				plan, _, err := derive.BuildPlan(s)
				if err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), "warning: derived files are not listed:", err)
				} else {
					items, diverging = plan.Items, plan.Diverging
				}
			}
			list := From(s.Config, items, diverging)
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(list)
			}
			Print(cmd.OutOrStdout(), list)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the list as JSON")
	return cmd
}

// Print writes the divergences, grouped by kind.
func Print(out io.Writer, list []Divergence) {
	if len(list) == 0 {
		fmt.Fprintln(out, "✓ No divergences: the workspace follows smith's defaults.")
		return
	}
	noun := "places where this workspace made its own choice. None of them is"
	if len(list) == 1 {
		noun = "place where this workspace made its own choice. It is not"
	}
	fmt.Fprintf(out, "%d %s an error.\n", len(list), noun)
	var kind Kind
	for _, d := range list {
		if d.Kind != kind {
			kind = d.Kind
			fmt.Fprintf(out, "\n%s\n", kind)
		}
		fmt.Fprintf(out, "  %s\n    why:  %s\n    undo: %s\n", d.Subject, d.Why, d.Undo)
	}
}
