package scaffold

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/architecture"
	"github.com/monosmith-app/monosmith/apps/smith/commandargs"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/spf13/cobra"
)

// newIsInteractive says whether smith new may ask questions; a test replaces it.
var newIsInteractive = isInteractive

// newFlags are the variables the flags of smith new write to, besides the
// ones NewOptions holds.
type newFlags struct {
	yes, here                              bool
	ci, defaultStacks                      []string
	language, architecture, branch, remote string
	githubReleases                         bool
	agent, noAgent                         bool
	devcontainer, noDevcontainer           bool
	noCommit, includeSecrets               bool
	firstCommit                            string
	cacheMaxMB                             int
}

// NewCommand is `smith new`.
func NewCommand(version string) *cobra.Command {
	var opts NewOptions
	var f newFlags
	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a monorepo: configuration, git, CI pipelines and a first commit",
		Long: `Creates the workspace: smith.toml, the root files, a git repository, the CI
pipelines and a first commit.

At a terminal it asks a few questions: the CI, the main language, the tests
and lint for Node, the code layout to check, and behind "more options" the
release branch and remote, GitHub Releases and the task cache. Enter takes the
default shown. Every question has a flag, and a question whose flag is given is
not asked; --yes takes the defaults for the rest, and without a terminal
nothing is asked.`,
		Example: `  smith new acme
  smith new --here                                      # in this empty folder, named after it
  smith new acme --yes                                  # no questions
  smith new acme --ci github,azure --default-stack node=node-vitest
  smith new acme --dir . --architecture hexagonal       # in an existing clone`,
		Args: newArgs(argsOrAsk(commandargs.Spec{Names: []string{"a name for the workspace"}, Example: "smith new acme", More: "to bring in a repository you already have, run smith adopt --dry-run instead"}, isInteractive)),
		RunE: func(cmd *cobra.Command, args []string) error {
			in := bufio.NewReader(cmd.InOrStdin())
			out := cmd.OutOrStdout()
			if f.here {
				if opts.Dir != "" {
					return problem.New("--here and --dir both say where to create the workspace", "use one of them")
				}
				if err := checkFolderEmpty("."); err != nil {
					return err
				}
				opts.Dir = "."
				args = nameFromFolder(args)
			}
			args, err := askMissingArguments(args, newQuestions(), in, out)
			if err != nil {
				return err
			}
			answers, err := answersFromFlags(cmd, &opts, f)
			if err != nil {
				return err
			}
			opts.Name, opts.Version, opts.Out = args[0], version, out
			asked := !f.yes && newIsInteractive()
			// What the recipe settles is not asked; a flag still wins over it.
			settled := recipeAnswers(opts)
			for key, value := range settled {
				if _, given := answers[key]; !given {
					answers[key] = value
				}
			}
			if asked {
				fmt.Fprintln(out, newBanner(opts.Name, theme.Choose(os.Getenv, true).Icons))
				if len(settled) > 0 {
					fmt.Fprintln(out, describeSettled(settled))
				}
				askNewQuestions(answers, in, out)
			}
			if opts, err = WithAnswers(opts, answers); err != nil {
				return err
			}
			if asked {
				describeChoices(opts, out)
				fmt.Fprintln(out)
				if opts.SaveRecipe == "" {
					opts.SaveRecipe = askSaveRecipe(opts, in, out)
					fmt.Fprintln(out)
				}
			}
			if !asked {
				describeDefaults(cmd, answers, opts, out)
			}
			opts.ModulePrefix = ChooseModulePrefix(PrefixRequest{
				Given: opts.ModulePrefix, Name: opts.Name, Dir: dirOf(opts), Language: opts.Language,
				Interactive: asked, In: in, Out: out,
			})
			decision, err := ChooseFirstCommit(CommitRequest{
				Dir: dirOf(opts), Mode: f.firstCommit, NoCommit: f.noCommit, IncludeSecrets: f.includeSecrets,
				Yes: f.yes, Interactive: asked, In: in, Out: out,
			})
			if err != nil {
				return err
			}
			opts.FirstCommit, opts.IncludeSecrets = decision.Mode, decision.IncludeSecrets
			dir, err := NewWorkspace(opts)
			if err != nil {
				return err
			}
			cwd, _ := filepath.Abs(".")
			rel, err := filepath.Rel(cwd, dir)
			if err != nil {
				rel = dir
			}
			fmt.Fprint(out, nextSteps(opts.Name, rel, opts.Language))
			if opts.SaveRecipe != "" {
				fmt.Fprint(out, recipeSavedText(dir, opts.SaveRecipe))
			}
			return nil
		},
	}
	flags := cmd.Flags()
	flags.BoolVar(&f.here, "here", false, "create the workspace in the current folder, which must be empty; the name defaults to the folder's")
	flags.StringVar(&opts.Dir, "dir", "", "directory to create the workspace in (default ./<name>)")
	flags.StringVar(&opts.Scope, "scope", "", "npm scope for packages (default @<name>)")
	flags.StringVar(&opts.ModulePrefix, "module-prefix", "", "Go module prefix (default from the origin remote, else example.com/<name>)")
	flags.StringSliceVar(&f.ci, "ci", nil, "pipelines to generate: github, azure, gitlab, circleci, bitbucket (default github)")
	flags.StringVar(&f.language, "language", "", "the main language: node, python, go, jvm, dotnet, rust, or later (default later)")
	flags.StringSliceVar(&f.defaultStacks, "default-stack", nil, "the stack Node projects get, as node=<stack>: node-jest (default), node-vitest or node-ts")
	flags.StringVar(&f.architecture, "architecture", "", "the code layout smith checks: "+strings.Join(architecture.Names(), ", ")+" (default vertical-slices)")
	flags.StringVar(&f.branch, "release-branch", "", "the branch releases run from (default main)")
	flags.StringVar(&f.remote, "release-remote", "", "the git remote that receives release tags (default origin)")
	flags.BoolVar(&f.agent, "agent", true, "add the MonoSmith agent: .claude/agents/smith.md, .mcp.json and .vscode/mcp.json (default)")
	flags.BoolVar(&f.noAgent, "no-agent", false, "leave the MonoSmith agent out; smith agent install adds it later")
	flags.BoolVar(&f.devcontainer, "devcontainer", true, "add .devcontainer/devcontainer.json (default)")
	flags.BoolVar(&f.noDevcontainer, "no-devcontainer", false, "leave the dev container out; sets [ci] devcontainer = false")
	flags.BoolVar(&f.noCommit, "no-commit", false, "write the files and commit nothing")
	flags.StringVar(&f.firstCommit, "first-commit", "", "what the first commit holds in a folder that has files and no git repository: smith (only smith's files), all (everything), none (asked when omitted)")
	flags.BoolVar(&f.includeSecrets, "include-secrets", false, "let --first-commit all commit files whose names look like secrets (.env, keys)")
	flags.BoolVar(&f.githubReleases, "github-releases", false, "create a GitHub Release for every release tag")
	flags.StringVar(&opts.CacheDir, "cache-dir", "", "the task cache folder inside the workspace (default .smith/cache)")
	flags.IntVar(&f.cacheMaxMB, "cache-max-mb", 0, "the largest the task cache may grow, in MB; 0 for no limit (default 2048)")
	flags.BoolVarP(&f.yes, "yes", "y", false, "take the default for every question that has no flag")
	flags.StringVar(&opts.SaveRecipe, "save-recipe", "", "save the answers as recipes/<name>.recipe.toml in the workspace, to reuse on others (needs a main language)")
	flags.StringVar(&opts.RecipeFile, "recipe-file", "", "a recipe file on this machine (one saved by another workspace): copied to recipes/ and used as the recipe; its answers are not asked")
	flags.StringVar(&opts.Recipe, "recipe", "", "the recipe smith add creates projects from: a path relative to smith.toml or an https URL (docs/reference/recipes.md)")
	return cmd
}

