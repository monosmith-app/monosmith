package cachecommand

import (
	"fmt"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/spf13/cobra"
)

func verifyCommand() *cobra.Command {
	var fix, asJSON bool
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Check the local cache for damage; --fix removes what is broken",
		Long: `Check the local cache: every blob must hash to its name, every entry must be
readable and name only blobs that exist, no blob may be without an entry, and
no half-written file may be left behind. It reads every blob, so a large cache
takes a moment. --fix deletes what is broken: corrupt blobs, the entries that
name them, entries that cannot be used, unused blobs and leftover files. A run
simply computes again what a deleted entry held.

Exits non-zero while problems remain.`,
		Example: "  smith cache verify\n  smith cache verify --fix",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			loc, err := locate()
			if err != nil {
				return err
			}
			report, err := cachemaintenance.Verify(loc.Dir, fix, time.Now())
			if err != nil {
				return problem.New("could not read the cache at "+loc.Dir+": "+err.Error(), "close other smith processes and try again", "`smith cache clear` deletes the cache and starts over")
			}
			out := cmd.OutOrStdout()
			if asJSON {
				if err := encode(out, report); err != nil {
					return err
				}
			} else if _, err := fmt.Fprint(out, verifyText(report)); err != nil {
				return err
			}
			if remaining := report.Remaining(); remaining > 0 {
				return problem.New(fmt.Sprintf("the cache at %s has %d %s", loc.Dir, remaining, plural(remaining, "problem", "problems")), "run `smith cache verify --fix` to remove the damaged files; the next run computes what they held again")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&fix, "fix", false, "delete what is broken: corrupt blobs and the entries that name them, unused blobs, leftover files")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the report as JSON")
	return cmd
}
