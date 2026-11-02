package mcpserver

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/monosmith-app/monosmith/libs/actions/catalog"
	"github.com/spf13/cobra"
)

// Command is `smith mcp`.
func Command(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Serve the workspace to an AI assistant over the Model Context Protocol (stdio)",
		Long: `Serve the workspace to an AI assistant over the Model Context Protocol, on
stdin and stdout: tools to list the projects, graph, kinds, stacks and commands,
check the workspace, preview a release and the available updates, run tasks,
add and remove projects, install and remove dependencies, sync versions and add
a pipeline. Each tool runs the smith command it stands for. Register smith mcp
as the server's command, started in the workspace.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			self, err := os.Executable()
			if err != nil {
				return err
			}
			server := Server{Tools: Tools, Version: version, Run: func(args []string) Outcome {
				// The assistant's host asks the person before it calls a tool, which is the confirmation.
				args = catalog.WithConfirmation(args)
				c := exec.Command(self, args...)
				// Not a terminal and in CI mode: no command stops to ask.
				c.Env = append(os.Environ(), "CI=1", "NO_COLOR=1")
				var out bytes.Buffer
				c.Stdout, c.Stderr = &out, &out
				err := c.Run()
				text := strings.TrimSpace(out.String())
				if text == "" && err == nil {
					text = "done"
				}
				return Outcome{Text: "$ smith " + strings.Join(args, " ") + "\n" + text, IsError: err != nil}
			}}
			in := bufio.NewScanner(cmd.InOrStdin())
			in.Buffer(make([]byte, 1<<20), 16<<20)
			w := cmd.OutOrStdout()
			for in.Scan() {
				line := bytes.TrimSpace(in.Bytes())
				if len(line) == 0 {
					continue
				}
				if answer := server.Handle(line); answer != nil {
					if _, err := w.Write(append(answer, '\n')); err != nil {
						return err
					}
				}
			}
			return in.Err()
		},
	}
}
