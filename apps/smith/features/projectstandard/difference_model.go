package projectstandard

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// CopySuffix ends the name of the copy kept of a file the standard replaces,
// as for a pipeline smith adopt replaces.
const CopySuffix = ".pre-smith"

// Difference is one file that is not as smith's standard has it.
type Difference struct {
	// File is the plugin's standard for the file.
	File protocol.StandardFileContent
	// Exists is true when the file is on disk.
	Exists bool
	// Old is the file on disk; New is the whole file with the standard in.
	Old, New []byte
	// Copy is the workspace-relative path the replaced file is kept at, for a
	// whole file that exists; empty otherwise.
	Copy string
}

// Edits are the file writes that apply the difference: the copy of the file
// it replaces first, then the new content.
func (d Difference) Edits() []protocol.FileEdit {
	var out []protocol.FileEdit
	if d.Copy != "" {
		out = append(out, protocol.FileEdit{Path: d.Copy, Content: string(d.Old)})
	}
	return append(out, protocol.FileEdit{Path: d.File.Path, Content: string(d.New)})
}

// Summary says in a few words what applying the difference does to the file.
func (d Difference) Summary() string {
	switch {
	case d.File.Mode == protocol.StandardKeys:
		return d.File.Path + " (" + keyList(d.File.Keys) + ")"
	case d.Copy != "":
		return d.File.Path + " (yours kept as " + d.Copy + ")"
	}
	return d.File.Path + " (new)"
}
