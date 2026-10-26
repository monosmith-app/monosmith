package adoptflow

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// The largest file whose diff the review shows: the diff is worked out line
// by line, so a big file would make the screen wait and fill it with text
// nobody reads. A created file is cheap to diff, an edited one is not.
const (
	maxCreatedDiffBytes = 100 * 1024
	maxEditedDiffBytes  = 40 * 1024
)

// ShowChanges writes the files the conversion changes the way the Doctor fix
// dialog shows a fix: its description, then the diff of each file against the
// repository. It needs Prepare to have made the scratch copy, and c to be what
// Changes returned.
func (s *Session) ShowChanges(c Changes) (string, error) {
	if s.view == nil {
		return "", ErrNotConvertible
	}
	fix, left := changesAsFix(s.view.Dir, c)
	var b strings.Builder
	if err := doctorfix.ShowFix(&b, s.Root, fix); err != nil {
		return "", err
	}
	for _, l := range left {
		fmt.Fprintf(&b, "  %s\n", l)
	}
	return b.String(), nil
}

// changesAsFix is the conversion's file changes as a fix whose files are
// shown, not applied: created and edited files carry the scratch copy's
// content, removed ones are deletions. Files too big or not text are left out
// and named in the lines returned.
func changesAsFix(scratch string, c Changes) (doctorfix.Planned, []string) {
	var edits []protocol.FileEdit
	var left []string
	add := func(rel string, limit int) {
		data, err := os.ReadFile(filepath.Join(scratch, filepath.FromSlash(rel)))
		switch {
		case err != nil:
			left = append(left, rel+": could not be read for the preview")
		case bytes.IndexByte(data, 0) >= 0:
			left = append(left, rel+": not text, so no preview")
		case len(data) > limit:
			left = append(left, fmt.Sprintf("%s: %d KB, too big to preview here", rel, len(data)/1024))
		default:
			edits = append(edits, protocol.FileEdit{Path: rel, Content: string(data)})
		}
	}
	for _, rel := range c.Created {
		add(rel, maxCreatedDiffBytes)
	}
	for _, rel := range c.Edited {
		add(rel, maxEditedDiffBytes)
	}
	for _, rel := range c.Removed {
		edits = append(edits, protocol.FileEdit{Path: rel, Delete: true})
	}
	desc := fmt.Sprintf("The files smith writes: %d created, %d edited, %d removed", len(c.Created), len(c.Edited), len(c.Removed))
	return doctorfix.Planned{Fix: protocol.Fix{Description: desc, Files: edits}}, left
}
