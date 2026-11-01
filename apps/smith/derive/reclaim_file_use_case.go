package derive

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/derivefile"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// ReclaimChange is what taking one file back from its owner does to the file
// on disk.
type ReclaimChange struct {
	// Path is the workspace-relative path of the file.
	Path string
	// Before is the file on disk (nil when it is gone) and After the file once
	// every renderer of the file has put its content in.
	Before, After []byte
	// Owner is who owns the file in the record: "you" for an ejected file,
	// "smith" for a drifted or diverged one.
	Owner string
	// Exists tells whether the file is on disk.
	Exists bool
}

// Changed reports whether the file's content changes.
func (c ReclaimChange) Changed() bool { return !c.Exists || !bytes.Equal(c.Before, c.After) }

// PlanReclaim works out what giving the named files back to smith does: for
// each path, the file as every renderer of it would leave it. It reads the
// files and writes nothing. It returns an error when smith derives none of a
// path's content, or a renderer cannot edit the file (a region without its
// markers, a file that does not parse).
//
// Parameters: root is the workspace root; plan is the derive plan; paths are
// workspace-relative paths.
func PlanReclaim(root string, plan Plan, paths []string) (map[string]ReclaimChange, error) {
	out := map[string]ReclaimChange{}
	for _, p := range paths {
		items := itemsAt(plan, p)
		if len(items) == 0 {
			return nil, problem.New("smith derives no content for "+p+" now, so there is nothing to give back", "smith derive --check lists the files it derives")
		}
		disk, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
		exists := err == nil
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		change := ReclaimChange{Path: p, Before: disk, After: disk, Exists: exists, Owner: items[0].Owner}
		for _, item := range items {
			if item.Status == StatusBroken {
				return nil, problem.New(p+" cannot be given back: "+item.Detail)
			}
			merged, err := derivefile.Merge(item.File, change.After, exists)
			if err != nil {
				return nil, problem.New(fmt.Sprintf("%s cannot be given back: %v", p, err))
			}
			change.After, exists = merged, true
		}
		out[p] = change
	}
	return out, nil
}

// PrintReclaimDiffs shows, for each file, the diff of what smith overwrites.
func PrintReclaimDiffs(w io.Writer, paths []string, changes map[string]ReclaimChange) {
	for _, p := range paths {
		c := changes[p]
		switch {
		case !c.Exists:
			fmt.Fprintf(w, "\n%s (owner %s) is gone; smith writes it:\n%s", p, c.Owner, doctorfix.UnifiedDiff(p, "", string(c.After)))
		case !c.Changed():
			fmt.Fprintf(w, "\n%s (owner %s) already holds what smith derives; only the record changes.\n", p, c.Owner)
		default:
			fmt.Fprintf(w, "\n%s (owner %s) will be overwritten:\n%s", p, c.Owner, doctorfix.UnifiedDiff(p, string(c.Before), string(c.After)))
		}
	}
}

// ReclaimPlan is the plan smith presents before it takes files back: each file
// it rewrites or creates, and the record.
func ReclaimPlan(paths []string, changes map[string]ReclaimChange) operationplan.Plan {
	out := operationplan.Plan{Title: "give the files back to smith"}
	for _, p := range paths {
		c := changes[p]
		line := operationplan.Line{Action: operationplan.Change, Path: p, Owner: c.Owner, Note: "smith owns it again"}
		switch {
		case !c.Exists:
			line.Action = operationplan.Create
		case !c.Changed():
			line.Note = "content kept; smith owns it again"
		}
		out.Add("Files", line)
	}
	if !out.Empty() {
		out.Add("Record", operationplan.Line{Action: operationplan.Change, Path: "smith.lock"})
	}
	return out
}

// reclaimed reports whether Apply already wrote path for a reclaim.
func reclaimed(out Outcome, path string) bool { return slices.Contains(out.Reclaimed, path) }
