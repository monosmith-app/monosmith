package scaffold

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/commandargs"
	"github.com/monosmith-app/monosmith/apps/smith/operation"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/recipechoice"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// AddCommand is `smith add`.
func AddCommand(version string) *cobra.Command {
	var lang, stack, like string
	var choices, with []string
	var verbose, fresh bool
	cmd := &cobra.Command{
		Use:   "add <kind> <name>",
		Short: "Add a project: an app, an internal lib or a publishable package (to add a dependency, use smith install)",
		Long: `Add a project. The kind says what happens to it:

  app      something you run or deploy (a website, an API) → apps/<name>
  lib      code the other projects here import           → libs/<name>
  package  published to a registry on release            → packages/<name>

A combined kind from ` + "`smith kinds`" + ` names the language too: node-app is
app --lang node.

When smith.toml names a recipe ([workspace] recipe), the recipe creates the
project: it asks its choices one at a time (or take --choice name=answer, as
often as needed, without asking), shows the plugins it picks, then makes the
project as a stack's project is made. --stack bypasses the recipe.

--with builds the project from choices instead of a ready-made stack: a
framework, a test runner, a style (` + "`smith choices`" + ` lists them). smith checks
that they fit (one per axis, what each requires, what conflicts), fills an axis they
need with its default, and saves the result in smith.toml as a stack. A
published stack is a saved selection too.

The project follows the siblings: when at least three projects of the same
kind and language, and at least two thirds of them, share a build tool, a
test runner or where their entry points lead, the new project does the same
and says why. If they split, a terminal is asked; elsewhere the default is
used and a warning says so. --fresh gives the stack's default instead, and
--like <project> copies that one project's shape instead of the majority's.

The project lints, type-checks, tests and builds immediately, and the CI
pipelines are regenerated to install its toolchain.

smith add leaves the files it wrote uncommitted: review them with git status,
then commit. With --commit (or commit_operations = true in smith.toml) smith
commits them as one operation that smith undo can revert.`,
		Example: `  smith add app web --lang node
  smith add app api --stack node-vitest
  smith add app web --with react,vitest,bootstrap
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
			began := time.Now()
			addOpts := AddOptions{
				Kind: kind, Name: args[1], Lang: lang, Stack: stack, Out: cmd.OutOrStdout(), Fresh: fresh, With: with, Like: like,
				Recipe: &RecipeUse{Answers: answers, In: cmd.InOrStdin(), Interactive: interactive},
			}
			if err := CheckAdd(s, addOpts); err != nil {
				return err
			}
			yes, _ := cmd.Flags().GetBool("yes")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			proceed, err := operationplan.Present(cmd.OutOrStdout(), cmd.InOrStdin(), PlanAdd(s, addOpts), operationplan.Options{
				Yes: yes, DryRun: dryRun, Ask: true, Env: operationplan.DetectEnvironment(),
			})
			if err != nil || !proceed {
				return err
			}
			res, err := AddProject(s, addOpts)
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
			if err := CheckNewProject(s, project, out); err != nil {
				return err
			}
			fmt.Fprintln(out, doneStep(kind, args[1], len(res.Files), time.Since(began).Round(100*time.Millisecond).String()))
			if hint := tryHint(project); hint != "" {
				fmt.Fprintln(out, hint)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&lang, "lang", "l", "", "language: node, go, python, rust, dotnet, ruby, php, cpp, dart, jvm, bicep, lua, applescript, vba, swift, objc, bundle, recipe (default: the stack's, or the workspace's only language)")
	cmd.Flags().StringVarP(&stack, "stack", "s", "", "stack the project is built with, bypassing the workspace's recipe (default: the recipe, else [default_stacks] in smith.toml; smith stacks lists them; none: the language plugin alone)")
	cmd.Flags().StringSliceVar(&with, "with", nil, "the choices the project is built from, comma separated: smith choices lists them (react,vitest,bootstrap); an axis left out takes its default")
	cmd.Flags().BoolVar(&fresh, "fresh", false, "give the project the built-in default of its stack, without following the sibling projects")
	cmd.Flags().StringVar(&like, "like", "", "copy the build, test, layout and lint of this project instead of what most siblings share; the plan and the output show what was taken")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "show the package manager's output as it runs (it is shown only if it fails otherwise)")
	cmd.Flags().StringArrayVar(&choices, "choice", nil, "answer a choice of the workspace's recipe, name=answer (repeatable); the rest are asked in a terminal")
	return operation.Wrap(cmd, operation.Spec{AddYes: true, AddDryRun: true, Describe: func(_ *cobra.Command, args []string) string {
		return "add " + strings.Join(args, " ")
	}})
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
func nextSteps(name, rel, language string) string {
	app := firstAppName(language)
	if language == "" {
		language = "<language>   # smith kinds lists the languages"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\nCreated %s. Next:\n", name)
	if rel != "." {
		fmt.Fprintf(&b, "  cd %s\n", rel)
	}
	fmt.Fprintf(&b, "  smith add app %s --lang %s\n", app, language)
	return b.String()
}

// firstAppName is the name smith new suggests for the first app: "web" for
// the languages that build websites, "api" for the rest, so a .NET or Go
// workspace is not told to add a website.
func firstAppName(language string) string {
	switch language {
	case "", "node", "dart":
		return "web"
	}
	return "api"
}
