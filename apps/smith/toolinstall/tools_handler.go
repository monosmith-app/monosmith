package toolinstall

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/commandflags"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/libs/words/words"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolcatalog"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// ErrMissing makes smith tools install exit non-zero: a tool that the
// workspace needs is still missing. The output already says which.
var ErrMissing = errors.New("some tools are still missing")

// Interactive is whether a person is at the terminal and this is not CI.
func Interactive() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && os.Getenv("CI") == ""
}

// RunningAsRoot is whether smith runs with administrator rights on a Unix
// system, so sudo is not needed.
func RunningAsRoot() bool { return runtime.GOOS != "windows" && os.Geteuid() == 0 }

// Command is `smith tools`: what the workspace needs on the machine, and
// `smith tools install` to get what is missing.
func Command(version string) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "tools",
		Short: "List the programs the workspace needs and which are installed",
		Long: `A language needs a program on this machine: Python, the .NET SDK, Go, a JDK,
Docker. The language plugins declare what they need. smith detects each program
(it runs its version command, with a 3 second limit) and shows the version
the workspace pins in .nvmrc, go.mod, pyproject.toml, global.json or
rust-toolchain.toml.

  smith tools                    the list and the state of each
  smith tools install            offers the ways to get what is missing
  smith tools install python     only Python
  smith tools use dart           use a tool from the folder winget put it in

On Windows, winget unpacks some packages (the Dart SDK) into a folder under
%LOCALAPPDATA%\Microsoft\WinGet\Packages and does not put it on PATH. smith
finds the program there and offers to use it from that folder, to add the
folder to your user PATH, or to show the folder and change nothing.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			list := DetectAll(cmd.Context(), ExecDetector{}, WorkspaceNeeds(s, nil, nil))
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(list)
			}
			fmt.Fprint(cmd.OutOrStdout(), RenderList(list))
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the tools and their state as JSON")
	cmd.AddCommand(installCommand(version), useCommand())
	return cmd
}

// installCommand is `smith tools install`.
func installCommand(version string) *cobra.Command {
	var manager string
	var yes, dryRun, withAdmin, asJSON bool
	cmd := &cobra.Command{
		Use:   "install [tool...]",
		Short: "Install the programs the workspace needs, asking how for each",
		Long: `For each tool that is missing or too old, smith shows the ways to get it and
asks which you want. Nothing is installed before you choose. The ways are:

  install with a package manager on this machine (winget, brew, apt, dnf,
  pacman, or the language's own uv, fnm, rustup); smith prints the exact command
  first and shows its output
  use the official page (the address)
  I will install it myself (smith shows the commands and goes on)
  skip it (smith says what then stops working)

A command that needs administrator rights (sudo, or winget, which may open the
Windows approval window) is never run in silence: you choose to run it yourself,
to let smith run it (the system asks for its own permission), or to skip it.

--yes takes the marked default of each choice. It never runs a command that needs
administrator rights; --with-admin names that choice for a script. Without a
terminal smith refuses unless a flag answers. --dry-run prints the plan only.
After an install smith detects the tool again and says when a new terminal is
needed for the PATH to update. When winget says the package is already
installed, smith says so and looks for the folder winget put it in.`,
		Example: `  smith tools install               every missing tool, one question each
  smith tools install python        only Python
  smith tools install --dry-run     what smith would offer
  smith tools install --manager winget --with-admin go`,
		RunE: func(cmd *cobra.Command, names []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			needs := WorkspaceNeeds(s, nil, nil)
			if len(names) > 0 {
				var unknown []string
				if needs, unknown = NamedNeeds(s, names); len(unknown) > 0 {
					return problem.New("no plugin declares the tool "+strings.Join(unknown, ", "), "smith tools lists the tools of this workspace")
				}
			}
			out := cmd.OutOrStdout()
			list := DetectAll(cmd.Context(), ExecDetector{}, needs)
			plans := PlansFor(list, !Interactive() && (yes || manager != ""))
			if asJSON && dryRun {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(plans)
			}
			if len(plans) == 0 {
				fmt.Fprintln(out, "Every tool the workspace needs is installed.")
				return nil
			}
			if dryRun {
				fmt.Fprint(out, RenderPlans(plans))
				fmt.Fprintln(out, "\nNothing was installed. Run smith tools install to choose.")
				return nil
			}
			f := Flow{
				In: bufio.NewReader(cmd.InOrStdin()), Out: out, Interactive: Interactive(),
				Yes: yes, WithAdmin: withAdmin, Manager: manager, Runner: ExecRunner{}, Detector: ExecDetector{},
				ToolFolders: ToolFoldersFile(), LocalAppData: os.Getenv("LOCALAPPDATA"),
			}
			outcomes, err := f.Run(cmd.Context(), plans)
			if err != nil {
				return err
			}
			return Verdict(outcomes)
		},
	}
	cmd.Flags().StringVar(&manager, "manager", "", "install every tool with this package manager (winget, brew, apt, dnf, pacman, uv, fnm, rustup, npm), or smith-path or user-path for a tool found off PATH")
	commandflags.AddYesFlag(cmd, &yes, "take the marked default of each choice (never for commands that need administrator rights)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the plan and install nothing")
	cmd.Flags().BoolVar(&withAdmin, "with-admin", false, "let smith run commands that need administrator rights (the system still asks for its own permission)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "with --dry-run, print the plan as JSON")
	return cmd
}

// PlansFor builds the plan of every tool that needs an install, for the
// package managers on this machine. nonInteractive adds the flags that stop a
// manager from asking questions.
func PlansFor(list []Status, nonInteractive bool) []ToolPlan {
	managers := DetectManagers(runtime.GOOS, exec.LookPath)
	env := Env{OS: runtime.GOOS, Root: RunningAsRoot(), NonInteractive: nonInteractive, ToolFolders: ToolFoldersFile()}
	var plans []ToolPlan
	for _, st := range list {
		if st.Lacking() {
			plans = append(plans, BuildPlan(st, managers, env))
		}
	}
	return plans
}

// Verdict turns the outcomes into the exit status: ErrMissing when a tool that
// is not optional is still lacking.
func Verdict(outcomes []Outcome) error {
	for _, o := range outcomes {
		if o.State != Found && !o.Optional {
			return ErrMissing
		}
	}
	return nil
}

// useCommand is `smith tools use <tool>`: the "let smith use it from that
// folder" answer, as a command a fix step and a person can run.
func useCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "use <tool>",
		Short:   words.Say("toolfolder.use.short"),
		Args:    cobra.ExactArgs(1),
		Example: "  smith tools use dart",
		RunE: func(cmd *cobra.Command, args []string) error {
			entry, ok := toolcatalog.Find(args[0])
			if !ok {
				return problem.New(words.Say("toolfolder.use.unknown", "tool", args[0]), words.Say("toolfolder.use.unknown.remedy"))
			}
			dir, err := UseFolder(Need{ToolNeed: entry}, os.Getenv("LOCALAPPDATA"), ToolFoldersFile())
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), words.Say("toolfolder.used", "title", entry.Title, "folder", dir))
			return nil
		},
	}
}
