package npmpolicy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/libs/plugin-node/pipeline"
	"github.com/spf13/cobra"
)

// Command is `smith npm-policy [--allow-scripts a,b]`.
func Command() *cobra.Command {
	var allow []string
	cmd := &cobra.Command{
		Use:    "npm-policy",
		Short:  "Check package-lock.json against the audit policy: integrity hashes, no git or http dependencies, install-script allow-list",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := workspace.FindRoot(".")
			if err != nil {
				return err
			}
			src, err := os.ReadFile(filepath.Join(root, "package-lock.json"))
			if err != nil {
				return err
			}
			violations, err := pipeline.CheckNpmPolicy(src, allow)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			for _, v := range violations {
				_, _ = fmt.Fprintln(out, v)
			}
			if len(violations) > 0 {
				return fmt.Errorf("npm policy: %d violation(s) in package-lock.json", len(violations))
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&allow, "allow-scripts", nil, "packages whose install scripts are accepted")
	return cmd
}
