package convert

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convertchoices"
	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/turboplan"
	"github.com/monosmith-app/monosmith/apps/smith/turboworkspace"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// ErrBlocked is returned when the plan has findings that stop the conversion.
var ErrBlocked = errors.New("the conversion is blocked; see the report")

// Command is `smith convert`.
func Command(version string) *cobra.Command {
	var dryRun, asJSON, strict, skipInstall, allowDirty, verbose, yes, skipVerify bool
	choices := convertchoices.Choices{Level: convertchoices.LevelConservative}
	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert an Nx, Turborepo or mnci workspace to smith (adopt sends you here when it recognises one)",
		Long: `convert turns the Nx or Turborepo workspace in the current directory into a
smith workspace: projects keep their directories, targets (Nx) or turbo.json
tasks (Turborepo) become plugin tasks, package.json scripts or smith.toml
tasks, release settings and existing tags carry over, the old tool is removed
and the CI pipelines are regenerated.

A Turborepo workspace on pnpm or yarn becomes an npm workspace, the one smith
install (smith i) installs: pnpm-workspace.yaml moves into package.json
"workspaces", workspace: and catalog: ranges become what npm accepts, the
pnpm and yarn lockfiles and settings go, and scripts that call turbo, pnpm or
yarn call smith and npm.

Run it with --dry-run first: the report lists what maps, what changes behaviour
and what needs you. Nothing is dropped silently. An Nx workspace's packages must
be installed, because Nx itself describes the projects: in a terminal convert
asks and installs them (about two minutes), --yes installs without asking, and
anywhere else it stops and says so. A Turborepo workspace is read from its
files, so nothing needs to be installed.

--level full goes further: it makes smith the only tool. After the conversion
it works out which files and devDependencies of the old tooling nothing uses
any more (ESLint configs and plugins, the test runner you did not choose,
rollup or webpack, Nx and mnci), and asks about each group before deleting it.
--lint, --test, --build and --ci are the choices behind it; in a terminal
the ones you do not give are asked. Elsewhere the plan is printed and the
command stops, naming the flags that apply it. It needs a clean git tree, so
git undoes everything.`,
		Example: "  smith convert --dry-run\n  smith convert\n  smith convert --level full --dry-run\n  smith convert --level full --lint smith --test jest --build smith --ci smith --yes",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			given := convertchoices.Given{
				Lint: cmd.Flags().Changed("lint"), Test: cmd.Flags().Changed("test"),
				Build: cmd.Flags().Changed("build"), CI: cmd.Flags().Changed("ci"),
			}
			if err := convertchoices.Validate(choices, given); err != nil {
				return problem.New(err.Error())
			}
			full := choices.Level == convertchoices.LevelFull
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			if _, err := os.Stat(filepath.Join(root, workspace.FileName)); err == nil {
				return problem.New("this is already a smith workspace ("+workspace.FileName+" exists)", "smith upgrade brings it up to the installed smith")
			}
			git := gitclient.New(root)
			if !git.IsRepository() {
				return problem.New("this folder is not a git repository: convert makes its changes through git, so each one can be reviewed and undone", `git init && git add -A && git commit -m "before smith", then run this again`)
			}
			if !dryRun && (!allowDirty || full) {
				if status, err := git.Run("status", "--porcelain"); err == nil && status != "" {
					if full {
						return problem.New("the working tree has uncommitted changes: --level full deletes files, and git restore is its way back, so the tree must be clean", "git add -A && git commit -m \"before smith\", then run this again")
					}
					return problem.New("the working tree has uncommitted changes, so the conversion would not be one reviewable change", "git add -A && git commit -m \"before smith\", then run this again", "or pass --allow-dirty to convert anyway")
				}
			}
			files, err := git.Files()
			if err != nil {
				return err
			}
			tags, _ := git.Tags("")
			input := convertplan.Input{
				DirName: filepath.Base(root),
				Files:   files,
				Tags:    tags,
				Read: func(rel string) string {
					raw, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
					return string(raw)
				},
			}
			told := out
			if asJSON {
				told = cmd.ErrOrStderr()
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			if err := EnsureInstalled(ctx, root, InstallChoice{Yes: yes, Terminal: interactive(), In: os.Stdin, Out: told}); err != nil {
				return err
			}
			plan, err := readPlan(root, files, input)
			if err != nil {
				return err
			}
			blocked := plan.Blocked(strict)
			options := ReportOptions{DryRun: dryRun, Strict: strict, Verbose: verbose}
			switch {
			case full && !blocked && !asJSON && !dryRun:
				fmt.Fprintf(out, "Converting %s %q, full level\n\n%s", sourceLabel(plan), plan.Config.Workspace.Name, Verdict(plan, strict))
			case full && !asJSON:
				fmt.Fprint(out, Report(plan, options))
			case full:
				// The JSON holds both plans; runFull prints it.
			case asJSON:
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				if err := enc.Encode(plan); err != nil {
					return err
				}
			case dryRun || blocked:
				fmt.Fprint(out, Report(plan, options))
			default:
				// Applying: the short verdict, not the whole plan. The full
				// report is saved afterwards, and --dry-run prints it.
				fmt.Fprintf(out, "Converting %s %q\n\n%s", sourceLabel(plan), plan.Config.Workspace.Name, Verdict(plan, strict))
			}
			if blocked {
				return ErrBlocked
			}
			if full {
				return runFull(cmd, fullArgs{
					Root: root, Git: git, Plan: plan, Version: version, Files: files, Strict: strict,
					Choices: choices, Given: given, DryRun: dryRun, JSON: asJSON, Yes: yes,
					SkipInstall: skipInstall, SkipVerify: skipVerify,
				})
			}
			if dryRun {
				return nil
			}
			log := out
			if asJSON {
				log = cmd.ErrOrStderr()
			}
			if !yes && interactive() {
				fmt.Fprint(log, "\nThis deletes the old tool's files and rewrites package.json files and the CI pipelines (git can undo it). `smith convert --dry-run` lists every change.\nApply? [y/N] ")
				answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(answer), "y") && !strings.EqualFold(strings.TrimSpace(answer), "yes") {
					return errors.New("convert cancelled: nothing changed")
				}
			}
			fmt.Fprintln(log)
			res, err := Apply(plan, ApplyOptions{Version: version, SkipInstall: skipInstall, Out: log})
			if err != nil {
				return err
			}
			if err := saveReport(root, Report(plan, ReportOptions{Verbose: true, Strict: strict})); err != nil {
				fmt.Fprintf(log, "! could not save the report: %v\n", err)
			}
			staged := stageNewFiles(git, plan, res)
			fmt.Fprint(log, Summary(plan, res.LostSlots, staged))
			return nil
		},
	}
	f := cmd.Flags()
	f.BoolVar(&dryRun, "dry-run", false, "report the plan without changing anything")
	f.BoolVar(&asJSON, "json", false, "print the plan as JSON")
	f.BoolVar(&strict, "strict", false, "refuse to convert while anything needs manual work or is unsupported")
	f.BoolVar(&skipInstall, "skip-install", false, "do not run npm install afterwards")
	f.BoolVar(&verbose, "verbose", false, "name every project a finding applies to, not just a few")
	f.BoolVar(&yes, "yes", false, "apply without asking, and install an Nx workspace's packages first (a terminal asks before changing anything)")
	f.BoolVar(&allowDirty, "allow-dirty", false, "convert even with uncommitted changes (not with --level full)")
	level := (*string)(&choices.Level)
	f.StringVar(level, "level", string(convertchoices.LevelConservative), "how far to go: conservative (keep what works, list the rest) or full (make smith the only tool, deleting the old tooling after asking)")
	f.StringVar(&choices.Lint, "lint", convertchoices.Keep, "full level: lint with smith's ESLint config (smith) or keep the current setup (keep)")
	f.StringVar(&choices.Test, "test", convertchoices.Keep, "full level: the test runner to keep, jest or vitest (the other's configs and packages go), or keep")
	f.StringVar(&choices.Build, "build", convertchoices.Smith, "full level: build with smith's stack build (smith) or keep the current build setup (keep)")
	f.StringVar(&choices.CI, "ci", convertchoices.Smith, "full level: the pipeline smith generates (smith) or keep the current one (keep)")
	f.BoolVar(&skipVerify, "skip-verify", false, "full level: install, but do not build, test, lint and run doctor afterwards")
	return cmd
}

