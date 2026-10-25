package pluginhost

import (
	"bufio"
	"fmt"
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/pluginadmit"
	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// trustCommands are `smith plugin trust` and `smith plugin untrust`.
func trustCommands() []*cobra.Command {
	var yes, pin bool
	trust := &cobra.Command{
		Use:   "trust <name>",
		Short: "Review an external plugin and trust it, so smith may start it",
		Long: `Shows an external plugin the workspace uses (where it comes from, its command,
the SHA-256 of the files that command runs and the environment variables it
may see) and, once you say yes, records that you trust it in your smith home
(SMITH_HOME, or ~/.smith). The decision is for this workspace, this plugin
name, this command line and these files: if the command or a file changes,
smith asks again.

A trusted plugin runs with your privileges; smith does not sandbox it. Read
the plugin's code before you trust it.

Without a terminal it refuses unless --yes is given. With --pin it also writes
the plugin's sha256 to smith.toml, which is what CI checks: in CI an external
plugin starts only when smith.toml pins its digest and the file still has it.`,
		Example: "  smith plugin trust widget\n  smith plugin trust widget --pin\n  smith plugin trust widget --yes",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			root, cfg, err := loadWorkspace()
			if err != nil {
				return err
			}
			p, err := resolveForTrust(root, cfg, name)
			if err != nil {
				return err
			}
			store, err := plugintrust.OpenStore()
			if err != nil {
				return err
			}
			st := pluginadmit.Describe(store, root, []pluginadmit.Plugin{p})[0]
			if st.Problem != "" {
				return pluginadmit.Refusal(st, false)
			}
			out := cmd.OutOrStdout()
			if !recordedTrust(store, root, p) {
				if !yes {
					if !term.IsTerminal(int(os.Stdin.Fd())) {
						return problem.New("refusing to trust a plugin without a terminal to ask in",
							"read what it is with `smith plugin list`, then run `smith plugin trust "+name+" --yes`")
					}
					if !pluginadmit.Ask(&pluginadmit.Prompter{Out: out, In: bufio.NewReader(os.Stdin)}, st) {
						return fmt.Errorf("%s was not trusted", name)
					}
				}
				if err := store.Trust(root, name, st.Fingerprint); err != nil {
					return err
				}
				fmt.Fprintf(out, "Trusted %s in %s.\n", name, root)
			} else {
				fmt.Fprintf(out, "%s is already trusted in %s.\n", name, root)
			}
			if pin {
				digest := plugintrust.PinnedDigest(st.Fingerprint)
				table := "plugins.external." + configedit.QuoteKey(name)
				if err := configedit.SetTableValue(root, table, "sha256", digest); err != nil {
					return err
				}
				fmt.Fprintf(out, "Pinned sha256 %s in smith.toml [%s], which is what CI checks.\n", digest, table)
			}
			return nil
		},
	}
	trust.Flags().BoolVarP(&yes, "yes", "y", false, "trust without asking")
	trust.Flags().BoolVar(&pin, "pin", false, "also write the plugin's sha256 to smith.toml, for CI")
	untrust := &cobra.Command{
		Use:     "untrust <name>",
		Short:   "Forget that you trust an external plugin",
		Long:    "Removes your trust decision for an external plugin in this workspace. smith asks again the next time the workspace needs the plugin. A sha256 pinned in smith.toml is not touched.",
		Example: "  smith plugin untrust widget",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, _, err := loadWorkspace()
			if err != nil {
				return err
			}
			store, err := plugintrust.OpenStore()
			if err != nil {
				return err
			}
			removed, err := store.Untrust(root, args[0])
			if err != nil {
				return err
			}
			if !removed {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s was not trusted in %s.\n", args[0], root)
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "No longer trusting %s in %s.\n", args[0], root)
			return err
		},
	}
	return []*cobra.Command{trust, untrust}
}

// recordedTrust is whether the user's own decision, as it stands, trusts the
// plugin (a CI pin does not count here).
func recordedTrust(store *plugintrust.Store, root string, p pluginadmit.Plugin) bool {
	d, err := plugintrust.Check(store, plugintrust.Request{Workspace: root, Plugin: p.Name, Command: p.Command})
	return err == nil && d.Trusted
}

func loadWorkspace() (string, workspace.Config, error) {
	root, err := workspace.FindRoot(".")
	if err != nil {
		return "", workspace.Config{}, err
	}
	cfg, err := workspace.Load(root)
	return root, cfg, err
}

// resolveForTrust finds the external plugin a name stands for, or says why
// there is nothing to trust.
func resolveForTrust(root string, cfg workspace.Config, name string) (pluginadmit.Plugin, error) {
	if _, declared := cfg.Plugins.External[name]; !declared {
		if _, builtin := Builtin(name); builtin {
			return pluginadmit.Plugin{}, problem.New(fmt.Sprintf("%s is a built-in plugin: it is the smith binary itself, so there is nothing to trust", name))
		}
	}
	p, err := ResolveExternal(root, cfg.Plugins.External, name)
	if err != nil {
		return pluginadmit.Plugin{}, problem.New(err.Error(), "run `smith plugin list` to see the plugins smith can find")
	}
	return p, nil
}
