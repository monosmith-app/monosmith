package npmaudit

import (
	"errors"
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
	"github.com/monosmith-app/monosmith/libs/plugin-node/pipeline"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
	"github.com/spf13/cobra"
)

// Command is `smith npm-audit`.
func Command() *cobra.Command {
	return &cobra.Command{
		Use:    "npm-audit",
		Short:  "Audit the workspace's npm dependencies, failing only on a production advisory npm can fix",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := workspace.FindRoot(".")
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			npm := func(args ...string) string {
				// npm audit exits non-zero whenever it finds something; the
				// report on stdout is what counts.
				stdout, _, _ := toolrun.Capture(root, append([]string{"npm"}, args...))
				return stdout
			}
			if !pipeline.AuditNpm(npm, func(line string) { _, _ = fmt.Fprintln(out, line) }) {
				return errors.New("npm audit found production advisories with a fix")
			}
			return nil
		},
	}
}
