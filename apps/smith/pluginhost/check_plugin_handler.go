package pluginhost

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/monosmith-app/monosmith/apps/smith/pluginenv"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/conformance"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/spf13/cobra"
)

// isDeclared is whether smith.toml declares an external plugin of that name,
// which then wins over a built-in one.
func isDeclared(external map[string]workspace.ExternalPlugin, name string) bool {
	_, ok := external[name]
	return ok
}

// explicitPath makes a file in the current directory (plugin.cmd) a path
// (.\plugin.cmd), so the program is started from where it is and not looked up
// on PATH, which Windows would refuse for the current directory.
func explicitPath(name string) string {
	if filepath.Base(name) != name {
		return name
	}
	return "." + string(filepath.Separator) + name
}

// checkCommand is `smith plugin check`: the conformance suite.
func checkCommand(version string) *cobra.Command {
	var generate, asJSON bool
	var grant []string
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
built-in plugins, then the plugin search path. A path runs that executable.

Checking is for the author of a plugin, so it does not need the plugin to be
trusted. The plugin still starts with the minimal environment external plugins
get (PATH, HOME, the temporary directory, locale and proxy variables): pass
--env NAME (or a prefix such as --env 'AWS_*') for each variable it needs. A
plugin found by name also gets the variables its plugin.json and smith.toml
declare. A built-in plugin keeps the whole environment.`,
		Example: "  smith plugin check kiwi --generate\n  smith plugin check ./bin/smith-plugin-kiwi\n  smith plugin check ./bin/smith-plugin-kiwi --env NPM_TOKEN --env 'AWS_*'",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var argv []string
			env := append(os.Environ(), "SMITH_PLUGIN=1")
			if _, err := os.Stat(args[0]); err == nil {
				argv = append([]string{explicitPath(args[0])}, args[1:]...)
				env = append(pluginenv.Build(os.Environ(), nil, grant, runtime.GOOS), "SMITH_PLUGIN=1")
			} else {
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
				if _, builtin := Builtin(args[0]); !builtin || isDeclared(external, args[0]) {
					var declared, workspaceGrant []string
					if p, err := ResolveExternal(root, external, args[0]); err == nil {
						declared, workspaceGrant = p.ManifestEnv, p.WorkspaceEnv
					}
					env = append(pluginenv.Build(os.Environ(), declared, append(workspaceGrant, grant...), runtime.GOOS), "SMITH_PLUGIN=1")
				}
			}
			ws, err := os.MkdirTemp("", "smith-plugin-check-")
			if err != nil {
				return err
			}
			defer func() { _ = os.RemoveAll(ws) }()
			c, err := host.Start(host.Options{
				Command: argv, Env: env, WorkspaceRoot: ws,
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
	cmd.Flags().StringArrayVar(&grant, "env", nil, "an environment variable the plugin may see, beyond the minimal environment: a name, or a prefix ending in * (repeatable)")
	return cmd
}
