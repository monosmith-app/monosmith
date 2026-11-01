package swapchoice

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/nearmatch"
	"github.com/monosmith-app/monosmith/apps/smith/operation"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// Command is `smith swap`.
func Command(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap <project> <axis> <choice>",
		Short: "Change one decision of a project, such as its test runner or linter",
		Long: `Change one decision of an existing project. The axes are the questions that
smith choices lists (test, lint, style, framework); the choice is one of the
answers on that axis.

smith checks the swap with the choice solver, the same one smith add --with
uses: the project keeps one choice per axis, and what the new choice requires or
conflicts with is checked. It then shows the plan: the choice that leaves, the
choice that joins, and the choices the new one pulls in.

smith moves a project's dependencies, config and imports only where the plugin
that owns the choice can do it without losing your edits. No plugin does it yet,
so smith swap stops after the plan and says what to change by hand. Run it with
--dry-run to see the plan and nothing else.`,
		Example: "  smith swap web test vitest --dry-run\n  smith swap api lint oxlint",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			p := s.Graph.Get(args[0])
			if p == nil {
				return fmt.Errorf("no project named %q (projects: %s).%s", args[0], strings.Join(s.Graph.Names(), ", "), nearmatch.Suggestion(args[0], s.Graph.Names()))
			}
			if p.Stack == "" {
				return fmt.Errorf("%s records no stack, so smith does not know its choices. Name them with smith add --with in a new project, or set the stack in smith.toml", p.Name)
			}
			st, err := s.Stacks.Get(p.Stack)
			if err != nil {
				return err
			}
			swap, err := PlanSwap(s.Plugins.Offers(), st.Choices, args[1], args[2])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Swap %s of %s: %s -> %s\n", swap.Axis, p.Name, swap.From, swap.To)
			fmt.Fprintf(out, "  leaves:  %s\n  joins:   %s\n  result:  %s\n", strings.Join(swap.Removed, ", "), strings.Join(swap.Added, ", "), strings.Join(swap.Choices, ", "))
			if dry, _ := cmd.Flags().GetBool("dry-run"); dry {
				fmt.Fprintln(out, "Dry run: nothing was changed.")
				return nil
			}
			return problem.New(fmt.Sprintf("smith cannot move %s from %s to %s by itself yet", p.Name, swap.From, swap.To),
				"change its dependencies, config files and imports by hand, then run smith doctor to check the result",
				"or make a new project with smith add --with "+strings.Join(swap.Choices, ",")+" and move the code over")
		},
	}
	return operation.Wrap(cmd, operation.Spec{AddYes: true, AddDryRun: true, Describe: func(_ *cobra.Command, args []string) string {
		return "swap " + strings.Join(args, " ")
	}})
}
