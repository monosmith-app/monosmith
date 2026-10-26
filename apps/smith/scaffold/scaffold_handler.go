package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/commandargs"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/recipechoice"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// NewCommand is `smith new`.
func NewCommand(version string) *cobra.Command {
	var opts NewOptions
	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a monorepo: configuration, git, CI pipelines and a first commit",
		Example: `  smith new acme
  smith new acme --dir . --ci github,azure   # in an existing clone`,
		Args: argsOrAsk(commandargs.Spec{Names: []string{"a name for the workspace"}, Example: "smith new acme", More: "to bring in a repository you already have, run smith adopt --dry-run instead"}, isInteractive),
		RunE: func(cmd *cobra.Command, args []string) error {
			args, err := askMissingArguments(args, newQuestions(), cmd.InOrStdin(), cmd.OutOrStdout())
			if err != nil {
				return err
			}
			opts.Name, opts.Version, opts.Out = args[0], version, cmd.OutOrStdout()
			dir, err := NewWorkspace(opts)
			if err != nil {
				return err
			}
			cwd, _ := filepath.Abs(".")
			rel, err := filepath.Rel(cwd, dir)
			if err != nil {
				rel = dir
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nCreated %s. Next:\n  cd %s\n  smith add app web --lang node\n", opts.Name, rel)
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.Dir, "dir", "", "directory to create the workspace in (default ./<name>)")
	cmd.Flags().StringVar(&opts.Scope, "scope", "", "npm scope for packages (default @<name>)")
	cmd.Flags().StringVar(&opts.ModulePrefix, "module-prefix", "", "Go module prefix (default from the origin remote, else example.com/<name>)")
	cmd.Flags().StringSliceVar(&opts.Providers, "ci", nil, "pipelines to generate: github, azure (default github)")
	cmd.Flags().StringVar(&opts.Recipe, "recipe", "", "the recipe smith add creates projects from: a path relative to smith.toml or an https URL (docs/reference/recipes.md)")
	return cmd
}

// AddCommand is `smith add`.
func AddCommand(version string) *cobra.Command {
	var lang, stack string
	var choices []string
	var verbose bool
	cmd := &cobra.Command{
		Use:   "add <kind> <name>",
		Short: "Add an app, an internal lib or a publishable package",
		Long: `Add a project. The kind says what happens to it:

  app      deployed, never published         → apps/<name>
  lib      imported internally, never published → libs/<name>
  package  published to its registry on release → packages/<name>

A combined kind from ` + "`smith kinds`" + ` names the language too: node-app is
app --lang node.

When smith.toml names a recipe ([workspace] recipe), the recipe creates the
project: it asks its choices one at a time (or take --choice name=answer, as
often as needed, without asking), shows the plugins it picks, then makes the
project as a stack's project is made. --stack bypasses the recipe.

The project lints, type-checks, tests and builds immediately, and the CI
pipelines are regenerated to install its toolchain.

smith add leaves the files it wrote uncommitted (smith new commits what it
wrote): review them with git status, then commit.`,
		Example: `  smith add app web --lang node
  smith add app api --stack node-vitest
  smith add lib core --lang go
  smith add package sdk --lang python
  smith add package billing --choice language=python --choice kind=library   # from the workspace's recipe`,
		Args: argsOrAsk(commandargs.Spec{Names: []string{"a kind (app, lib or package)", "a name"}, Example: "smith add app web --lang node", More: "smith kinds lists the kinds and languages"}, isInteractive),
		RunE: func(cmd *cobra.Command, args []string) error {
			args, err := askMissingArguments(args, addQuestions(), cmd.InOrStdin(), cmd.OutOrStdout())
			if err != nil {
				return err
			}
			if verbose {
				// The plugins inherit it: a package manager's output is shown
				// as it runs, instead of only when it fails.
				_ = os.Setenv(toolrun.VerboseEnv, "1")
			}
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			kind := args[0]
			// "node-app" is "app --lang node": the form `smith kinds` lists.
			if k, l, ok := SplitKind(kind); ok && lang == "" {
				kind, lang = k, l
			}
			answers, err := recipechoice.ParseAnswers(choices)
			if err != nil {
				return err
			}
			interactive := term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) && os.Getenv("CI") == ""
			res, err := AddProject(s, AddOptions{
				Kind: kind, Name: args[1], Lang: lang, Stack: stack, Out: cmd.OutOrStdout(),
				Recipe: &RecipeUse{Answers: answers, In: cmd.InOrStdin(), Interactive: interactive},
			})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Wrote %d file(s):\n  %s\n", len(res.Files), strings.Join(res.Files, "\n  "))
			for _, n := range res.Notes {
				fmt.Fprintln(out, n)
			}
			project := s.Graph.Get(args[1])
			if err := VerifyProject(cmd.Context(), s, project, out); err != nil {
				return err
			}
			if hint := tryHint(project); hint != "" {
				fmt.Fprintln(out, hint)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&lang, "lang", "l", "", "language: node, go, python, rust, dotnet, ruby, php, cpp, dart, jvm, bicep, recipe (default: the stack's, or the workspace's only language)")
	cmd.Flags().StringVarP(&stack, "stack", "s", "", "stack the project is built with, bypassing the workspace's recipe (default: the recipe, else [default_stacks] in smith.toml; smith stacks lists them; none: the language plugin alone)")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "show the package manager's output as it runs (it is shown only if it fails otherwise)")
	cmd.Flags().StringArrayVar(&choices, "choice", nil, "answer a choice of the workspace's recipe, name=answer (repeatable); the rest are asked in a terminal")
	return cmd
}

// tryHint is the command smith add suggests for the new project: its test
// task, else its build (a Bicep project or a plain JavaScript library has no
// test task), else nothing.
func tryHint(p *projectgraph.Project) string {
	if p == nil {
		return ""
	}
	for _, task := range []string{"test", "build"} {
		if _, ok := p.Tasks[task]; ok {
			return "Try: smith " + task + " -p " + p.Name
		}
	}
	return ""
}

// nextSteps is what smith new prints last: a cd only when the workspace is not
// the current directory, then the first project to add (smith kinds lists the
// languages).
func nextSteps(name, rel string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\nCreated %s. Next:\n", name)
	if rel != "." {
		fmt.Fprintf(&b, "  cd %s\n", rel)
	}
	b.WriteString("  smith add app web --lang <language>   # smith kinds lists the languages\n")
	return b.String()
}
