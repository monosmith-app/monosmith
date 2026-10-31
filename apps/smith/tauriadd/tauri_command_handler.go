package tauriadd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/commandargs"
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

func addCommand(version string) *cobra.Command {
	var asJSON bool
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
			res, err := AddTauri(s, args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
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
	return cmd
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
	fmt.Fprintln(out, "Review the changes with git status, set the bundle identifier in tauri.conf.json, then commit.")
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
