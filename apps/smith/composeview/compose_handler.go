package composeview

import (
	"encoding/json"
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/commandflags"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/compose"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/tui/composer"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Command is `smith compose`.
//
// Parameters: version is smith's version, which the built-in plugins report.
// Returns: the command.
func Command(version string) *cobra.Command {
	var o Options
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "compose",
		Short: "Search everything smith can build a project from, and compose the stack",
		Long: `One place for the language, the framework, the desktop shell (Tauri, Electron),
the cloud host (Azure Functions, AWS Lambda, Google Cloud Functions) and the test,
lint and style tools of a project. In a terminal it is a screen: search, tick what
you want, see each conflict explained as you tick, answer what smith does not
choose for you, and end in the plan: the exact smith add line, the files it
writes and the tools it needs. You run it, copy it or save the plugins as a recipe.

Without a terminal it prints JSON instead: the catalogue with --json, filtered
with --search, and with --with it also checks a selection and prints its plan.
Nothing is written.`,
		Example: `  smith compose
  smith compose --json --search desktop
  smith compose --json --lang node --with electron
  smith compose --json --lang node --with react,vitest`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			interactive := term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) && os.Getenv("CI") == ""
			if !asJSON && !interactive {
				return Refusal()
			}
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			if asJSON {
				rep, err := Run(s, o)
				if err != nil {
					return err
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(rep)
			}
			return screen(cmd, s, o)
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON: the catalogue, and with --with the check and the plan")
	cmd.Flags().StringVar(&o.Search, "search", "", "find items by words: name, title, alias (desktop, serverless, web) or description")
	commandflags.AddLangFlag(cmd, &o.Lang, "", "show only this language's items, and check --with for it")
	cmd.Flags().StringSliceVar(&o.With, "with", nil, "check this selection and print its plan, comma separated (react,vitest)")
	cmd.Flags().StringVar(&o.Kind, "kind", "app", "the kind of project the plan is for: app, lib or package")
	cmd.Flags().StringVar(&o.Name, "name", "", "the project name the plan uses (<name> when not given)")
	return cmd
}

// screen runs the terminal screen and then does what the person chose.
func screen(cmd *cobra.Command, s *session.Session, o Options) error {
	cat := compose.Build(Source(s))
	var result composer.Result
	planner := func(r composer.PlanRequest) (compose.Plan, error) {
		rep, err := Run(s, Options{Lang: r.Language, With: r.Picks, Kind: r.Kind, Name: r.Name})
		if err != nil {
			return compose.Plan{}, err
		}
		if rep.Plan == nil {
			return compose.Plan{}, fmt.Errorf("no plan was made")
		}
		return *rep.Plan, nil
	}
	m := composer.New(cat)
	if o.Search != "" {
		m = m.WithSearch(o.Search)
	}
	if _, err := tea.NewProgram(composer.NewProgram(m, planner, &result), tea.WithAltScreen()).Run(); err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	if result.Copied != "" {
		fmt.Fprintf(out, "The command:\n  %s\n", result.Copied)
	}
	if result.Run == nil {
		return nil
	}
	self, err := os.Executable()
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Running: smith %s\n", strings.Join(result.Run, " "))
	run := exec.Command(self, result.Run...)
	run.Stdin, run.Stdout, run.Stderr = cmd.InOrStdin(), out, cmd.ErrOrStderr()
	return run.Run()
}
