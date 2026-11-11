package cachecommand

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"
	"github.com/monosmith-app/monosmith/apps/smith/commandflags"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/remotesync"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// clearFlags are the answers of `smith cache clear`.
type clearFlags struct {
	remote, all, dryRun, yes, asJSON bool
}

func clearCommand() *cobra.Command {
	var f clearFlags
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Delete the local cache; --remote deletes this workspace's remote cache, --all both",
		Long: `Delete the local cache: every entry, blob and record under [cache] dir. The next
run computes everything again.

--remote deletes this workspace's objects on the remote cache instead (every
object under the remote's prefix), and --all deletes both. The team and CI
share the remote cache: their next runs build again what it held. A plain
https:// store cannot list its objects, so smith deletes there only the entries
this machine's cache holds, and says so.

In a terminal it asks first, naming what goes; without one it refuses unless
--yes is given. --dry-run prints the same question and deletes nothing.`,
		Example: "  smith cache clear\n  smith cache clear --yes\n  smith cache clear --remote --dry-run\n  smith cache clear --all",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if f.remote && f.all {
				return problem.New("--remote and --all cannot go together", "use --remote for the remote cache only, or --all for the local and the remote cache")
			}
			loc, err := locate()
			if err != nil {
				return err
			}
			if !f.remote && !f.all {
				return clearLocal(cmd, loc, f)
			}
			return clearTiers(cmd, loc, f)
		},
	}
	fl := cmd.Flags()
	fl.BoolVar(&f.remote, "remote", false, "delete this workspace's objects on the remote cache, not the local cache")
	fl.BoolVar(&f.all, "all", false, "delete the local cache and this workspace's objects on the remote cache")
	fl.BoolVar(&f.dryRun, "dry-run", false, "print what would be deleted, with the question; delete nothing")
	commandflags.AddYesFlag(cmd, &f.yes, "do not ask for confirmation (needed without a terminal)")
	fl.BoolVar(&f.asJSON, "json", false, "print what was deleted as JSON")
	return cmd
}

// clearLocal is `smith cache clear` without --remote or --all: the local
// cache only; the remote is never reached.
func clearLocal(cmd *cobra.Command, loc location, f clearFlags) error {
	out := cmd.OutOrStdout()
	status, err := cachemaintenance.Inspect(loc.Dir, loc.LimitBytes)
	if err != nil {
		return problem.New("could not read the cache at "+loc.Dir+": "+err.Error(), "run `smith cache clear --yes` to delete it without looking")
	}
	if !status.Exists {
		if f.asJSON {
			return encode(out, cachemaintenance.ClearResult{Dir: loc.Dir})
		}
		_, err := fmt.Fprintf(out, "The cache at %s is already empty.\n", loc.Dir)
		return err
	}
	question := fmt.Sprintf("Delete the local cache at %s (%d entries, %s)? The remote cache is not touched.", loc.Dir, status.Entries, formatSize(status.Bytes))
	if f.dryRun {
		_, err := fmt.Fprintf(out, "%s\nDry run: nothing was deleted.\n", question)
		return err
	}
	if err := ask(out, f.yes, question, "smith cache clear --yes", "delete "+formatSize(status.Bytes)+" at "+loc.Dir); err != nil {
		return err
	}
	result, err := clearLocalNow(loc)
	if err != nil {
		return err
	}
	if f.asJSON {
		return encode(out, result)
	}
	_, err = fmt.Fprintln(out, localClearedLine(result))
	return err
}

// clearLocalNow deletes the local cache.
func clearLocalNow(loc location) (cachemaintenance.ClearResult, error) {
	result, err := cachemaintenance.Clear(loc.Dir)
	if err != nil {
		return result, problem.New("could not delete the cache at "+loc.Dir+": "+err.Error(), "close other smith processes (a build may be reading it) and run `smith cache clear --yes` again")
	}
	return result, nil
}

// localClearedLine says what a local clear deleted.
func localClearedLine(r cachemaintenance.ClearResult) string {
	return fmt.Sprintf("Deleted the local cache at %s: %d %s and %d %s (%s).", r.Dir, r.Entries, plural(r.Entries, "entry", "entries"), r.Blobs, plural(r.Blobs, "blob", "blobs"), formatSize(r.Bytes))
}

