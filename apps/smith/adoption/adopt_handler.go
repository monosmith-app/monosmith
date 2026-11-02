package adoption

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/adoptdecision"
	"github.com/monosmith-app/monosmith/apps/smith/adoptmove"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/spf13/cobra"
)

// FullLevelIssue tracks the full level for repositories that are not Nx or
// Turborepo workspaces.
const FullLevelIssue = "https://github.com/monosmith-app/monosmith/issues/553"

// The levels --level takes.
const (
	LevelConservative = "conservative"
	LevelFull         = "full"
)

// ErrBlocked is a report with blockers left.
var ErrBlocked = errors.New("something blocks; each blocker above names the step that clears it")

// Command is `smith adopt`.
func Command(version string) *cobra.Command {
	var asJSON, tags, initialise, dryRun, install, eslint, commit, yes bool
	var move string
	var decided adoptdecision.Flags
	level := LevelConservative
	cmd := &cobra.Command{
		Use:   "adopt",
		Short: "Bring an existing repository under smith: report what it involves; --init takes the steps (start here)",
		Long: `Read an existing repository (npm workspaces, Go modules, Python projects) and
report what bringing it under smith involves: the projects smith would see, the
blockers and warnings, and the step that clears each. Read-only unless a step
is asked for, and both steps need a clean git tree:

  --tags  creates, locally, the release tags a project needs where its tags
          were made under another name (nothing is pushed)
  --init  writes smith.toml and the pipelines (with --install it also runs npm
          install; with --eslint it writes the root eslint.config.mjs and adds
          @monosmith/eslint-config to the devDependencies); a pipeline of yours
          at a path smith owns is kept beside it as .pre-smith. It does not
          commit: add --commit to commit the change as one operation that
          smith undo reverts

A repository whose one project is at the root (a package.json, go.mod,
pyproject.toml or Cargo.toml there) needs the project in a folder of its own,
because smith manages one folder per project. --init asks where (apps/<name>,
packages/<name>, libs/<name> or another folder) and never picks for you. Without
a terminal, or with --yes, it moves nothing unless --move <folder> says where
(--move none keeps the project where it is). The move is git mv of the project's
files; the repository's own files (.git, .github, README, LICENSE, ...) stay.
With --commit the move is a commit of its own, before the adoption's. --dry-run
lists the git mv commands for the suggested folder, or for --move.

--level conservative (the default) keeps the repository's build and lint setup.
--level full, which makes smith the only tool, applies to Nx and Turborepo
workspaces and is smith convert --level full's; for a plain repository adopt
says so and stops.

An Nx or Turborepo workspace is converted rather than adopted: adopt hands it to
smith convert (a report stays a --dry-run, --init applies), and --level, --json and
--dry-run carry over, so one command brings in either kind of repository. Exits
non-zero while something blocks.`,
		Example: `  smith adopt --dry-run
  smith adopt --tags --init
  smith adopt --init --move apps/web --commit`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			if move != "" && !initialise && !dryRun {
				return problem.New("--move says where --init puts the project, and --init is not given", "smith adopt --init --move "+move, "or add --dry-run to see the git mv commands")
			}
			if Convertible(".") {
				return handOverToConvert(cmd, level, asJSON, dryRun, initialise, tags, install, eslint, commit)
			}
			switch level {
			case LevelConservative:
			case LevelFull:
				return problem.New("--level full makes smith the only tool, which smith does for Nx and Turborepo workspaces (smith convert) and not yet for a plain repository: adopt keeps its build and lint setup",
					"smith convert --level full --dry-run, in an Nx or Turborepo workspace",
					"smith adopt (the conservative level) for this repository",
					"the full level for plain repositories is tracked in "+FullLevelIssue)
			default:
				return problem.New(fmt.Sprintf("--level %q is not one of conservative, full", level), "smith adopt --level conservative")
			}
			if commit && !initialise {
				return problem.New("--commit commits what --init writes, and --init is not given", "smith adopt --init --commit")
			}
			if dryRun && tags {
				return problem.New("--dry-run reports and changes nothing, so it cannot be combined with --tags", "drop --dry-run to take the step, or drop the step to only look")
			}
			report, err := Inspect(".", version, cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			if (tags || initialise) && !report.Clean && report.Adoptable() {
				return ErrDirty
			}
			if dryRun && move != "" && move != adoptmove.Leave && report.Adoptable() {
				if err := planForDryRun(&report, move); err != nil {
					return err
				}
			}
			if initialise && !dryRun && report.Adoptable() {
				c, plan, err := ChooseMove(report, MoveOptions{Dir: move, Yes: yes, In: cmd.InOrStdin(), Out: out})
				if err != nil {
					return err
				}
				if plan != nil {
					if report, err = MoveRootProject(report, *c, *plan, commit, version, cmd.ErrOrStderr(), out); err != nil {
						return err
					}
				}
			}
			if tags {
				if _, err := CreateBaselines(report, out); err != nil {
					return err
				}
			}
			if initialise && !dryRun {
				decided.CIGiven = cmd.Flags().Changed("ci")
				env := operationplan.DetectEnvironment()
				// One reader for the questions and the confirmation: two would each swallow the other's input.
				in := bufio.NewReader(cmd.InOrStdin())
				choices, err := adoptdecision.Resolve(adoptdecision.Request{
					Flags: decided, Facts: InitFacts(report), Yes: yes,
					Interactive: env.Terminal && !env.CI,
					In:          in, Out: out,
				})
				if err != nil {
					return err
				}
				fmt.Fprintln(out)
				proceed, err := operationplan.Present(out, in, PlanInit(report, choices, install), operationplan.Options{
					Yes: yes, Ask: true, Env: env,
				})
				if err != nil || !proceed {
					return err
				}
				return Init(report, version, install, eslint, commit, choices, out)
			}
			if tags {
				if report, err = Inspect(".", version, cmd.ErrOrStderr()); err != nil {
					return err
				}
			}
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				if err := enc.Encode(report); err != nil {
					return err
				}
			} else {
				printReport(out, report)
				if initialise && dryRun && report.Adoptable() {
					// A dry run settles the open choices with their safe defaults, asks nothing, and shows the plan.
					decided.CIGiven = cmd.Flags().Changed("ci")
					choices, err := adoptdecision.Resolve(adoptdecision.Request{Flags: decided, Facts: InitFacts(report), Preview: true, Out: out})
					if err != nil {
						return err
					}
					fmt.Fprintln(out)
					operationplan.Render(out, PlanInit(report, choices, install))
					fmt.Fprintln(out, "Dry run: nothing was changed.")
				}
				if !tags && !initialise && report.Adoptable() {
					fmt.Fprintln(out, "\nNothing changed (adopt only reports). To apply:\n  → smith adopt --init")
				}
			}
			if len(report.Blockers) > 0 {
				return ErrBlocked
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&level, "level", LevelConservative, "how far to go: conservative (keep the build and lint setup) or full (smith convert's, for Nx and Turborepo workspaces)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the report as JSON")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "only report (this is the default; the flag makes it explicit)")
	cmd.Flags().BoolVar(&tags, "tags", false, "create the baseline release tags the report asks for, locally")
	cmd.Flags().BoolVar(&install, "install", false, "with --init, also run npm install for the Node projects")
	cmd.Flags().BoolVar(&eslint, "eslint", false, "with --init, write the two-line root eslint.config.mjs and add @monosmith/eslint-config to the root devDependencies")
	cmd.Flags().BoolVar(&initialise, "init", false, "write smith.toml and the pipelines (nothing is committed unless --commit)")
	cmd.Flags().StringVar(&move, "move", "", "with --init (or --dry-run), the folder to move the project at the repository root into (apps/web); none keeps it where it is")
	cmd.Flags().BoolVar(&yes, "yes", false, "do not ask: a project at the root moves only when --move says where, and every other open choice of --init takes its safe default and prints it")
	cmd.Flags().StringVar(&decided.Gitattributes, "gitattributes", "", "with --init, add or skip the .gitattributes that sets `* text=auto eol=lf` (a repository that has one keeps it)")
	cmd.Flags().StringVar(&decided.Agent, "agent", "", "with --init, add the MonoSmith agent files (and merge .mcp.json), skip them for now, or never (records [agent] enabled = false)")
	cmd.Flags().StringSliceVar(&decided.CI, "ci", nil, "with --init, the CI providers to write pipelines for (github, azure, gitlab, circleci, bitbucket), or none")
	cmd.Flags().StringVar(&decided.Existing, "ci-existing", "", "with --init, replace a pipeline of yours (kept as <file>.pre-smith) or keep it and write none for that provider")
	cmd.Flags().BoolVar(&commit, "commit", false, "with --init, commit the change as one operation, so smith undo can revert it")
	return cmd
}

