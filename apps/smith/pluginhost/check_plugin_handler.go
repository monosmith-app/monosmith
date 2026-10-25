package pluginhost

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/conformance"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/spf13/cobra"
)

// checkCommand is `smith plugin check`: the conformance suite.
func checkCommand(version string) *cobra.Command {
	var generate, asJSON bool
	cmd := &cobra.Command{
		Use:   "check <name | path to plugin> [args...]",
		Short: "Check that a plugin keeps the protocol's promises",
		Long: `Runs the conformance suite against a plugin: it names itself, its roles
match what it implements, its stacks are complete, and every method it
advertises answers. It also checks how the plugin wants to be called: requests
sent together are answered, one at a time unless the plugin declares a
concurrency, and a plugin that advertises cancel stops a cancelled request
within a bound and goes on serving. With --generate, a plugin that generates projects makes
an app, a lib and a package in a temporary directory, which a language plugin
must then discover as generated (this may install dependencies).

A name is looked up the way the workspace would start it: smith.toml, then the
built-in plugins, then the plugin search path. A path runs that executable.`,
		Example: "  smith plugin check kiwi --generate\n  smith plugin check ./bin/smith-plugin-kiwi",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			argv := args
			if _, err := os.Stat(args[0]); err != nil {
				root, external := ".", map[string]workspace.ExternalPlugin{}
				if r, err := workspace.FindRoot("."); err == nil {
					root = r
					if cfg, err := workspace.Load(r); err == nil {
						external = cfg.Plugins.External
					}
				}
				if argv, err = LaunchCommand(root, external, args[0]); err != nil {
					return err
				}
			}
			ws, err := os.MkdirTemp("", "smith-plugin-check-")
			if err != nil {
				return err
			}
			defer func() { _ = os.RemoveAll(ws) }()
			c, err := host.Start(host.Options{
				Command: argv, Env: append(os.Environ(), "SMITH_PLUGIN=1"), WorkspaceRoot: ws,
				Stderr: cmd.ErrOrStderr(), Host: protocol.Peer{Name: "smith", Version: version},
			})
			if err != nil {
				return fmt.Errorf("the plugin did not start and initialize: %w", err)
			}
			opts := conformance.Options{}
			if generate {
				opts.Workspace = ws
			}
			report := conformance.Check(c, opts)
			closeErr := c.Close()
			report.Results = append(report.Results, conformance.Result{Check: "shuts down when asked", OK: closeErr == nil, Detail: fmt.Sprint(closeErr)})
			out := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				if err := enc.Encode(report); err != nil {
					return err
				}
			} else {
				fmt.Fprintf(out, "%s %s\n", report.Plugin.Name, report.Plugin.Version)
				for _, r := range report.Results {
					mark := "✔"
					if !r.OK {
						mark = "✖"
					}
					fmt.Fprintf(out, "  %s %s\n", mark, r.Check)
					if !r.OK {
						fmt.Fprintf(out, "      %s\n", r.Detail)
					}
				}
			}
			if !report.OK() {
				return errors.New("the plugin does not conform")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&generate, "generate", false, "also generate projects in a temporary directory and discover them")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON")
	return cmd
}
