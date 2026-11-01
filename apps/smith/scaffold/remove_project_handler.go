package scaffold

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/commandargs"
	"github.com/monosmith-app/monosmith/apps/smith/operation"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// RemoveCommand is `smith remove`.
func RemoveCommand(version string) *cobra.Command {
	var opts RemoveOptions
	var yes, asJSON bool
	cmd := &cobra.Command{
		Use:   "remove <project>",
		Short: "Remove a project: its directory, its traces in the workspace, its CI",
		Long: `Remove a project: delete its directory, let its plugin clean what it owns
outside it (the go.work entry, the npm lockfile, the Python venv), drop its
tables from smith.toml and regenerate the pipelines. A project that others
import is refused unless --force. Its release tags are kept.

Afterwards it lists the tracked files that still name the project (a
hand-written workflow, an editor task, a doc). It does not change them.`,
		Example: "  smith remove web --dry-run\n  smith remove web",
		Args:    commandargs.Exactly(commandargs.Spec{Names: []string{"the project to remove"}, Example: "smith remove web --dry-run", More: "smith projects lists them"}),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			out := cmd.OutOrStdout()
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			if !opts.DryRun {
				if s.Graph.Get(opts.Name) == nil {
					return fmt.Errorf("no project named %q", opts.Name)
				}
				if !yes && !term.IsTerminal(int(os.Stdin.Fd())) {
					return errors.New("refusing to remove without a terminal; pass --yes")
				}
				// The dry run is the plan: it also refuses a project others import.
				preview, err := RemoveProject(s, RemoveOptions{Name: opts.Name, Force: opts.Force, DryRun: true, SkipMentions: true})
				if err != nil {
					return err
				}
				if _, err := operationplan.Present(planOut(cmd, asJSON), cmd.InOrStdin(), PlanRemoval(s.Root, preview), operationplan.Options{
					Yes: yes, Ask: true, Env: operationplan.DetectEnvironment(),
				}); err != nil {
					return err
				}
			}
			r, err := RemoveProject(s, opts)
			if err != nil {
				return err
			}
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(r)
			}
			verb := "Removed"
			if opts.DryRun {
				verb = "Would remove"
			}
			fmt.Fprintf(out, "%s %s (%s)\n", verb, r.Project, r.Root)
			if len(r.Dependents) > 0 {
				fmt.Fprintf(out, "  still imported by %s: their builds will break until the imports go\n", strings.Join(r.Dependents, ", "))
			}
			for _, f := range r.Files {
				fmt.Fprintf(out, "  changes %s\n", f)
			}
			for _, n := range r.Notes {
				fmt.Fprintf(out, "  %s\n", n)
			}
			WriteMentions(out, r)
			return nil
		},
	}
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "show what would change; change nothing")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "remove a project other projects still import")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "do not ask for confirmation")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print what changed as JSON")
	return operation.Wrap(cmd, operation.Spec{Describe: func(_ *cobra.Command, args []string) string { return "remove " + args[0] }})
}

// planOut is where the plan goes: stderr when stdout carries JSON.
func planOut(cmd *cobra.Command, asJSON bool) io.Writer {
	if asJSON {
		return cmd.ErrOrStderr()
	}
	return cmd.OutOrStdout()
}
