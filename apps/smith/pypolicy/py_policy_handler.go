package pypolicy

import (
	"fmt"

	"github.com/monosmith-app/monosmith/libs/plugin-python/pipeline"
	"github.com/spf13/cobra"
)

// Command is `smith py-policy [--index url] requirements-file...`.
func Command() *cobra.Command {
	var index string
	cmd := &cobra.Command{
		Use:    "py-policy [requirements-file...]",
		Short:  "Check requirements files against the audit policy: every pin hashed, no direct URLs, only the expected index",
		Hidden: true,
		RunE: func(cmd *cobra.Command, files []string) error {
			violations, err := pipeline.CheckPythonPolicyFiles(files, index)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			for _, v := range violations {
				_, _ = fmt.Fprintln(out, v)
			}
			if len(violations) > 0 {
				return fmt.Errorf("python policy: %d violation(s)", len(violations))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&index, "index", "", "the only index requirements may use (default: PyPI)")
	return cmd
}
