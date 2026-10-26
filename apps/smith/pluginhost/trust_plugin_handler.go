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
			review, err := ReviewForTrust(root, cfg, name)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if !review.AlreadyTrusted && !yes {
				if !term.IsTerminal(int(os.Stdin.Fd())) {
					return problem.New("refusing to trust a plugin without a terminal to ask in",
						"read what it is with `smith plugin list`, then run `smith plugin trust "+name+" --yes`")
				}
				if !pluginadmit.Ask(&pluginadmit.Prompter{Out: out, In: bufio.NewReader(os.Stdin)}, review.Status) {
					return fmt.Errorf("%s was not trusted", name)
				}
			}
			recorded, digest, err := RecordTrust(root, review, pin)
			if err != nil {
				return err
			}
			if recorded {
				fmt.Fprintf(out, "Trusted %s in %s.\n", name, root)
			} else {
				fmt.Fprintf(out, "%s is already trusted in %s.\n", name, root)
			}
			if pin {
				fmt.Fprintf(out, "Pinned sha256 %s in smith.toml [%s], which is what CI checks.\n", digest, PinTable(name))
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
			removed, err := ForgetTrust(root, args[0])
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

// TrustReview is an external plugin ready for the person to decide on: what
// would run, as `smith plugin trust` shows it before it asks.
type TrustReview struct {
	// Status is the plugin as it is now: source, command, the files it runs
	// with their digests, what it may see.
	Status pluginadmit.Status
	// AlreadyTrusted is true when the person's own recorded decision still
	// matches (a CI pin does not count).
	AlreadyTrusted bool
	store          *plugintrust.Store
}

// ReviewForTrust resolves and measures the external plugin name for a trust
// decision. It refuses, with the same problem `smith plugin trust` prints, a
// built-in plugin, an unknown name and a command that cannot be measured.
func ReviewForTrust(root string, cfg workspace.Config, name string) (TrustReview, error) {
	p, err := resolveForTrust(root, cfg, name)
	if err != nil {
		return TrustReview{}, err
	}
	store, err := plugintrust.OpenStore()
	if err != nil {
		return TrustReview{}, err
	}
	st := pluginadmit.Describe(store, root, []pluginadmit.Plugin{p})[0]
	if st.Problem != "" {
		return TrustReview{}, pluginadmit.Refusal(st, false)
	}
	return TrustReview{Status: st, AlreadyTrusted: recordedTrust(store, root, p), store: store}, nil
}

// RecordTrust records the person's yes for a reviewed plugin in the smith
// home and, with pin, writes its sha256 to smith.toml for CI. recorded is
// false when the decision was already on file; digest is empty without pin.
func RecordTrust(root string, r TrustReview, pin bool) (recorded bool, digest string, err error) {
	if !r.AlreadyTrusted {
		if err = r.store.Trust(root, r.Status.Name, r.Status.Fingerprint); err != nil {
			return false, "", err
		}
		recorded = true
	}
	if pin {
		digest = plugintrust.PinnedDigest(r.Status.Fingerprint)
		if err = configedit.SetTableValue(root, PinTable(r.Status.Name), "sha256", digest); err != nil {
			return recorded, "", err
		}
	}
	return recorded, digest, nil
}

// PinTable is the smith.toml table that pins a plugin's digest.
func PinTable(name string) string { return "plugins.external." + configedit.QuoteKey(name) }

// ForgetTrust removes the person's trust decision for a plugin in the
// workspace and reports whether there was one. A sha256 pinned in smith.toml
// is not touched.
func ForgetTrust(root, name string) (bool, error) {
	store, err := plugintrust.OpenStore()
	if err != nil {
		return false, err
	}
	return store.Untrust(root, name)
}
