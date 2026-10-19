package convert

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
	"github.com/monosmith-app/monosmith/apps/smith/turboplan"
	"github.com/monosmith-app/monosmith/apps/smith/turboworkspace"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// ErrBlocked is returned when the plan has findings that stop the conversion.
var ErrBlocked = errors.New("the conversion is blocked; see the report")

// Command is `smith convert`.
func Command(version string) *cobra.Command {
	var dryRun, asJSON, strict, skipInstall, allowDirty bool
	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert an Nx or Turborepo workspace (or one made by the old mnci) to smith",
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
and what needs you. Nothing is dropped silently. An Nx workspace's
dependencies must be installed (npm ci), because Nx itself describes the
projects; a Turborepo workspace is read from its files, so nothing needs to be
installed.`,
		Example: "  smith convert --dry-run\n  smith convert",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			if _, err := os.Stat(filepath.Join(root, workspace.FileName)); err == nil {
				return fmt.Errorf("%s already exists: this workspace is already a smith workspace", workspace.FileName)
			}
			git := gitclient.New(root)
			if !git.IsRepository() {
				return errors.New("convert runs in a git repository, so every change it makes can be reviewed and undone")
			}
			if !dryRun && !allowDirty {
				if status, err := git.Run("status", "--porcelain"); err == nil && status != "" {
					return errors.New("the working tree has changes: commit or stash them first, so the conversion is one reviewable change (or pass --allow-dirty)")
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
			plan, err := readPlan(root, files, input)
			if err != nil {
				return err
			}
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				if err := enc.Encode(plan); err != nil {
					return err
				}
			} else {
				fmt.Fprint(out, Report(plan, dryRun))
			}
			if plan.Blocked(strict) {
				return ErrBlocked
			}
			if dryRun {
				if !asJSON {
					fmt.Fprintln(out, "\nNothing changed (dry run). Run `smith convert` to apply.")
				}
				return nil
			}
			log := out
			if asJSON {
				log = cmd.ErrOrStderr()
			}
			fmt.Fprintln(log)
			res, err := Apply(plan, ApplyOptions{Version: version, SkipInstall: skipInstall, Out: log})
			if err != nil {
				return err
			}
			for _, s := range res.LostSlots {
				fmt.Fprintf(log, "pipeline slot %s holds steps the new pipeline has no place for: move them into another slot\n", s)
			}
			fmt.Fprintln(log, "\nConverted. Next: review `git status`, run `smith doctor` and `smith run lint typecheck test build`, then commit.")
			if plan.Source == convertplan.SourceTurborepo {
				fmt.Fprintln(log, "Install dependencies with `smith i` from now on, not pnpm or yarn.")
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.BoolVar(&dryRun, "dry-run", false, "report the plan without changing anything")
	f.BoolVar(&asJSON, "json", false, "print the plan as JSON")
	f.BoolVar(&strict, "strict", false, "refuse to convert while anything needs manual work or is unsupported")
	f.BoolVar(&skipInstall, "skip-install", false, "do not run npm install afterwards")
	f.BoolVar(&allowDirty, "allow-dirty", false, "convert even with uncommitted changes")
	return cmd
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
