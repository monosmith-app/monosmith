package cachecommand

import (
	"fmt"
	"io"
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/remotesync"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
	"github.com/spf13/cobra"
)

func pushCommand() *cobra.Command {
	var dryRun, asJSON bool
	cmd := &cobra.Command{
		Use:   "push",
		Short: "Upload every local cache entry the remote cache does not hold",
		Long: `Upload every local cache entry the remote cache does not hold, with the
files it stores, signed with SMITH_CACHE_KEY as a run signs them. smith asks
the remote first and never uploads an entry it already holds.

A run uploads what it stores at once. Use push to send what a run could not
upload (the remote did not answer, or this machine was offline), or to fill
a new remote from this machine. --dry-run asks the remote and uploads nothing.`,
		Example: "  smith cache push --dry-run\n  smith cache push\n  smith cache push --json",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			loc, err := locate()
			if err != nil {
				return err
			}
			return push(cmd, loc, dryRun, asJSON)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be uploaded; upload nothing")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the result as JSON")
	return cmd
}

// push uploads what the remote of loc lacks and prints the outcome.
func push(cmd *cobra.Command, loc location, dryRun, asJSON bool) error {
	out := cmd.OutOrStdout()
	remote, err := remotesync.Connect(loc.Remote)
	if err != nil {
		return remoteProblem(err, loc.Remote.URL)
	}
	name := remotesync.StoreName(remote.URL)
	if _, err := os.Stat(loc.Dir); err != nil {
		result := remotesync.PushResult{Store: name, DryRun: dryRun, Steps: []remotesync.PushStep{}}
		if asJSON {
			return encode(out, result)
		}
		_, err := fmt.Fprintf(out, "The local cache at %s is empty: there is nothing to upload.\n", loc.Dir)
		return err
	}
	local, err := taskcache.NewLocalStorage(loc.Dir)
	if err != nil {
		return err
	}
	opts := remotesync.PushOptions{DryRun: dryRun}
	if !asJSON {
		// Something must happen at once: the banner, then a line per entry sent.
		_, _ = fmt.Fprint(out, pushBanner(name, dryRun))
		opts.Progress = func(step remotesync.PushStep) { printStep(out, step) }
	}
	result, err := remotesync.Push(cmd.Context(), local, remote, opts)
	result.Store = name
	if err != nil {
		return remoteProblem(err, remote.URL)
	}
	if asJSON {
		return encode(out, result)
	}
	_, err = fmt.Fprint(out, pushClosing(result))
	return err
}

// printStep writes the line of an entry that was sent, would be sent, or
// failed; an entry the remote holds already has no line.
func printStep(out io.Writer, step remotesync.PushStep) {
	if line := pushStepLine(step); line != "" {
		_, _ = fmt.Fprint(out, line)
	}
}
