package bootstrap

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/monosmith-app/monosmith/apps/smith/operation"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// Command is `smith bootstrap`.
func Command() *cobra.Command {
	var check, dryRun bool
	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Write the launcher that runs the smith version the workspace pins",
		Long: `Write tools/smith/smith (sh) and tools/smith/smith.cmd (Windows): a launcher
you commit, so a laptop and CI run the same smith. It reads [ci] smith_version
from smith.toml each time it runs. With a pin it runs that version: the smith on
PATH when it matches, else a copy under .smith/bin/<version>, installed with
tools/smith/install.sh when the workspace has one, or with npx. Without a pin,
or with "latest", it runs the smith on PATH.

It changes no setting. Run it again to refresh the files after upgrading smith;
--check fails when they are missing or out of date, for CI.

smith shows what it will write first, with a diff of a launcher that exists and
differs, and asks; without a terminal it needs --yes. --dry-run shows the plan.`,
		Example: "  smith bootstrap\n  tools/smith/smith doctor",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := workspace.FindRoot(".")
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if !check {
				plan, err := PlanShims(root)
				if err != nil {
					return err
				}
				yes, _ := cmd.Flags().GetBool("yes")
				goOn, err := operationplan.Present(out, cmd.InOrStdin(), plan, operationplan.Options{Yes: yes, DryRun: dryRun, Ask: true, Env: operationplan.DetectEnvironment()})
				if err != nil || !goOn {
					if err == nil && dryRun && plan.Empty() {
						fmt.Fprintln(out, "The launcher is up to date.")
					}
					return err
				}
			}
			changed, err := WriteShims(root, check)
			if err != nil {
				return err
			}
			if len(changed) == 0 {
				fmt.Fprintln(out, "The launcher is up to date.")
				return nil
			}
			verb := "wrote"
			if check {
				verb = "would write"
			}
			for _, p := range changed {
				fmt.Fprintf(out, "%s %s\n", verb, p)
			}
			if check {
				return errors.New("the launcher is missing or out of date: run smith bootstrap")
			}
			if runtime.GOOS == "windows" {
				fmt.Fprintf(out, "Windows does not keep the executable bit. After git add, run: git update-index --chmod=+x %s\n", ShellPath)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "change nothing; fail when the launcher is missing or out of date")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be written, with a diff of a launcher that exists, and change nothing")
	return operation.Wrap(cmd, operation.Spec{AddYes: true, Describe: func(*cobra.Command, []string) string { return "write the smith launcher" }})
}
