package lintfix

import (
	"errors"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

// ErrProblemsLeft makes the command exit 1 once its report is written, so the
// closing step of a conversion shows the report as a finding. The report says
// everything: smith prints nothing more for it.
var ErrProblemsLeft = errors.New("ESLint left problems")

// Lister finds the workspace that holds dir and the Node projects in it that
// have a lint task. It returns the workspace root and the projects.
type Lister func(dir string) (root string, targets []Target, err error)

// Command is `smith eslint-fix`: ESLint fixes what it can in every Node
// project that has a lint task, then the problems left are counted per
// project. Hidden: the closing checks of a full conversion with smith's lint
// run it. It exits 1 when a problem is left or ESLint could not run.
//
// Parameters: list finds the projects (the command line opens the workspace
// for it, so this package never does).
func Command(list Lister) *cobra.Command {
	return &cobra.Command{
		Use:           "eslint-fix",
		Short:         "Fix what ESLint can in every Node project, then count the problems left",
		Hidden:        true,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, targets, err := list(".")
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(targets) == 0 {
				_, err := fmt.Fprintln(out, "Nothing to fix: no Node project has a lint task")
				return err
			}
			sort.Slice(targets, func(i, j int) bool { return targets[i].Name < targets[j].Name })
			text, failed := Report(Fix(Options{Root: root, Targets: targets, Run: runProcess, ESLintVersion: installedESLint(root)}))
			fmt.Fprint(out, text)
			if failed {
				return ErrProblemsLeft
			}
			return nil
		},
	}
}
