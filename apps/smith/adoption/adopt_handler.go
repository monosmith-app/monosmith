package adoption

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

// ErrBlocked is a report with blockers left.
var ErrBlocked = errors.New("something blocks; each blocker above names the step that clears it")

// Command is `smith adopt`.
func Command(version string) *cobra.Command {
	var asJSON, tags, initialise bool
	cmd := &cobra.Command{
		Use:   "adopt",
		Short: "Bring an existing repository under smith: report what it involves, then take the steps (start here)",
		Long: `Read an existing repository (npm workspaces, Go modules, Python projects) and
report what bringing it under smith involves: the projects smith would see, the
blockers and warnings, and the step that clears each. Read-only unless a step
is asked for, and both steps need a clean git tree:

  --tags  creates, locally, the release tags a project needs where its tags
          were made under another name (nothing is pushed)
  --init  writes smith.toml and the pipelines and commits; a pipeline of yours
          at a path smith owns is kept beside it as .pre-smith

An Nx or Turborepo workspace is smith convert's: adopt says so and stops. Exits
non-zero while something blocks.`,
		Example: `  smith adopt
  smith adopt --tags --init`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
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
				if err := Init(report, version, out); err != nil {
					return err
				}
				fmt.Fprintln(out, "✔ adopted: smith doctor checks the workspace, smith run build builds it")
				return nil
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
	cmd.Flags().BoolVar(&tags, "tags", false, "create the baseline release tags the report asks for, locally")
	cmd.Flags().BoolVar(&initialise, "init", false, "write smith.toml and the pipelines, and commit")
	return cmd
}

func printReport(w io.Writer, r Report) {
	if len(r.Projects) > 0 {
		fmt.Fprintf(w, "%s would be workspace %q (scope %s, CI %s) with:\n", r.Root, r.Name, r.Scope, strings.Join(r.Providers, ", "))
		for _, p := range r.Projects {
			fmt.Fprintf(w, "  %-20s %-8s %-8s %s\n", p.Name, p.Language, p.Kind, p.Root)
		}
		fmt.Fprintln(w)
	}
	for _, b := range r.Blockers {
		fmt.Fprintf(w, "✖ %s\n  → %s\n", b.Message, b.Step)
	}
	for _, x := range r.Warnings {
		fmt.Fprintf(w, "! %s\n  → %s\n", x.Message, x.Step)
	}
	switch {
	case len(r.Blockers) == 0 && r.Adoptable():
		fmt.Fprintln(w, "✔ nothing blocks: smith adopt --init makes it a workspace")
	case r.Adoptable():
		fmt.Fprintln(w, "Nothing blocks --init; the release blockers above need clearing before the first release.")
	}
	if !r.Clean && r.Adoptable() {
		fmt.Fprintln(w, "(uncommitted changes: commit them before --tags or --init)")
	}
}
