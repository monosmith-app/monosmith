package languageserver

import (
	"bufio"
	"errors"
	"io"

	"github.com/spf13/cobra"
)

// Command is `smith lsp`.
func Command(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "lsp",
		Short: "Serve smith.toml to an editor over the Language Server Protocol (stdio)",
		Long: `Serve smith.toml to an editor over the Language Server Protocol, on stdin
and stdout: diagnostics for what smith would reject, completion of table names,
keys and values, and hover text from the schema (smith schema). Register
smith lsp as the language server for smith.toml.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			server := NewServer(version)
			in := bufio.NewReaderSize(cmd.InOrStdin(), 1<<16)
			out := cmd.OutOrStdout()
			for !server.Exited() {
				body, err := ReadFrame(in)
				if errors.Is(err, io.EOF) {
					return nil
				}
				if err != nil {
					return err
				}
				for _, answer := range server.Handle(body) {
					if err := WriteFrame(out, answer); err != nil {
						return err
					}
				}
			}
			if !server.CleanExit() {
				return errors.New("lsp: exit before shutdown")
			}
			return nil
		},
	}
}
