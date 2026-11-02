package choiceview

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// row is one choice as smith choices reports it.
type row struct {
	choiceselect.Offer
	// Problem says why the choice cannot be used; empty when it can.
	Problem string `json:"problem,omitempty"`
}

// Command is `smith choices`.
func Command(version string) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "choices",
		Short: "List the choices a project can be built from, by axis (framework, test, style)",
		Long: `A choice is one option on an axis of a project: a framework, a test runner, a
style library. Plugins publish them. smith add builds a project from a selection:

  smith add app web --with react,vitest,bootstrap

A selection has one choice per axis, unless the axis allows several. A choice may
require another (chakra requires react) or conflict with one. An axis that a choice
needs and the selection leaves out takes its default, marked (default).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			offers := s.Plugins.Offers()
			problems := choiceselect.Problems(offers, func(language string) bool { _, err := s.Plugins.ForLanguage(language); return err == nil })
			rows := make([]row, len(offers))
			for i, o := range offers {
				rows[i] = row{Offer: o, Problem: problems[choiceselect.ProblemKey(o)]}
			}
			sort.SliceStable(rows, func(i, j int) bool {
				if rows[i].Axis != rows[j].Axis {
					return rows[i].Axis < rows[j].Axis
				}
				return rows[i].Name < rows[j].Name
			})
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(rows)
			}
			printRows(cmd.OutOrStdout(), rows)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON")
	return cmd
}

// printRows writes the rows grouped by axis.
func printRows(out io.Writer, rows []row) {
	if len(rows) == 0 {
		fmt.Fprintln(out, "No choices: no enabled plugin publishes one.")
		return
	}
	axis := ""
	for _, r := range rows {
		if r.Axis != axis {
			if axis != "" {
				fmt.Fprintln(out)
			}
			axis = r.Axis
			fmt.Fprintf(out, "%s\n", axis)
		}
		mark := ""
		if r.Default {
			mark = "  (default)"
		}
		fmt.Fprintf(out, "  %-12s %s  [%s]%s\n", r.Name, r.Description, r.Plugin, mark)
		if len(r.Requires) > 0 {
			fmt.Fprintf(out, "  %-12s requires %s\n", "", strings.Join(r.Requires, ", "))
		}
		if len(r.Conflicts) > 0 {
			fmt.Fprintf(out, "  %-12s conflicts with %s\n", "", strings.Join(r.Conflicts, ", "))
		}
		if r.Problem != "" {
			fmt.Fprintf(out, "  %-12s ✖ cannot be used: %s\n", "", r.Problem)
		}
	}
}