// answersFromFlags collects the flags that were given into the form the
// questions use, so a question with a flag is not asked.
func answersFromFlags(cmd *cobra.Command, opts *NewOptions, f newFlags) (map[string]string, error) {
	answers := map[string]string{}
	changed := cmd.Flags().Changed
	if changed("ci") {
		answers[AnswerCI] = strings.Join(f.ci, ",")
	}
	if changed("language") {
		answers[AnswerLanguage] = f.language
	}
	for _, pair := range f.defaultStacks {
		lang, stack, ok := strings.Cut(pair, "=")
		if !ok || lang != "node" {
			return nil, fmt.Errorf("--default-stack %s: write node=<stack>; smith new offers a choice for node only (%s)", pair, optionValues(NodeStacks()))
		}
		answers[AnswerNodeStack] = stack
	}
	if changed("architecture") {
		answers[AnswerArchitecture] = f.architecture
	}
	if changed("release-branch") {
		answers[AnswerReleaseBranch] = f.branch
	}
	if changed("release-remote") {
		answers[AnswerReleaseRemote] = f.remote
	}
	if changed("agent") || changed("no-agent") {
		if changed("agent") && changed("no-agent") && f.agent == f.noAgent {
			return nil, fmt.Errorf("--agent and --no-agent say opposite things; use one")
		}
		answers[AnswerAgent] = map[bool]string{true: "yes", false: "no"}[f.agent && !f.noAgent]
	}
	if changed("devcontainer") || changed("no-devcontainer") {
		if changed("devcontainer") && changed("no-devcontainer") && f.devcontainer == f.noDevcontainer {
			return nil, fmt.Errorf("--devcontainer and --no-devcontainer say opposite things; use one")
		}
		answers[AnswerDevcontainer] = map[bool]string{true: "yes", false: "no"}[f.devcontainer && !f.noDevcontainer]
	}
	if changed("github-releases") {
		answers[AnswerGitHubReleases] = map[bool]string{true: "yes", false: "no"}[f.githubReleases]
	}
	if changed("cache-dir") {
		answers[AnswerCacheDir] = opts.CacheDir
	}
	if changed("cache-max-mb") {
		answers[AnswerCacheMaxMB] = strconv.Itoa(f.cacheMaxMB)
	}
	// A flag that settles an advanced answer closes "more options": the others
	// take their defaults instead of being asked about.
	for _, q := range NewQuestions() {
		if _, given := answers[q.Key]; given && q.Advanced {
			answers[AnswerMore] = "no"
		}
	}
	return answers, nil
}

