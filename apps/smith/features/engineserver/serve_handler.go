package engineserver

import (
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

// Command is `smith serve`.
func Command(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Serve the engine to an editor or a terminal interface over JSON-RPC (stdio)",
		Long: `Serve the engine over JSON-RPC on stdin and stdout, one JSON message per line,
in the workspace smith is started in. It opens no network port.

A method per function: projects, graph, affected, doctor, release.preview,
explain, plugins, run, cancel, install, uninstall, add, upgrade and more; the
answer of a query is the JSON its smith command prints with --json. A run's
events (run.started, task.started, task.output, task.finished, run.finished)
arrive as "event" notifications while it goes, and the run method answers when
it ends. Stopping the client, or the cancel method, stops the run's processes.
The protocol is in docs/reference/serve.md.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
			defer stop()
			server := &Server{Version: version}
			return server.Serve(ctx, cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
}
