package sandbox

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// Command is `smith try`.
func Command() *cobra.Command {
	var checks []string
	var keep, asJSON, yes bool
	cmd := &cobra.Command{
		Use:   "try <command> [args...]",
		Short: "Try a command that changes files in a temporary copy, check the result, then offer to run it for real",
		Long: `Run a command that changes files (` + strings.Join(Allowed, ", ") + `) in a temporary copy of
the work tree, run doctor on the result, and say what changed. Your files are
not touched. Then smith offers to run the same command for real.

The copy is a git worktree of HEAD with your uncommitted changes laid over it;
folders git ignores (node_modules, dist) are not copied, and the task cache is
shared with your tree. Dependencies are not installed in the copy, so ask for
build or test only for a change that does not need them to be installed again.

In CI, or without a terminal, it never asks: it prints the command to run.
--yes runs it for real when the command and every check passed.`,
		Example: "  smith try add package billing\n  smith try up --checks doctor,build\n  smith try upgrade --yes",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := workspace.FindRoot(".")
			if err != nil {
				return err
			}
			cfg, err := workspace.Load(root)
			if err != nil {
				return err
			}
			env := []string{
				"SMITH_CACHE_DIR=" + session.LocalCacheDir(root, cfg, os.Getenv),
				"SMITH_NO_UPDATE_CHECK=1",
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Try: running smith %s in a temporary copy (your files are not touched).\n", strings.Join(args, " "))
			res, err := Try(Options{Root: root, Command: args, Checks: checks, Env: env, Keep: keep, Run: ExecRunner})
			if err != nil {
				return err
			}
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(res)
			}
			Report(out, res)
			if !res.OK() {
				return errors.New("the try failed: nothing was changed in your files")
			}
			realCmd := "smith " + strings.Join(args, " ")
			env2 := operationplan.DetectEnvironment()
			if !yes && !operationplan.ShouldAsk(env2, false) {
				fmt.Fprintf(out, "→ To apply it for real, run: %s\n", realCmd)
				return nil
			}
			if !yes {
				plan := operationplan.Plan{Title: "apply for real: " + realCmd}
				for _, line := range res.Changed {
					plan.Add("Files the try changed", operationplan.Line{Action: operationplan.Change, Path: strings.TrimSpace(line[min(3, len(line)):])})
				}
				ok, err := operationplan.Present(out, cmd.InOrStdin(), plan, operationplan.Options{Ask: true, Env: env2})
				if err != nil || !ok {
					return err
				}
			}
			output, code, err := ExecRunner(root, []string{"SMITH_NO_UPDATE_CHECK=1"}, args...)
			fmt.Fprint(out, output)
			if err != nil {
				return err
			}
			if code != 0 {
				return fmt.Errorf("%s failed with exit code %d", realCmd, code)
			}
			return nil
		},
	}
	cmd.Flags().SetInterspersed(false)
	cmd.Flags().StringSliceVar(&checks, "checks", []string{"doctor"}, "checks to run in the copy after the command: "+strings.Join(Checks, ", "))
	cmd.Flags().BoolVar(&keep, "keep", false, "leave the copy on disk and print its folder")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the result as JSON and do not apply")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "run the command for real when the try passed, without asking")
	return cmd
}

// Report prints what a try found: the command, the files it changed and each
// check, with the cause and the remedy for a failure.
func Report(out interface{ Write([]byte) (int, error) }, r Result) {
	w := func(format string, a ...any) { fmt.Fprintf(out, format, a...) }
	if !r.CommandOK {
		w("✗ smith %s failed in the copy:\n%s\n", strings.Join(r.Command, " "), indent(r.CommandTail))
		w("→ fix the cause, then try again. Your files were not touched.\n")
		return
	}
	w("✓ smith %s ran in the copy: %d files changed\n", strings.Join(r.Command, " "), len(r.Changed))
	for i, l := range r.Changed {
		if i == 20 {
			w("  … %d more\n", len(r.Changed)-20)
			break
		}
		w("  %s\n", l)
	}
	for _, c := range r.Checks {
		if c.OK {
			w("✓ %s passed on the result\n", c.Name)
			continue
		}
		w("✗ %s failed on the result:\n%s\n", c.Name, indent(c.Tail))
		w("→ try again with --keep to look at the copy. Your files were not touched.\n")
	}
	if r.Kept != "" {
		w("The copy is kept in %s (remove it with: git worktree remove --force %s)\n", r.Kept, r.Kept)
	}
}

func indent(s string) string { return "  " + strings.ReplaceAll(s, "\n", "\n  ") }
