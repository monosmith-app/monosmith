package scaffold

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/commandargs"
	"github.com/monosmith-app/monosmith/apps/smith/operation"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
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

A folder that holds files git cannot bring back (untracked, or changed and not
committed) is never deleted by surprise: smith lists them and asks whether to
move the folder to .smith/removed/<name>-<time>/ (the default), delete it for
good, or cancel. Without a terminal it refuses unless --keep-copy or --delete
is given; --yes alone does not choose.

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
				if opts.KeepCopy && opts.Delete {
					return errors.New("--keep-copy and --delete say opposite things; use one")
				}
				if !yes && !operationplan.DetectEnvironment().Terminal {
					return errors.New("refusing to remove without a terminal; pass --yes")
				}
				// The dry run is the plan: it also refuses a project others import.
				preview, err := RemoveProject(s, RemoveOptions{Name: opts.Name, Force: opts.Force, DryRun: true, SkipMentions: true, KeepCopy: opts.KeepCopy})
				if err != nil {
					return err
				}
				plan := PlanRemoval(s.Root, preview)
				planned := planOut(cmd, asJSON)
				if len(preview.Unsaved) > 0 && !opts.KeepCopy && !opts.Delete {
					// Deleting would lose files git cannot bring back: smith asks, and --yes
					// does not answer for the person.
					operationplan.Render(planned, plan)
					choice, err := chooseUnsaved(preview, operationplan.ShouldAsk(operationplan.DetectEnvironment(), yes), bufio.NewReader(cmd.InOrStdin()), planned)
					if err != nil {
						return err
					}
					opts.KeepCopy, opts.Delete = choice == removeKeepCopy, choice == removeDelete
					yes = true // the choice was the confirmation
				}
				if opts.KeepCopy {
					preview.KeptAt = ".smith/removed/" + opts.Name + "-<time>"
					plan = PlanRemoval(s.Root, preview)
				}
				if _, err := operationplan.Present(planned, cmd.InOrStdin(), plan, operationplan.Options{
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
			if r.KeptAt != "" {
				fmt.Fprintf(out, "  the folder was kept at %s\n", r.KeptAt)
			} else if len(r.Unsaved) > 0 {
				fmt.Fprintf(out, "  %d file(s) git cannot bring back are in the folder: --keep-copy keeps the folder, --delete deletes it\n", len(r.Unsaved))
			}
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
	cmd.Flags().BoolVar(&opts.KeepCopy, "keep-copy", false, "move the project folder to .smith/removed/<name>-<time>/ instead of deleting it")
	cmd.Flags().BoolVar(&opts.Delete, "delete", false, "delete the folder for good, including files git cannot bring back")
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
