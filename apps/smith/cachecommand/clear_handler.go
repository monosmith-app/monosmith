package cachecommand

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func clearCommand() *cobra.Command {
	var yes, asJSON bool
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Delete the local cache (the remote cache is not touched)",
		Long: `Delete the local cache: every entry, blob and record under [cache] dir. The next
run computes everything again. In a terminal it asks first; without one it
refuses unless --yes is given. The remote cache is never touched.`,
		Example: "  smith cache clear\n  smith cache clear --yes",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			loc, err := locate()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			status, err := cachemaintenance.Inspect(loc.Dir, loc.LimitBytes)
			if err != nil {
				return problem.New("could not read the cache at "+loc.Dir+": "+err.Error(), "run `smith cache clear --yes` to delete it without looking")
			}
			if !status.Exists {
				if asJSON {
					return encode(out, cachemaintenance.ClearResult{Dir: loc.Dir})
				}
				_, err := fmt.Fprintf(out, "The cache at %s is already empty.\n", loc.Dir)
				return err
			}
			if !yes {
				if !term.IsTerminal(int(os.Stdin.Fd())) {
					return problem.New("refusing to delete the cache without a terminal to ask in", "run `smith cache clear --yes` to delete "+formatSize(status.Bytes)+" at "+loc.Dir)
				}
				if !confirm(os.Stdin, out, fmt.Sprintf("Delete the local cache at %s (%d entries, %s)? The remote cache is not touched. [y/N] ", loc.Dir, status.Entries, formatSize(status.Bytes))) {
					return errors.New("clear cancelled")
				}
			}
			result, err := cachemaintenance.Clear(loc.Dir)
			if err != nil {
				return problem.New("could not delete the cache at "+loc.Dir+": "+err.Error(), "close other smith processes (a build may be reading it) and run `smith cache clear --yes` again")
			}
			if asJSON {
				return encode(out, result)
			}
			_, err = fmt.Fprintf(out, "Deleted the local cache at %s: %d %s and %d %s (%s).\n", result.Dir, result.Entries, plural(result.Entries, "entry", "entries"), result.Blobs, plural(result.Blobs, "blob", "blobs"), formatSize(result.Bytes))
			return err
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "do not ask for confirmation")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print what was deleted as JSON")
	return cmd
}

// confirm asks a yes/no question and reports whether the answer was y.
func confirm(in io.Reader, out io.Writer, question string) bool {
	_, _ = fmt.Fprint(out, question)
	answer, _ := bufio.NewReader(in).ReadString('\n')
	return strings.EqualFold(strings.TrimSpace(answer), "y")
}
