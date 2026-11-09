package copymeasure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// Command is `smith copies`.
func Command(version string) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "copies",
		Short: "Count the files that projects keep as exact copies of each other, and the lines they cost",
		Long: `List the files at the root of three projects or more that hold the same content
(a tsconfig, a lint config), with the lines they cost and the lines they would
cost if each lived once at the workspace. Line ends and trailing spaces do not
count as a difference. It changes nothing.

A file that most projects hold and a few hold differently is named with the
projects that differ: those are the real exceptions.`,
		Example: "  smith copies\n  smith copies --json",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr(), AllowCycles: true})
			if err != nil {
				return err
			}
			defer s.Close()
			files, err := s.Git.Files()
			if err != nil {
				return err
			}
			var projects []Project
			for _, p := range s.Graph.Projects {
				prefix := p.Root + "/"
				proj := Project{Name: p.Name, Root: p.Root}
				for _, f := range files {
					if strings.HasPrefix(f, prefix) {
						proj.Files = append(proj.Files, f)
					}
				}
				projects = append(projects, proj)
			}
			r := Measure(projects, func(f string) ([]byte, error) { return os.ReadFile(filepath.Join(s.Root, filepath.FromSlash(f))) })
			out := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(r)
			}
			Print(out, r, len(projects))
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the result as JSON")
	return cmd
}

// Print writes the report as a table and a closing line.
func Print(out interface{ Write([]byte) (int, error) }, r Report, projects int) {
	if len(r.Copies) == 0 {
		fmt.Fprintf(out, "Copies: no file is the same in %d projects or more (%d projects looked at).\n", MinCopies, projects)
		return
	}
	fmt.Fprintf(out, "Copies: %d files that %d projects or more keep the same.\n\n", len(r.Copies), MinCopies)
	fmt.Fprintf(out, "%-28s %8s %6s  %s\n", "File", "Projects", "Lines", "Different")
	for _, c := range r.Copies {
		diff := "-"
		if len(c.Different) > 0 {
			diff = strings.Join(c.Different, ", ")
			if len(diff) > 40 {
				diff = fmt.Sprintf("%d projects", len(c.Different))
			}
		}
		fmt.Fprintf(out, "%-28s %8d %6d  %s\n", c.File, len(c.Projects), c.Lines, diff)
	}
	fmt.Fprintf(out, "\nThese copies cost %d lines. Kept once they would cost %d (of %d lines of config at project roots).\n", r.LinesNow, r.LinesOnce, r.ConfigLines)
	fmt.Fprintln(out, "→ A copy that a shared base could replace is a candidate for deletion; smith derive and smith doctor name the ones a plugin can render.")
}
