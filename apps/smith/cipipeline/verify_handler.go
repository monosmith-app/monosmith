package cipipeline

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// VerifyCommand is `smith verify`: the verify phase of the pipelines, run where
// you are. It is the same phase as `smith ci verify`, so a green run here is the
// run CI makes: the dependency boundaries first, then lint, typecheck, test and
// build, and a check that nothing changed in the files git tracks.
func VerifyCommand(version string) *cobra.Command {
	var noCache bool
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Check the workspace as CI does: boundaries, lint, typecheck, test and build",
		Long: `Run the verify phase of the CI pipelines on this machine, so a green run here is the run CI makes:
the dependency boundaries first, then lint, typecheck, test and build of every project (only
the affected ones in a pull request build), and a check that no file git tracks was changed.

It is the same as smith ci verify. Results are replayed from the local cache when the
inputs did not change, and a clean CI agent has no cache: when a run passes here and
fails in CI, repeat it with --no-cache.`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: c.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			ctx, stop := signal.NotifyContext(c.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			run := AssetsRun{NoCache: noCache}
			run.In, run.Terminal = c.InOrStdin(), operationplan.DetectEnvironment().Terminal
			return RunPhaseFor(ctx, s, "verify", false, run, c.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&noCache, "no-cache", false, "run every task even if a cached result exists, as a clean CI agent does")
	return cmd
}
