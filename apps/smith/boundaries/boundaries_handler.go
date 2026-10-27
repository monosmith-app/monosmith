package boundaries

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// ErrViolations makes `smith boundaries` exit non-zero.
var ErrViolations = errors.New("some projects depend on projects they may not")

// Command is `smith boundaries`.
func Command(version string) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "boundaries",
		Short: "Check which projects depend on which against your rules",
		Long: `Check every dependency between projects against the rules in smith.toml
[boundaries] and the built-in one: no project depends on an app. A dependency
that breaks a rule is printed as the two projects (importer -> imported), the
rule and what to do about it. The command fails when one breaks a rule that is
an error: your own rules, and the built-in one once [boundaries] kinds = true.
smith doctor reports the same findings, and smith ci verify runs this check.

  [boundaries]
  kinds = true
  [[boundaries.rules]]
  from = "tag:ui"
  deny = ["tag:server"]
  reason = "the browser must not pull in server code"
  [[boundaries.exceptions]]
  from = "web"
  to = "admin"
  reason = "the shared login page, until it moves into a lib"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr(), AllowCycles: true})
			if err != nil {
				return err
			}
			defer s.Close()
			violations := Check(s.Graph, s.Config.Boundaries)
			out := cmd.OutOrStdout()
			if asJSON {
				if violations == nil {
					violations = []Violation{}
				}
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				if err := enc.Encode(violations); err != nil {
					return err
				}
			} else {
				fmt.Fprint(out, Text(violations))
			}
			if len(Errors(violations)) > 0 {
				return ErrViolations
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the violations as JSON")
	return cmd
}
