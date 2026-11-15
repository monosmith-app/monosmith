package npmdepscheck

import (
	"fmt"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
	"github.com/monosmith-app/monosmith/libs/plugin-node/declaredimports"
	"github.com/spf13/cobra"
)

// Command is `smith npm-deps-check <project dir>`.
func Command() *cobra.Command {
	return &cobra.Command{
		Use:    "npm-deps-check <project-dir>",
		Short:  "Check that a Node package declares every package it imports",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := workspace.FindRoot(".")
			if err != nil {
				return err
			}
			project := filepath.ToSlash(args[0])
			undeclared, err := declaredimports.Check(root, project)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			for _, u := range undeclared {
				_, _ = fmt.Fprintln(out, u.String())
			}
			if len(undeclared) > 0 {
				return fmt.Errorf("%s imports %d package(s) its package.json does not declare where installing it installs them: add them to dependencies (peerDependencies when the installing app provides them)", project, len(undeclared))
			}
			return nil
		},
	}
}
