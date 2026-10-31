package upgrade

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/operation"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// Command is `smith upgrade`.
func Command(version string) *cobra.Command {
	var check, dryRun, asJSON bool
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Bring the workspace up to this smith and its recipes: migrate smith.toml, regenerate the files smith owns",
		Long: `After installing a newer smith, run smith upgrade: it migrates smith.toml to the
current schema (comments and layout kept) and regenerates the files smith owns,
deleting one the workspace no longer asks for (a dropped CI provider). Files
smith wrote once (README, .gitignore) and your own are left alone; steps in
smith:slot markers are kept.

A project made from a recipe is brought up to date with the recipe as it is
now: smith resolves it again with the answers the project recorded, updates the
recipe version, the stack and the settings recorded in smith.toml, and has the
project's plugins rewrite the files they own from the new settings. Starter
files, which a plugin wrote once when the project was created, are never
touched; when the recipe adds a plugin whose starter files the project lacks,
upgrade says so.

--dry-run shows the plan, including the change to smith.toml, and changes
nothing. --check does the same and fails when something would change, for CI.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			check = check || dryRun
			yes, _ := cmd.Flags().GetBool("yes")
			if !check && operationplan.ShouldAsk(operationplan.DetectEnvironment(), yes) {
				// A person at a terminal sees the plan, from a dry run, and confirms it.
				preview, err := Upgrade(".", version, true, io.Discard)
				if err != nil {
					return err
				}
				if preview.Changed() {
					root, err := workspace.FindRoot(".")
					if err != nil {
						return err
					}
					if _, err := operationplan.Present(cmd.OutOrStdout(), cmd.InOrStdin(), PlanUpgrade(root, preview), operationplan.Options{
						Ask: true, Env: operationplan.DetectEnvironment(),
					}); err != nil {
						return err
					}
				}
			}
			r, err := Upgrade(".", version, check, cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				if err := enc.Encode(r); err != nil {
					return err
				}
			} else {
				verb, projectVerb := "upgraded", "upgraded"
				if check {
					verb, projectVerb = "would change", "would upgrade"
				}
				for _, m := range r.Migrations {
					fmt.Fprintf(out, "smith.toml %s\n", m)
				}
				for _, c := range r.Projects {
					fmt.Fprintf(out, "%s %s: recipe %s%s → %s\n", projectVerb, c.Project, c.Recipe, versionOf(c.From), orUnversioned(c.To))
					for _, change := range c.Changes {
						fmt.Fprintf(out, "    %s\n", change)
					}
				}
				for _, f := range r.Files {
					fmt.Fprintf(out, "%s: %s\n", verb, f)
				}
				for _, c := range r.Recipes {
					fmt.Fprintf(out, "recipe %s\n", c)
				}
				for _, l := range r.Lost {
					fmt.Fprintf(out, "warning: slot %s has no place in the new pipeline; its steps were not kept\n", l)
				}
				for _, skipped := range r.Skipped {
					fmt.Fprintf(out, "left as it is: %s\n", skipped)
				}
				for _, starter := range r.Starters {
					fmt.Fprintf(out, "starter files: %s\n", starter)
				}
				if r.Diff != "" {
					fmt.Fprintf(out, "\n%s", r.Diff)
				}
				if !r.Changed() {
					fmt.Fprintln(out, "The workspace is up to date.")
				}
				if len(r.Hints) > 0 {
					fmt.Fprintln(out, "\nSuggestions:")
					for _, h := range r.Hints {
						lines := strings.Split(h, "\n")
						fmt.Fprintf(out, "  - %s\n", lines[0])
						for _, l := range lines[1:] {
							fmt.Fprintf(out, "    %s\n", l)
						}
					}
				}
			}
			if check && !dryRun && r.Changed() {
				return errors.New("the workspace is not up to date: run smith upgrade")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "change nothing; fail when something would change")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show the plan, with the change to smith.toml, and change nothing")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the result as JSON")
	return operation.Wrap(cmd, operation.Spec{AddYes: true, Describe: func(*cobra.Command, []string) string { return "upgrade the workspace" }})
}

func versionOf(v string) string {
	if v == "" {
		return ""
	}
	return " " + v
}

func orUnversioned(v string) string {
	if v == "" {
		return "unversioned"
	}
	return v
}