func printReport(w io.Writer, r Report) {
	if len(r.Projects) > 0 {
		fmt.Fprintf(w, "%s would be workspace %q (scope %s, CI %s) with:\n", r.Root, r.Name, r.Scope, strings.Join(r.Providers, ", "))
		fmt.Fprintf(w, "  %-20s %-8s %-8s %s\n", "PROJECT", "LANGUAGE", "KIND", "FOLDER")
		for _, p := range r.Projects {
			fmt.Fprintf(w, "  %-20s %-8s %-8s %s\n", p.Name, p.Language, p.Kind, p.Root)
		}
		fmt.Fprintln(w, "  kind: apps/ holds apps, libs/ libs, packages/ packages; elsewhere a Go main package is an app,")
		fmt.Fprintln(w, "  a public Node package is a package and the rest are libs. To change one: \"smith\": {\"kind\": \"app\"}")
		fmt.Fprintln(w, "  in its package.json, or a // smith:kind app comment in its go.mod.")
		fmt.Fprintln(w)
	}
	for _, b := range r.Blockers {
		printItem(w, "✖", b)
	}
	for _, x := range r.Warnings {
		printItem(w, "!", x)
	}
	switch {
	case len(r.Projects) == 0 && r.Adoptable() && r.RootProject != nil:
		printMoveOffer(w, r)
	case len(r.Projects) == 0 && r.Adoptable():
		fmt.Fprintln(w, "! no project to adopt yet: follow a step above, or run smith adopt --init to make an empty workspace and add projects with smith add")
	case len(r.Blockers) == 0 && r.Adoptable():
		fmt.Fprintln(w, "✔ nothing blocks: smith adopt --init makes it a workspace (add --install to run npm install, --eslint to write the lint config), then says what smith doctor still reports")
	case r.Adoptable():
		fmt.Fprintln(w, "Nothing blocks --init; the release blockers above need clearing before the first release.")
	}
	if !r.Clean && r.Adoptable() {
		fmt.Fprintln(w, "(uncommitted changes: commit them before --tags or --init)")
	}
}

// printItem prints a finding and each way out on its own arrow line.
func printItem(w io.Writer, mark string, it Item) {
	fmt.Fprintf(w, "%s %s\n", mark, it.Message)
	steps := it.Steps
	if len(steps) == 0 {
		steps = []string{it.Step}
	}
	for _, step := range steps {
		fmt.Fprintf(w, "  → %s\n", step)
	}
}