// ClearReport is the JSON of `smith cache clear --remote` and `--all`.
type ClearReport struct {
	// Local is what the local clear deleted; nil with --remote.
	Local *cachemaintenance.ClearResult `json:"local,omitempty"`
	// Remote is what the remote clear deleted.
	Remote remotesync.ClearResult `json:"remote"`
	DryRun bool                   `json:"dryRun"`
}

// clearTiers is `smith cache clear --remote` and `--all`.
func clearTiers(cmd *cobra.Command, loc location, f clearFlags) error {
	out := cmd.OutOrStdout()
	remote, err := remotesync.Connect(loc.Remote)
	if err != nil {
		return remoteProblem(err, loc.Remote.URL)
	}
	var local taskcache.Storage
	if _, err := os.Stat(loc.Dir); err == nil {
		if local, err = taskcache.NewLocalStorage(loc.Dir); err != nil {
			return err
		}
	}
	if !f.asJSON {
		// Something must happen at once: a listing takes a moment.
		_, _ = fmt.Fprintf(out, "Reading what %s holds for %s…\n", remotesync.StoreName(remote.URL), loc.Workspace)
	}
	objects, err := remotesync.FindObjects(remote, local)
	if err != nil {
		return clearProblem(err, remote.URL)
	}
	var status cachemaintenance.Status
	if f.all {
		if status, err = cachemaintenance.Inspect(loc.Dir, loc.LimitBytes); err != nil {
			return problem.New("could not read the cache at "+loc.Dir+": "+err.Error(), "run `smith cache clear --all --yes` to delete it without looking")
		}
	}
	question := clearQuestion(loc, remote.URL, objects, f.all, status)
	if f.dryRun {
		if f.asJSON {
			return encode(out, ClearReport{Remote: remotesync.ClearResult{Store: remotesync.StoreName(remote.URL), Listed: objects.Listed, Found: len(objects.Names)}, DryRun: true})
		}
		_, err := fmt.Fprintf(out, "%s\nDry run: nothing was deleted.\n", question)
		return err
	}
	flag := "--remote"
	if f.all {
		flag = "--all"
	}
	first, _, _ := strings.Cut(question, "\n")
	what := "delete " + strings.TrimSuffix(strings.TrimPrefix(first, "Delete "), "?") + ", which the team and CI share"
	if err := ask(out, f.yes, question, "smith cache clear "+flag+" --yes", what); err != nil {
		return err
	}
	report := ClearReport{}
	if f.all && status.Exists {
		result, err := clearLocalNow(loc)
		if err != nil {
			return err
		}
		report.Local = &result
		if !f.asJSON {
			_, _ = fmt.Fprintln(out, localClearedLine(result))
		}
	}
	report.Remote, err = remotesync.ClearRemote(cmd.Context(), remote, objects, nil)
	if err != nil {
		return clearProblem(err, remote.URL)
	}
	if f.asJSON {
		err = encode(out, report)
	} else {
		_, err = fmt.Fprint(out, remoteClearedText(report.Remote, loc.Workspace))
	}
	if err == nil && report.Remote.Failed > 0 {
		return problem.New(fmt.Sprintf("%d %s of %s could not be deleted", report.Remote.Failed, plural(report.Remote.Failed, "object", "objects"), report.Remote.Store), "run the command again: it deletes what is left")
	}
	return err
}

// clearProblem is remoteProblem, with the rights a clear needs when the
// store refuses the listing or a delete.
func clearProblem(err error, remote string) error {
	var p *problem.Error
	if errors.As(remoteProblem(err, remote), &p) && strings.Contains(p.Problem, "refused") {
		p.Remedy = []string{"the credentials need list and delete rights: " + deleteRightsRemedy(remote)}
		return p
	}
	return remoteProblem(err, remote)
}

// ask asks question in a terminal, or refuses without one unless yes.
func ask(out io.Writer, yes bool, question, withYes, what string) error {
	if yes {
		return nil
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return problem.New("refusing to delete the cache without a terminal to ask in", "run `"+withYes+"` to "+what)
	}
	if !confirm(os.Stdin, out, question+" [y/N] ") {
		return errors.New("clear cancelled")
	}
	return nil
}

// confirm asks a yes/no question and reports whether the answer was y.
func confirm(in io.Reader, out io.Writer, question string) bool {
	_, _ = fmt.Fprint(out, question)
	answer, _ := bufio.NewReader(in).ReadString('\n')
	return strings.EqualFold(strings.TrimSpace(answer), "y")
}
