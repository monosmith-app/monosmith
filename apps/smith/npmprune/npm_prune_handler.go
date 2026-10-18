package npmprune

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/libs/plugin-node/pruning"
	"github.com/spf13/cobra"
)

// Command is `smith npm-prune <project dir> [--out <dir>]`.
func Command() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:    "npm-prune <project-dir>",
		Short:  "Write a Node app's deploy output into its build directory: its production package.json and lockfile, and the workspace packages it uses",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := workspace.FindRoot(".")
			if err != nil {
				return err
			}
			res, err := pruning.Prune(pruning.Request{WorkspaceRoot: root, Root: filepath.ToSlash(args[0]), Out: filepath.ToSlash(out)})
			if err != nil {
				return err
			}
			modules := "no workspace packages"
			if len(res.Modules) > 0 {
				modules = strings.Join(res.Modules, ", ")
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "pruned %s into %s (workspace_modules: %s)\n", args[0], res.Out, modules)
			return err
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "the build directory to write into, workspace-relative (default dist/<project dir>)")
	return cmd
}
