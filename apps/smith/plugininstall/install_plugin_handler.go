package plugininstall

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// Command is `smith plugin install`. describe reads the plugin.json of a
// directory and resolves the command it would run on this machine: the
// manifest format belongs to the plugin host, not to the installer. resolve,
// when it is not nil, turns a marketplace name (widget, widget@1.2.0) into an
// archive URL and its digest before the install; it leaves every other
// argument alone.
func Command(describe func(dir string) (Described, error), resolve Resolver) *cobra.Command {
	var (
		global, force, trust, pin bool
		sha                       string
	)
	cmd := &cobra.Command{
		Use:   "install <source>",
		Short: "Install a packaged plugin, checked against its sha256",
		Long: `Installs a plugin from the marketplace by name (widget or widget@1.2.0, see
"smith plugin search"), from an archive made by "smith plugin pack", from a plugin
directory, from an https URL, or from a GitHub release (github:owner/repo@tag,
the release's one *.smith-plugin.zip asset, or github:owner/repo@tag#<asset>).
A name that is also a file or directory here is that path.

Before anything is installed, smith checks that
  - the archive matches the SHA-256 you give with --sha256, or the one the
    release publishes in <archive>.sha256, or the marketplace index lists (an
    archive without any is refused; a plugin directory has no archive and
    needs none);
  - the archive's checksums.sha256 lists exactly its files and every file
    matches it;
  - every path stays inside the plugin's directory (no "..", absolute paths,
    links or device files), and sizes stay within the limits;
  - the plugin has a program for this platform.
A failed check installs nothing. The plugin is never started.

Inside a workspace the plugin goes to .smith/plugins/<name>; outside one, or
with --global, to $SMITH_HOME/plugins (~/.smith/plugins). With --pin the SHA-256
of the file it runs is also written to smith.toml, as "smith plugin trust --pin"
does, which is what CI checks; a pin belongs to a workspace, so --pin needs one.

Installing does not trust the plugin: smith asks the first time it is used,
unless you pass --trust. A plugin runs with your privileges and smith does not
sandbox it; read it before you trust it. The digest ties the archive to what
the author published; it does not say who the author is.`,
		Example: "  smith plugin install widget@1.2.0\n  smith plugin install ./widget-0.1.0.smith-plugin.zip\n  smith plugin install https://example.com/widget-0.1.0.smith-plugin.zip --sha256 <digest>\n  smith plugin install github:acme/smith-plugin-widget@v0.1.0 --pin --trust",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source, want := args[0], sha
			if resolve != nil {
				r, ok, err := resolve(cmd.Context(), source)
				if err != nil {
					return err
				}
				if ok {
					if sha != "" {
						return problem.New("--sha256 cannot be used with a marketplace name: the index gives the digest", "drop --sha256, or install from the archive URL")
					}
					source, want = r.Source, r.SHA256
					fmt.Fprintf(cmd.OutOrStdout(), "Found %s in the marketplace %s.\n", r.Name, r.Origin)
				}
			}
			root, _ := workspace.FindRoot(".")
			inWorkspace := root != "" && !global
			if pin && !inWorkspace {
				return problem.New("--pin writes to smith.toml, and there is no workspace here (or --global was given)", "run it inside a workspace, or drop --pin")
			}
			pluginsDir := filepath.Join(plugintrust.Home(), "plugins")
			if inWorkspace {
				pluginsDir = filepath.Join(root, ".smith", "plugins")
			} else if plugintrust.Home() == "" {
				return problem.New("cannot find your home directory to install into", "set SMITH_HOME")
			}
			res, err := Install(cmd.Context(), NewClient(), Request{
				Source:     source,
				PluginsDir: pluginsDir,
				SHA256:     want,
				Replace:    force,
				Inspect:    func(dir string) (Plugin, error) { return inspectStaged(describe, dir) },
			})
			if err != nil {
				return installProblem(err)
			}
			w := cmd.OutOrStdout()
			verb := "Installed"
			if res.Replaced {
				verb = "Replaced"
			}
			fmt.Fprintf(w, "%s %s %s in %s (%d files, from %s).\n", verb, res.Name, res.Version, res.Dir, res.Files, res.Source)
			switch {
			case res.ArchiveVerified:
				fmt.Fprintf(w, "  archive sha256 %s matches the digest you gave, or the marketplace or release published.\n", res.ArchiveSHA256)
			}
			if res.Checksummed {
				fmt.Fprintln(w, "  every file matches checksums.sha256.")
			} else {
				fmt.Fprintln(w, "  the directory has no checksums.sha256; installed as it is.")
			}
			fmt.Fprintln(w, "  the plugin was not started.")

			installed, err := describe(res.Dir)
			if err != nil {
				return err
			}
			fp, err := plugintrust.Measure(installed.Argv)
			if err != nil {
				return problem.New(fmt.Sprintf("installed %s, but cannot read the file it runs: %v", res.Name, err))
			}
			digest := plugintrust.PinnedDigest(fp)
			if pin {
				table := "plugins.external." + configedit.QuoteKey(res.Name)
				if err := configedit.SetTableValue(root, table, "sha256", digest); err != nil {
					return err
				}
				fmt.Fprintf(w, "Pinned sha256 %s in smith.toml [%s], which is what CI checks.\n", digest, table)
			} else {
				fmt.Fprintf(w, "Nothing was pinned. In a workspace, add --pin to write the sha256 of the file it runs (%s) to smith.toml; CI needs it.\n", digest)
			}
			if !inWorkspace {
				return nil
			}
			if trust {
				store, err := plugintrust.OpenStore()
				if err != nil {
					return err
				}
				if err := store.Trust(root, res.Name, fp); err != nil {
					return err
				}
				fmt.Fprintf(w, "Trusted %s in %s.\n", res.Name, root)
			} else {
				fmt.Fprintf(w, "smith will ask before it first starts %s (or run `smith plugin trust %s`).\n", res.Name, res.Name)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&global, "global", false, "install into $SMITH_HOME/plugins even inside a workspace")
	cmd.Flags().BoolVar(&pin, "pin", false, "also pin the sha256 of the file the plugin runs in smith.toml (what CI checks)")
	cmd.Flags().StringVar(&sha, "sha256", "", "the SHA-256 the archive must have; the install is refused if it differs")
	cmd.Flags().BoolVar(&force, "force", false, "replace an installed plugin of the same name")
	cmd.Flags().BoolVar(&trust, "trust", false, "also record that you trust the plugin in this workspace")
	return cmd
}

// inspectStaged reads the plugin.json of freshly extracted files and checks
// that the plugin has a program for this platform. It hashes files; it runs
// nothing.
func inspectStaged(describe func(string) (Described, error), dir string) (Plugin, error) {
	d, err := describe(dir)
	if err != nil {
		return Plugin{}, err
	}
	if _, err := plugintrust.Measure(d.Argv); err != nil {
		return Plugin{}, fmt.Errorf("the plugin has no program for this platform: plugin.json runs %s (%w)", d.Command0, err)
	}
	return d.Plugin, nil
}

// installProblem puts a remedy on the failures people can act on.
func installProblem(err error) error {
	var mismatch *DigestMismatchError
	var noDigest *NoDigestError
	switch {
	case errors.As(err, &mismatch):
		return &problem.Error{
			Problem: err.Error(),
			Remedy:  []string{"nothing was installed: download the archive again, or check where the digest came from"},
			Cause:   err,
		}
	case errors.As(err, &noDigest):
		return &problem.Error{
			Problem: err.Error(),
			Remedy:  []string{noDigest.Hint, "nothing was installed"},
			Cause:   err,
		}
	case errors.Is(err, ErrAlreadyInstalled):
		return &problem.Error{Problem: err.Error(), Remedy: []string{"nothing was changed"}, Cause: err}
	}
	return problem.New(err.Error(), "nothing was installed")
}