// interactive is true in a terminal that can answer a question; CI and pipes
// apply without asking, as convert always has.
func interactive() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) && os.Getenv("CI") == ""
}

// saveReport writes the full report where the user can read it again, in the
// smith folder that is not committed.
func saveReport(root, text string) error {
	file := filepath.Join(root, filepath.FromSlash(ReportFile))
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	return os.WriteFile(file, []byte(text), 0o644)
}

// stageNewFiles stages the files the conversion created that the builds load
// (and smith.toml), so a clean of untracked files cannot lose them. It returns
// what it staged.
func stageNewFiles(repo *gitclient.Git, plan *convertplan.Plan, res Result) []string {
	var files []string
	for _, f := range res.Created {
		if _, kept := plan.Keep[f]; kept || f == workspace.FileName {
			files = append(files, f)
		}
	}
	if len(files) == 0 {
		return nil
	}
	if _, err := repo.Run(append([]string{"add", "--"}, files...)...); err != nil {
		return nil
	}
	return files
}

// readPlan reads the workspace in root, an Nx or a Turborepo one, and plans
// its conversion.
func readPlan(root string, files []string, input convertplan.Input) (*convertplan.Plan, error) {
	has := func(name string) bool {
		_, err := os.Stat(filepath.Join(root, name))
		return err == nil
	}
	turbo := has("turbo.json") || has("turbo.jsonc")
	switch {
	case has("nx.json") && turbo:
		return nil, errors.New("both nx.json and turbo.json are here: convert one tool at a time, so remove the one you no longer use first")
	case turbo:
		w, err := turboworkspace.Read(root, files)
		if err != nil {
			return nil, err
		}
		return turboplan.MakePlan(w, input), nil
	case !has("nx.json"):
		return nil, errors.New("no nx.json or turbo.json here: convert runs at the root of an Nx or Turborepo workspace")
	}
	w, err := nxworkspace.Read(root, files)
	if err != nil {
		return nil, err
	}
	return convertplan.MakePlan(w, input), nil
}
