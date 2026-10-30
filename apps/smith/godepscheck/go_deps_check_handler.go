package godepscheck

import (
	"fmt"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/libs/plugin-go/declaredimports"
	"github.com/spf13/cobra"
)

// Command is `smith go-deps-check <project dir>`.
func Command() *cobra.Command {
	return &cobra.Command{
		Use:    "go-deps-check <project-dir>",
		Short:  "Check that a Go module requires every module it imports",
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
				return fmt.Errorf("%s imports %d package(s) from modules its go.mod does not require where fetching it fetches them: run go mod tidy in %s", project, len(undeclared), project)
			}
			return nil
		},
	}
}
