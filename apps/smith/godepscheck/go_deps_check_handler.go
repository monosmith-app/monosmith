package godepscheck

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/libs/plugin-go/declaredimports"
	"github.com/spf13/cobra"
)

// shownDiffLines caps how much of tidy's diff the task prints.
const shownDiffLines = 40

// Command is `smith go-deps-check <module dir>`.
func Command() *cobra.Command {
	return &cobra.Command{
		Use:    "go-deps-check <module-dir>",
		Short:  "Check that a Go module's go.mod matches what the module imports",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := workspace.FindRoot(".")
			if err != nil {
				return err
			}
			module := filepath.ToSlash(args[0])
			result, err := declaredimports.Check(root, module)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "checked with: %s\n", result.Method)
			for _, issue := range result.Issues {
				_, _ = fmt.Fprintln(out, issue.String())
			}
			if result.Diff != "" {
				lines := strings.Split(strings.TrimRight(result.Diff, "\n"), "\n")
				if len(lines) > shownDiffLines {
					lines = append(lines[:shownDiffLines], fmt.Sprintf("... and %d more lines", len(lines)-shownDiffLines))
				}
				_, _ = fmt.Fprintln(out, strings.Join(lines, "\n"))
			}
			if !result.Clean() {
				return fmt.Errorf("%s/go.mod does not match what the module imports: run `go mod tidy` in %s (a requirement for a module in go.work is added by hand)", module, module)
			}
			return nil
		},
	}
}
