package adoption

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/spf13/cobra"
)

// ErrBlocked is a report with blockers left.
var ErrBlocked = errors.New("something blocks; each blocker above names the step that clears it")

// Command is `smith adopt`.
func Command(version string) *cobra.Command {
	var asJSON, tags, initialise, dryRun, install, eslint bool
	cmd := &cobra.Command{
		Use:   "adopt",
		Short: "Bring an existing repository under smith: report what it involves, then take the steps (start here)",
		Long: `Read an existing repository (npm workspaces, Go modules, Python projects) and
report what bringing it under smith involves: the projects smith would see, the
blockers and warnings, and the step that clears each. Read-only unless a step
is asked for, and both steps need a clean git tree:

  --tags  creates, locally, the release tags a project needs where its tags
          were made under another name (nothing is pushed)
  --init  writes smith.toml and the pipelines and commits (with --install it also
          runs npm install; with --eslint it writes the root eslint.config.mjs and
          adds @monosmith/eslint-config to the devDependencies); a pipeline of yours
          at a path smith owns is kept beside it as .pre-smith

An Nx or Turborepo workspace is smith convert's: adopt says so and stops. Exits
non-zero while something blocks.`,
		Example: `  smith adopt --dry-run
  smith adopt --tags --init`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			if dryRun && (tags || initialise) {
				return problem.New("--dry-run reports and changes nothing, so it cannot be combined with --tags or --init", "drop --dry-run to take the steps, or drop the steps to only look")
			}
			report, err := Inspect(".", version, cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			if (tags || initialise) && !report.Clean && report.Adoptable() {
				return ErrDirty
			}
			if tags {
				if _, err := CreateBaselines(report, out); err != nil {
					return err
				}
			}
			if initialise {
				return Init(report, version, install, eslint, out)
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
			}
			if len(report.Blockers) > 0 {
				return ErrBlocked
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the report as JSON")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "only report (this is the default; the flag makes it explicit)")
	cmd.Flags().BoolVar(&tags, "tags", false, "create the baseline release tags the report asks for, locally")
	cmd.Flags().BoolVar(&install, "install", false, "with --init, also run npm install for the Node projects")
	cmd.Flags().BoolVar(&eslint, "eslint", false, "with --init, write the two-line root eslint.config.mjs and add @monosmith/eslint-config to the root devDependencies")
	cmd.Flags().BoolVar(&initialise, "init", false, "write smith.toml and the pipelines, and commit")
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
