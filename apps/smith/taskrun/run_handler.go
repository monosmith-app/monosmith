package taskrun

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/monosmith-app/monosmith/apps/smith/runevents"
	"github.com/spf13/cobra"
)

// ErrTasksFailed makes smith exit non-zero after printing the summary.
var ErrTasksFailed = errors.New("some tasks did not succeed")

// Commands returns `smith run` and its shortcuts build, test, lint, typecheck.
func Commands(version string) []*cobra.Command {
	run := command(version, "run <task>...", "Run tasks on projects, in dependency order, from cache when nothing changed", nil)
	run.Args = cobra.MinimumNArgs(1)
	cmds := []*cobra.Command{run}
	for _, task := range []string{"build", "test", "lint", "typecheck"} {
		cmds = append(cmds, command(version, task, "Run the "+task+" task (same as `smith run "+task+"`)", []string{task}))
	}
	return cmds
}

// textOutput is where task output and progress lines go. With --json stdout
// carries the summary and nothing else, so the text moves to stderr.
func textOutput(stdout, stderr io.Writer, asJSON bool) io.Writer {
	if asJSON {
		return stderr
	}
	return stdout
}

func command(version, use, short string, fixed []string) *cobra.Command {
	var (
		req    Request
		asJSON bool
		// jsonEvents prints the event stream as JSON lines instead of text.
		jsonEvents bool
	)
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Example: `  smith build                       every project
  smith test -p web,ui              two projects (and what they depend on)
  smith run lint test --affected    only what changed since main (or the PR target)
  smith build -p kind:package       every publishable package`,
		RunE: func(cmd *cobra.Command, args []string) error {
			req.Tasks = fixed
			if fixed == nil {
				req.Tasks = args
			} else if len(args) > 0 {
				req.Projects = append(req.Projects, args...)
			}
			req.Out = textOutput(cmd.OutOrStdout(), cmd.ErrOrStderr(), asJSON)
			if jsonEvents {
				// Stdout carries the event stream and nothing else: each task's
				// output arrives as task.output events.
				req.Out = io.Discard
				req.Events = runevents.NewJSONLines(cmd.OutOrStdout())
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
			defer stop()
			summary, err := RunInDirectory(ctx, ".", version, cmd.ErrOrStderr(), req)
			if err != nil {
				return err
			}
			if jsonEvents {
				if !summary.OK() {
					return ErrTasksFailed
				}
				return nil
			}
			if len(summary.Results) == 0 {
				return nil
			}
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				_ = enc.Encode(summary)
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), summary.String())
			}
			if !summary.OK() {
				return ErrTasksFailed
			}
			return nil
		},
	}
	if fixed != nil {
		cmd.Use = use + " [project]..."
	}
	f := cmd.Flags()
	f.StringSliceVarP(&req.Projects, "projects", "p", nil, "projects to run: names, globs, kind:app, lang:go, tag:x, !exclude")
	f.BoolVar(&req.Affected, "affected", false, "only projects affected by changes since --base")
	f.StringVar(&req.Base, "base", "", "base branch for --affected (default: the PR target, else the release branch)")
	f.IntVar(&req.Parallel, "parallel", 0, "tasks to run at once (default: CPU count)")
	f.BoolVar(&req.Bail, "bail", false, "stop starting tasks after the first failure")
	f.BoolVar(&req.NoCache, "no-cache", false, "run every task even if a cached result exists")
	f.BoolVar(&asJSON, "json", false, "print the summary as JSON")
	f.BoolVar(&jsonEvents, "json-events", false, "print the run as JSON lines, one event each (see docs/reference/serve.md), and nothing else")
	cmd.MarkFlagsMutuallyExclusive("json", "json-events")
	return cmd
}
