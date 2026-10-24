package declarationspecifiers

import (
	"github.com/spf13/cobra"
)

// Name is the command's name; a converted build calls `smith <Name> <dir>`.
const Name = "declaration-specifiers"

// Command is `smith declaration-specifiers <declarations dir>`.
func Command() *cobra.Command {
	return &cobra.Command{
		Use:    Name + " <declarations-dir>",
		Short:  "Give the relative specifiers in generated .d.ts files the file they name (./hooks becomes ./hooks.js), for nodenext consumers",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			_, err := Run(args[0])
			return err
		},
	}
}
