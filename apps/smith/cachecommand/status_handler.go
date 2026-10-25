package cachecommand

import (
	"fmt"
	"math"

	"github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"
	"github.com/spf13/cobra"
)

func statusCommand() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the local cache: where it is, how big, how full, and the largest entries",
		Long: `Show the local cache: its directory, how many entries and blobs it holds, its size
against [cache] max_size_mb, when it was last pruned, whether a remote cache is
configured (scheme and host only) and whether it is in use, and the five tasks
with the largest entries. It reads the store without changing it.`,
		Example: "  smith cache status\n  smith cache status --json",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			loc, err := locate()
			if err != nil {
				return err
			}
			report, err := statusReport(loc)
			if err != nil {
				return err
			}
			if asJSON {
				return encode(cmd.OutOrStdout(), report)
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), statusText(report))
			return err
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the status as JSON")
	return cmd
}

// statusReport measures the cache at loc and describes its remote.
func statusReport(loc location) (StatusReport, error) {
	status, err := cachemaintenance.Inspect(loc.Dir, loc.LimitBytes)
	if err != nil {
		return StatusReport{}, err
	}
	report := StatusReport{Status: status, Remote: summarizeRemote(loc.Remote)}
	if loc.LimitBytes > 0 {
		percent := math.Round(float64(status.Bytes)*1000/float64(loc.LimitBytes)) / 10
		report.UsedPercent = &percent
	}
	return report, nil
}
