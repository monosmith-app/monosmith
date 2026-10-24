package whymiss

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// Command is `smith why-miss`.
func Command(version string) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "why-miss <project>:<task>",
		Short:   "Say why a task would run again instead of coming from the cache (a cache miss; explain shows its settings)",
		Example: "  smith why-miss web:build",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			r, err := ExplainMiss(s, args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(r)
			}
			printReport(out, r)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON")
	return cmd
}

func printReport(out io.Writer, r Report) {
	switch r.Status {
	case Hit:
		fmt.Fprintf(out, "%s would come from the cache: nothing changed since its last cached run.\n", r.Task)
	case NotCached:
		fmt.Fprintf(out, "%s does not use the cache (cache = false); it always runs.\n", r.Task)
	case NoEarlierRun:
		fmt.Fprintf(out, "%s would run: there is no cached run of it to compare with.\n", r.Task)
	case NoRecord:
		fmt.Fprintf(out, "%s would run, but its last cached run (%s) has no record of its inputs; run it once with this smith and ask again next time.\n", r.Task, r.LastRun.Local().Format("2006-01-02 15:04"))
	case Changed:
		fmt.Fprintf(out, "%s would run: %d change(s) since its last cached run (%s)\n", r.Task, len(r.Differences), r.LastRun.Local().Format("2006-01-02 15:04"))
		for _, d := range r.Differences {
			fmt.Fprintf(out, "  %s", d)
			if d.What == "dependency" && d.Change == "changed" {
				fmt.Fprintf(out, "   (smith why-miss %s)", d.Name)
			}
			fmt.Fprintln(out)
		}
	}
}
