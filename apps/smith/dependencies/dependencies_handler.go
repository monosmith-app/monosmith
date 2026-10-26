package dependencies

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Commands are `smith install`, `smith uninstall`, `smith sync` and `smith up`.
func Commands(version string) []*cobra.Command {
	var targets []string
	var dev bool
	install := &cobra.Command{
		Use:     "install [packages...]",
		Aliases: []string{"i"},
		Short:   "Add a package to a project (-w), or with none install everything the workspace declares (up updates, sync aligns versions)",
		Long: `Add packages to one project with its own toolchain: npm install --workspace,
the project's pyproject.toml and the workspace venv, go get. Each project owns
its dependencies, so a package needs a target: smith install -w <project>
<package>. With no package, smith install installs everything a fresh clone
needs, for every language in the workspace.`,
		Example: `  smith install                       # a fresh clone: npm install, the venv, go mod download
  smith install -w web zod
  smith install -w apps/api -D pytest-cov
  smith install -w api github.com/pkg/errors@v0.9.1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			return Install(s, targets, args, dev, cmd.OutOrStdout())
		},
	}
	install.Flags().StringArrayVarP(&targets, "workspace", "w", nil, "target project, by name or directory; repeat it for several")
	install.Flags().BoolVarP(&dev, "save-dev", "D", false, "add as a development dependency (npm devDependencies, the Python dev extra, a Go tool)")

	var from []string
	uninstall := &cobra.Command{
		Use:     "uninstall <packages...>",
		Aliases: []string{"un"},
		Short:   "Remove packages from a project (the opposite of install)",
		Long: `Remove packages from one project, the counterpart of smith install. A
dependency belongs to the project that uses it, so it needs a target: smith
uninstall -w <project> <package>.`,
		Example: `  smith uninstall -w web zod`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			return Uninstall(s, from, args, cmd.OutOrStdout())
		},
	}
	uninstall.Flags().StringArrayVarP(&from, "workspace", "w", nil, "target project, by name or directory; repeat it for several")
	return []*cobra.Command{install, uninstall, syncCommand(version), upCommand(version), depsCommand(version)}
}

func upCommand(version string) *cobra.Command {
	return newUpCommand(version)
}

// depsCommand is `smith deps`, the group for workspace-wide dependency work:
// `smith deps up` is `smith up`, spelled where it is found with the other
// dependency commands.
func depsCommand(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deps",
		Short: "Group for dependency work across projects; deps up is the same as up",
		Long:  "Dependency work that spans projects. smith deps up is smith up: it updates every project that declares a package in one change. smith deps up --align brings a dependency that projects declare at different versions to one.",
	}
	cmd.AddCommand(newUpCommand(version))
	return cmd
}

func newUpCommand(version string) *cobra.Command {
	var check, yes, asJSON, noInstall, align bool
	cmd := &cobra.Command{
		Use:   "up [packages...]",
		Short: "Show dependencies with newer releases and update the ones you pick (same as deps up; --align is sync)",
		Long: `Show every declared dependency with a newer release, grouped patch, minor and
major, with the projects declaring it, then update the ones you pick: each
declaration keeps its operator and moves to the latest version, and the
workspace is installed again so lockfiles and the venv follow. Named packages
are updated without a prompt; --yes takes every update. Without a terminal
(or with --check) up only reports. A last section compares the tools smith
pins (Node.js, Python, Go, golangci-lint) with their newest releases.

Every project that declares a package is updated in the same change. With
--align nothing is updated: the packages named (or, with none, every one)
that projects declare at different versions are brought to one, as smith sync
does, leaving the ones smith.toml [dependencies] allow lists.`,
		Example: `  smith up              # pick in the terminal
  smith up zod jest     # exactly these, in every project that uses them
  smith up --yes        # everything
  smith up --align axios
  smith up --json`,
		RunE: func(cmd *cobra.Command, names []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			if align {
				aligned, err := Align(s, names, check, !noInstall)
				if asJSON {
					enc := json.NewEncoder(cmd.OutOrStdout())
					enc.SetIndent("", "  ")
					if encErr := enc.Encode(aligned); encErr != nil {
						return encErr
					}
				} else {
					PrintSync(cmd.OutOrStdout(), aligned, check)
				}
				return err
			}
			report, err := FindAvailable(s)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(report)
			}
			PrintUp(out, report)
			chosen := report.Updates
			switch {
			case len(report.Updates) == 0:
				return nil
			case len(names) > 0:
				var unknown []string
				chosen, unknown = Choose(report.Updates, names)
				if len(unknown) > 0 {
					return fmt.Errorf("no update for %s", strings.Join(unknown, ", "))
				}
			case yes:
			case check || !interactive():
				fmt.Fprintln(out, "\nReport only: pick with smith up in a terminal, name packages, or pass --yes.")
				return nil
			default:
				final, err := tea.NewProgram(newPicker(report.Updates), tea.WithOutput(out)).Run()
				if err != nil {
					return err
				}
				if chosen = final.(picker).Chosen(); len(chosen) == 0 {
					fmt.Fprintln(out, "Nothing updated.")
					return nil
				}
			}
			files, err := Apply(s, chosen, !noInstall)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "✔ %d updated (%s)\n", len(chosen), strings.Join(files, ", "))
			return nil
		},
	}
	cmd.Flags().BoolVar(&align, "align", false, "bring dependencies declared at different versions to one (the named ones, or all); with --check, fail when any differ")
	cmd.Flags().BoolVar(&check, "check", false, "report only; never prompt, never write")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "take every update without prompting")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the updates and tools as JSON; never write")
	cmd.Flags().BoolVar(&noInstall, "no-install", false, "rewrite the manifests without installing")
	return cmd
}

// interactive reports whether up can ask: a terminal on both ends, not CI.
func interactive() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) && os.Getenv("CI") == ""
}

func syncCommand(version string) *cobra.Command {
	var check, asJSON, noInstall bool
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Make every project declare a shared dependency at the same version (same as up --align)",
		Long: `Make every project declare a shared dependency at the same spec, language by
language. The target is the version installed, with the operator most
declarations use; when nothing says what is installed, the highest version
declared. Peer dependencies and the workspace's own packages are left alone,
and so is a spec smith does not understand well enough to rewrite (a URL, a git
ref, a compound range): sync reports those. After rewriting it installs, so
lockfiles and the venv follow.`,
		Example: `  smith sync --check   # in CI: fail when declarations disagree
  smith sync`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			report, err := Sync(s, check, !noInstall)
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				if encErr := enc.Encode(report); encErr != nil {
					return encErr
				}
			} else {
				PrintSync(cmd.OutOrStdout(), report, check)
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "report only; fail when declarations disagree")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the report as JSON")
	cmd.Flags().BoolVar(&noInstall, "no-install", false, "rewrite the manifests without installing")
	return cmd
}
