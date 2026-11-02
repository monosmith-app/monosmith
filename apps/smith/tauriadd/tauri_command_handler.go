package tauriadd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/commandargs"
	"github.com/monosmith-app/monosmith/apps/smith/operation"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// Command is `smith tauri`, with `add` and `icon`.
func Command(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tauri",
		Short: "Add Tauri to a Node app, or make its icons",
	}
	cmd.AddCommand(addCommand(version), iconCommandHandler(version))
	return cmd
}

// askIdentifier asks for the bundle identifier; Enter takes def, and a wrong
// answer is asked again, three times at most, then def is taken.
func askIdentifier(in *bufio.Reader, out io.Writer, def string) string {
	fmt.Fprintf(out, "\nThe bundle identifier names the app to the operating system: it decides where\nthe app keeps its data and what an installer or the app store signs.\nIt is hard to change after you ship. Write it like com.yourcompany.yourapp.\n")
	for range 3 {
		fmt.Fprintf(out, "Bundle identifier [%s]: ", def)
		line, err := in.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			return def
		}
		if verr := ValidIdentifier(line); verr == nil {
			return line
		} else if err != nil {
			return def
		} else {
			fmt.Fprintln(out, verr)
		}
	}
	return def
}

func addCommand(version string) *cobra.Command {
	var asJSON, yes, dryRun bool
	var identifier string
	cmd := &cobra.Command{
		Use:   "add <project>",
		Short: "Make an existing Node app a Tauri desktop app",
		Long: `Make a Node app that already exists (a Vite app, say) a Tauri desktop app, as
if it had been made with the react-tauri stack. smith:

  - writes src-tauri: Cargo.toml (with its own empty [workspace]), tauri.conf.json
    whose frontendDist is the folder smith builds the web app into, build.rs,
    src/main.rs, src/lib.rs, capabilities and generated icons;
  - adds @tauri-apps/cli (to the root devDependencies, where smith keeps dev tools),
    @tauri-apps/api (to the project) and a tauri script (to the project);
  - lists src-tauri under exclude in a root Cargo workspace that would take it in;
  - gives the project a stack with tauri added last to its toolchain, in smith.toml:
    its own eslint, test and build tooling stay. A stack that already has that
    make-up (react becomes react-tauri) is reused, otherwise <stack>-tauri is declared;
  - installs the dependencies and regenerates the CI pipelines (Rust, and on Linux
    the system libraries).

smith shows what it will write first and asks the bundle identifier of the app
(--identifier names it; --yes takes com.<workspace>.<project>). --dry-run shows the plan
and changes nothing.

It never overwrites a file that exists, so it is safe to run again. The project
must be a Node app with an index.html.`,
		Example: "  smith tauri add studio",
		Args: commandargs.Exactly(commandargs.Spec{
			Names:   []string{"the Node app to wrap"},
			Example: "smith tauri add studio",
			More:    "smith projects lists the projects",
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			project, found, st, _, err := Inspect(s, args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if !found.SrcTauriExisted {
				if identifier == "" {
					def := DefaultIdentifier(s.Config.Workspace.Name, args[0])
					switch {
					case yes || dryRun:
						identifier = def
						fmt.Fprintf(out, "Bundle identifier: %s (--identifier <id> changes it; the app stores and signs under this name)\n", def)
					case operationplan.ShouldAsk(operationplan.DetectEnvironment(), false):
						identifier = askIdentifier(bufio.NewReader(cmd.InOrStdin()), out, def)
					default:
						return problem.New("nobody can answer the bundle identifier: there is no terminal and --yes was not given",
							"--identifier com.yourcompany.yourapp names it", "--yes takes "+def, "--dry-run shows the plan")
					}
				}
				if err := ValidIdentifier(identifier); err != nil {
					return err
				}
			}
			plan := PlanFor(args[0], project.Root, st.Name, found.SrcTauriExisted, identifier)
			goOn, err := operationplan.Present(out, cmd.InOrStdin(), plan, operationplan.Options{Yes: yes, DryRun: dryRun, Ask: true, Env: operationplan.DetectEnvironment()})
			if err != nil || !goOn {
				return err
			}
			res, err := AddTauri(s, args[0], identifier)
			if err != nil {
				return err
			}
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(res)
			}
			printAdded(out, res)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print what was done as JSON")
	cmd.Flags().StringVar(&identifier, "identifier", "", "the bundle identifier of the app, like com.yourcompany.yourapp (smith asks when you leave it out)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "write without asking; the bundle identifier is com.<workspace>.<project> unless --identifier says otherwise")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be written and change nothing")
	return operation.Wrap(cmd, operation.Spec{Describe: func(_ *cobra.Command, args []string) string { return "add tauri to " + args[0] }})
}

func printAdded(out io.Writer, r Result) {
	fmt.Fprintf(out, "Added Tauri to %s\n", r.Project)
	switch {
	case r.PreviousStack == r.Stack:
		fmt.Fprintf(out, "  stack   %s (already has tauri)\n", r.Stack)
	case r.DeclaredStack:
		fmt.Fprintf(out, "  stack   %s -> %s (declared in smith.toml; the toolchain is the old one plus tauri)\n", orNone(r.PreviousStack), r.Stack)
	default:
		fmt.Fprintf(out, "  stack   %s -> %s (the stack with that make-up)\n", orNone(r.PreviousStack), r.Stack)
	}
	if r.SrcTauriExisted {
		fmt.Fprintln(out, "  src-tauri was already there: files that exist were left as they are")
	}
	for _, f := range r.Files {
		fmt.Fprintf(out, "  writes  %s\n", f)
	}
	fmt.Fprintln(out, "Next:")
	for _, step := range [][2]string{
		{"smith doctor", "check the machine and the config"},
		{"smith run tauri-dev -p " + r.Project, "run the app in development mode"},
		{"smith run tauri-build -p " + r.Project, "build the installers"},
		{"smith tauri icon logo.png -p " + r.Project, "replace the placeholder icons"},
	} {
		fmt.Fprintf(out, "  %-40s %s\n", step[0], step[1])
	}
	fmt.Fprintln(out, "Review the changes with git status, then commit.")
}

func orNone(stack string) string {
	if stack == "" {
		return "no stack"
	}
	return stack
}

func iconCommandHandler(version string) *cobra.Command {
	var project string
	cmd := &cobra.Command{
		Use:   "icon <logo.png>",
		Short: "Make the app's icons from a square PNG",
		Long: `Make every icon of a Tauri app from one square PNG (1024x1024 is best) with the
project's own Tauri CLI, overwriting src-tauri/icons. It is the Tauri CLI's
icon command, run in the app's folder. The app is the only one with a src-tauri,
or the one named with -p.`,
		Example: "  smith tauri icon logo.png -p studio",
		Args: commandargs.Exactly(commandargs.Spec{
			Names:   []string{"the logo, a square PNG"},
			Example: "smith tauri icon logo.png -p studio",
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			name, err := MakeIcons(s, project, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Made the icons of %s from %s\n", name, args[0])
			return nil
		},
	}
	cmd.Flags().StringVarP(&project, "project", "p", "", "the app, when more than one project has a src-tauri")
	return cmd
}
