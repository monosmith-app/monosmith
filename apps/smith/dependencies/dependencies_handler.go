package dependencies

import (
	"bufio"
	"io"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/internallink"
	"github.com/monosmith-app/monosmith/apps/smith/operation"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Commands are `smith install`, `smith uninstall`, `smith sync` and `smith up`.
func Commands(version string) []*cobra.Command {
	var targets []string
	var dev, verbose, registry, yes bool
	var linkRange string
	install := &cobra.Command{
		Use:     "install [packages...]",
		Aliases: []string{"i"},
		Short:   "Add a package to a project (-w), or with none install everything the workspace declares (up updates, sync aligns versions)",
		Long: `Add packages to one project with its own toolchain: npm install --workspace,
the project's pyproject.toml and the workspace venv, go get. Each project owns
its dependencies, so a package needs a target: smith install -w <project>
<package>. With no package, smith install installs everything a fresh clone
needs, for every language in the workspace.

A name that is a library of this workspace (its name, its directory, or the
package name it exports) is linked: the project depends on the library in the
language's own form, and nothing is downloaded. When the registry has a
package of the same name, or cannot be asked, smith asks which you mean; with
no terminal (or --yes) it refuses and names the fix: workspace:<name> for the
project, --registry for the package.`,
		Example: `  smith install                       # a fresh clone: npm install, the venv, go mod download
  smith install -w web zod
  smith install -w apps/api -D pytest-cov
  smith install -w api github.com/pkg/errors@v0.9.1
  smith install -w web core                # a library of this workspace: linked, not downloaded
  smith install -w web workspace:core      # say it is the project
  smith install -w web --registry core     # say it is the registry package`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				// The plugins inherit it: the package manager's own warnings
				// are shown as they come, not folded into one line.
				_ = os.Setenv(toolrun.VerboseEnv, "1")
			}
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			link := LinkOptions{Registry: registry, Range: linkRange}
			if !yes && interactive() {
				link.Ask = internallink.NewTerminalAsker(cmd.InOrStdin(), cmd.OutOrStdout())
			}
			return Install(s, targets, args, dev, cmd.OutOrStdout(), link)
		},
	}
	install.Flags().BoolVar(&registry, "registry", false, "install the names from the registry, even when a name is also a project of this workspace")
	install.Flags().StringVar(&linkRange, "range", "", "the version range written for a library of the workspace (for example '*' or '^1.2.0'); without it smith follows the workspace or asks")
	install.Flags().BoolVarP(&yes, "yes", "y", false, "never ask: accept the defaults, and refuse a name that could be a project or a registry package")
	install.Flags().BoolVar(&verbose, "verbose", false, "show the package manager's own output and warnings (otherwise warnings fold into one line)")
	install.Flags().StringArrayVarP(&targets, "workspace", "w", nil, "target project, by name or directory; repeat it for several")
	install.Flags().BoolVarP(&dev, "save-dev", "D", false, "add as a development dependency (npm devDependencies, the Python dev extra, a Go tool)")

	var from []string
	var fromRegistry bool
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
			return Uninstall(s, from, args, cmd.OutOrStdout(), fromRegistry)
		},
	}
	uninstall.Flags().BoolVar(&fromRegistry, "registry", false, "remove the names as registry packages, even when a name is also a project of this workspace")
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
	var check, yes, asJSON, noInstall, align, dryRun, major bool
	var prefer string
	var to []string
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
				opts, err := syncOptions(cmd, names, check, dryRun, !noInstall, major, yes, prefer, to)
				if err != nil {
					return err
				}
				aligned, err := Align(s, opts)
				if asJSON {
					enc := json.NewEncoder(cmd.OutOrStdout())
					enc.SetIndent("", "  ")
					if encErr := enc.Encode(aligned); encErr != nil {
						return encErr
					}
				} else {
					printSyncResult(cmd.OutOrStdout(), aligned, check, dryRun)
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
			if majors := Majors(chosen); len(majors) > 0 && !major {
				// A major update can break the project, so it is a choice of its own.
				switch {
				case yes || len(names) > 0 && !operationplan.ShouldAsk(operationplan.DetectEnvironment(), false):
					fmt.Fprintf(out, "Left out: %s. A major update can break the project; read its release notes, then add --major to take it.\n", describeMajors(majors))
					chosen = WithoutMajors(chosen)
				case len(names) > 0:
					fmt.Fprintf(out, "\nThese are major updates and can break the project: %s.\nTake them? [y/N] ", describeMajors(majors))
					answer, _ := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
					if !strings.EqualFold(strings.TrimSpace(answer), "y") {
						chosen = WithoutMajors(chosen)
						fmt.Fprintln(out, "Left out the major updates.")
					}
				}
				if len(chosen) == 0 {
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
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "take every patch and minor update without prompting; major updates stay out unless --major is given")
	cmd.Flags().BoolVar(&major, "major", false, "also take major updates (with --yes or named packages), and with --align let a declaration move to another major version")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "with --align: show the plan and change nothing")
	cmd.Flags().StringVar(&prefer, "prefer", "", "with --align: which version wins when projects disagree: highest, lowest or installed (default: the installed one, else the highest)")
	cmd.Flags().StringSliceVar(&to, "to", nil, "with --align: the version of one dependency, as <name>=<version>")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the updates and tools as JSON; never write")
	cmd.Flags().BoolVar(&noInstall, "no-install", false, "rewrite the manifests without installing")
	return operation.Wrap(cmd, operation.Spec{Describe: func(*cobra.Command, []string) string { return "align dependency versions" }})
}

