package pydepscheck

import (
	"fmt"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/libs/plugin-python/declaredimports"
	"github.com/spf13/cobra"
)

// Command is `smith py-deps-check <project dir>`.
func Command() *cobra.Command {
	return &cobra.Command{
		Use:    "py-deps-check <project-dir>",
		Short:  "Check that a Python package declares every distribution it imports",
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
				return fmt.Errorf("%s imports %d module(s) its pyproject.toml does not declare where installing it installs them: add the distribution that provides each to dependencies", project, len(undeclared))
			}
			return nil
		},
	}
}