// dirOf is the folder smith new fills: --dir, or ./<name>.
func dirOf(opts NewOptions) string {
	dir := opts.Dir
	if dir == "" {
		dir = opts.Name
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return dir
	}
	return abs
}

// describeDefaults says, when nothing was asked (--yes or no terminal), what
// smith chose for the questions the person did not answer with a flag, and the
// flag that changes each. Nothing is written in silence.
func describeDefaults(cmd *cobra.Command, answers map[string]string, opts NewOptions, out io.Writer) {
	var lines []string
	if _, given := answers[AnswerAgent]; !given && !opts.NoAgent {
		lines = append(lines, "agent files (.claude/agents/smith.md, .mcp.json, .vscode/mcp.json): written; --no-agent leaves them out")
	}
	if _, given := answers[AnswerDevcontainer]; !given && !opts.NoDevcontainer {
		lines = append(lines, ".devcontainer/devcontainer.json: written; --no-devcontainer leaves it out")
	}
	if _, given := answers[AnswerCI]; !given {
		lines = append(lines, "CI: "+strings.Join(opts.Providers, ", ")+"; --ci github,azure,gitlab,circleci,bitbucket changes it")
	}
	if len(lines) == 0 {
		return
	}
	fmt.Fprintln(out, "Chosen for you:")
	for _, l := range lines {
		fmt.Fprintln(out, "  "+l)
	}
	fmt.Fprintln(out)
}
