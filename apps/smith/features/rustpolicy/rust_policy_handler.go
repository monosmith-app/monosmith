package rustpolicy

import (
	"errors"
	"fmt"

	"github.com/monosmith-app/monosmith/libs/plugin-rust/pipeline"
	"github.com/spf13/cobra"
)

// Command is `smith rust-policy [--max-unsafe n] dir...`.
func Command() *cobra.Command {
	var budget int
	cmd := &cobra.Command{
		Use:    "rust-policy dir...",
		Short:  "Check Rust sources against the audit policy: a budget of unsafe blocks",
		Hidden: true,
		Args:   cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, dirs []string) error {
			violations, err := pipeline.CheckRustPolicy(dirs, budget)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			for _, v := range violations {
				_, _ = fmt.Fprintln(out, v)
			}
			if len(violations) > 0 {
				return errors.New("rust policy: unsafe count over the budget")
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&budget, "max-unsafe", 0, "unsafe blocks, functions, impls and traits allowed in all")
	return cmd
}
