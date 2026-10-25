package cachecommand

import (
	"fmt"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/spf13/cobra"
)

func pruneCommand() *cobra.Command {
	var (
		maxSizeMB      int
		dryRun, asJSON bool
	)
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Trim the local cache to its limit now, removing the least recently used entries",
		Long: `Trim the local cache to [cache] max_size_mb (or --max-size-mb) now: remove the
least recently used entries until the rest fit, then the blobs no entry uses.
With a limit of 0 only the unused blobs go. --dry-run prints what would be
removed and the size afterwards, and removes nothing. The remote cache is never
pruned.

smith does this by itself after runs that stored results; run it to trim sooner
or to a smaller size.`,
		Example: "  smith cache prune --dry-run\n  smith cache prune --max-size-mb 512",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			loc, err := locate()
			if err != nil {
				return err
			}
			limit := loc.LimitBytes
			if cmd.Flags().Changed("max-size-mb") {
				if maxSizeMB < 0 {
					return problem.New(fmt.Sprintf("--max-size-mb %d is not a size", maxSizeMB), "pass a number of megabytes, or 0 to remove only unused blobs", "`smith cache clear` deletes the whole cache")
				}
				limit = int64(maxSizeMB) << 20
			}
			outcome, err := cachemaintenance.PruneNow(loc.Dir, limit, dryRun, time.Now())
			if err != nil {
				return problem.New("could not prune the cache at "+loc.Dir+": "+err.Error(), "run `smith cache verify` to check the store", "close other smith processes and try again")
			}
			if asJSON {
				return encode(cmd.OutOrStdout(), outcome)
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), pruneText(outcome))
			return err
		},
	}
	f := cmd.Flags()
	f.IntVar(&maxSizeMB, "max-size-mb", 0, "prune to this size in megabytes instead of [cache] max_size_mb")
	f.BoolVar(&dryRun, "dry-run", false, "show what would be removed and the size afterwards; remove nothing")
	f.BoolVar(&asJSON, "json", false, "print the result as JSON")
	return cmd
}
