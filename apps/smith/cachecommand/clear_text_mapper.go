package cachecommand

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"
	"github.com/monosmith-app/monosmith/apps/smith/marks"
	"github.com/monosmith-app/monosmith/apps/smith/remotesync"
)

// clearQuestion is the question of `smith cache clear --remote` and `--all`:
// the store (host and prefix), the workspace, the count, and that the team
// and CI share the remote. The TUI shows the same text (it runs --dry-run).
func clearQuestion(loc location, remote string, objects remotesync.RemoteObjects, all bool, local cachemaintenance.Status) string {
	store := remotesync.StoreName(remote)
	n := len(objects.Names)
	what := fmt.Sprintf("%d %s of the remote cache %s for the workspace %s", n, plural(n, "object", "objects"), store, loc.Workspace)
	if !objects.Listed {
		what = fmt.Sprintf("up to %d %s of the remote cache %s for the workspace %s", n, plural(n, "object", "objects"), store, loc.Workspace)
	}
	if all {
		if local.Exists {
			what = fmt.Sprintf("the local cache at %s (%d entries, %s) and %s", loc.Dir, local.Entries, formatSize(local.Bytes), what)
		} else {
			what = "the local cache (already empty) and " + what
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Delete %s?\n", what)
	b.WriteString("The team and CI share the remote cache: their next runs build again what it held.")
	if !objects.Listed {
		b.WriteString("\nThis store cannot list its objects: smith deletes only the entries this machine's cache holds, and leaves the rest on the remote.")
	}
	return b.String()
}

// remoteClearedText says what a remote clear deleted.
func remoteClearedText(r remotesync.ClearResult, workspace string) string {
	var b strings.Builder
	mark := marks.Passed
	if r.Failed > 0 {
		mark = marks.Failed
	}
	fmt.Fprintf(&b, "%s Deleted %d of %d %s of the remote cache %s for the workspace %s", mark, r.Deleted, r.Found, plural(r.Found, "object", "objects"), r.Store, workspace)
	if r.Failed > 0 {
		fmt.Fprintf(&b, "; %d failed", r.Failed)
	}
	b.WriteString(".\n")
	for _, e := range r.Errors {
		fmt.Fprintf(&b, "  %s %s\n", marks.Failed, e)
	}
	if !r.Listed {
		b.WriteString("  This store cannot list: objects that other machines stored and this one does not hold are still there.\n")
	}
	return b.String()
}
