package convert

import (
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/fullconversion"
	"github.com/monosmith-app/monosmith/apps/smith/operation"
)

// Progress receives one line of text for each stage of a long read: "Asking Nx
// to describe the projects…", "Planned 53 projects.". A command line prints
// the lines; the TUI shows them in place of a fixed "Reading…".
type Progress func(stage string)

// CheckLoads applies the plan to a scratch copy of the repository and opens the
// result, the way applying does, and says whether the converted workspace
// loads. Nothing in root changes. Projects that import each other in a cycle
// load (doctor reports the cycle), so a cycle is not a failure.
//
// Parameters: root and files are the repository and its tracked files; plan is
// the plan to check; version is the running smith's version.
// Returns: nil when the converted workspace loads, else why it does not.
func CheckLoads(root string, files []string, plan *convertplan.Plan, version string) error {
	view, err := fullconversion.MakeView(root, files, func(dir string) error {
		scratch := *plan
		scratch.Root = dir
		var res Result
		return writeConversion(&scratch, ApplyOptions{Version: version, SkipInstall: true, Out: io.Discard}, newSnapshot(dir), &res)
	})
	if err != nil {
		return err
	}
	view.Close()
	return nil
}

// CommitConversion commits what a conversion changed as one operation that
// smith undo reverts, and tells the person. It is what --commit does; without
// it nothing is committed. A commit git refuses is told, and the changes stay
// in the work tree.
//
// Parameters: root is the workspace; out is where the line is printed.
func CommitConversion(root string, out io.Writer) {
	pending := operation.Always(root)
	if pending == nil {
		fmt.Fprintln(out, "\nNot committed: this folder is not a git work tree.")
		return
	}
	hash, err := pending.Finish("convert to smith")
	switch {
	case err != nil:
		fmt.Fprintf(out, "\nNot committed (%v): the changes are in the work tree; commit them when you are ready.\n", err)
	case hash != "":
		fmt.Fprintf(out, "\nCommitted %.8s: chore: convert to smith (smith undo reverts it).\n", hash)
	}
}