// interactive reports whether up can ask: a terminal on both ends, not CI.
func interactive() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) && os.Getenv("CI") == ""
}

func syncCommand(version string) *cobra.Command {
	var check, asJSON, noInstall, yes, dryRun, major bool
	var prefer string
	var to []string
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Make every project declare a shared dependency at the same version (same as up --align)",
		Long: `Make every project declare a shared dependency at the same spec, language by
language. Declarations on another major version than the target are left alone
unless you pass --major. The target is the version installed, with the operator most
declarations use; when nothing says what is installed, the highest version
declared. Peer dependencies and the workspace's own packages are left alone,
and so is a spec smith does not understand well enough to rewrite (a URL, a git
ref, a compound range): sync reports those. After rewriting it installs, so
lockfiles and the venv follow.

Sync shows a plan first: for each dependency, what every project declares now
and what it will declare. In a terminal it asks which version wins (smith's
pick, the highest, the lowest, or one by one) and then asks to write. Without
a terminal it refuses unless you pass --yes, --prefer or --to, and --dry-run
shows the plan and changes nothing.`,
		Example: `  smith sync --check   # in CI: fail when declarations disagree
  smith sync --dry-run
  smith sync --prefer highest --yes
  smith sync --to axios=^1.6.0 --yes`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			opts, err := syncOptions(cmd, nil, check, dryRun, !noInstall, major, yes, prefer, to)
			if err != nil {
				return err
			}
			report, err := Sync(s, opts)
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				if encErr := enc.Encode(report); encErr != nil {
					return encErr
				}
			} else {
				printSyncResult(cmd.OutOrStdout(), report, check, dryRun)
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "report only; fail when declarations disagree")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the report as JSON")
	cmd.Flags().BoolVar(&noInstall, "no-install", false, "rewrite the manifests without installing")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "write the plan without asking")
	cmd.Flags().BoolVar(&major, "major", false, "also rewrite declarations whose major version differs from the target")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show the plan and change nothing")
	cmd.Flags().StringVar(&prefer, "prefer", "", "which version wins when projects disagree: highest, lowest or installed (default: the installed one, else the highest)")
	cmd.Flags().StringSliceVar(&to, "to", nil, "the version of one dependency, as <name>=<version>")
	return operation.Wrap(cmd, operation.Spec{Describe: func(*cobra.Command, []string) string { return "sync dependency versions" }})
}

// syncOptions builds the options of sync and up --align from the flags: the
// questions a person answers, or the flags that answer them.
func syncOptions(cmd *cobra.Command, names []string, check, dryRun, install, major, yes bool, prefer string, to []string) (SyncOptions, error) {
	if err := ValidatePrefer(prefer); err != nil {
		return SyncOptions{}, err
	}
	targets := map[string]string{}
	for _, v := range to {
		name, spec, ok := strings.Cut(v, "=")
		if !ok || name == "" || spec == "" {
			return SyncOptions{}, fmt.Errorf("--to %q: write it as <name>=<version>", v)
		}
		targets[name] = spec
	}
	r := Reviewer{
		In: bufio.NewReader(cmd.InOrStdin()), Out: cmd.OutOrStdout(),
		Yes: yes, Preview: dryRun, Install: install, CrossMajor: major,
		Asks:   operationplan.ShouldAsk(operationplan.DetectEnvironment(), false),
		Prefer: prefer, To: targets,
	}
	return SyncOptions{Names: names, Check: check, DryRun: dryRun, Install: install, CrossMajor: major, Settle: r.Settle}, nil
}

// printSyncResult prints what a sync or align run did after the plan was shown.
// A check prints the whole report, because no plan was shown.
func printSyncResult(w io.Writer, r SyncReport, check, dryRun bool) {
	switch {
	case check:
		PrintSync(w, r, true)
	case dryRun:
	case len(r.Drift) == 0:
		fmt.Fprintln(w, "✔ Every project agrees on every shared dependency version.")
	default:
		fmt.Fprintf(w, "✔ %d dependencies converged (%s)\n", len(r.Drift), strings.Join(r.Files, ", "))
	}
	if check || dryRun {
		return
	}
	for _, p := range r.Unchecked {
		fmt.Fprintf(w, "⊘ %s: not compared; its plugin does not list declarations\n", p)
	}
}

// describeMajors lists major updates as "name 1.2.0 -> 3.0.0".
func describeMajors(us []Update) string {
	var parts []string
	for _, u := range us {
		parts = append(parts, u.Name+" "+u.Current+" -> "+u.Latest)
	}
	return strings.Join(parts, ", ")
}
