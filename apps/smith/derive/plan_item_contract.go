package derive

import (
	"github.com/monosmith-app/monosmith/apps/smith/derivelock"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Status says what derive finds, or will do, for one file.
type Status string

// Statuses of a derived file.
const (
	// StatusCurrent: the file is as the renderer wants it and the record agrees.
	StatusCurrent Status = "current"
	// StatusStale: the renderer's inputs changed and nobody edited the file;
	// derive rewrites it.
	StatusStale Status = "stale"
	// StatusMissing: smith owns the file and it is gone; derive restores it.
	StatusMissing Status = "missing"
	// StatusUnrecorded: the file is as the renderer wants it but the record
	// does not list it; derive records it.
	StatusUnrecorded Status = "unrecorded"
	// StatusDrift: smith owns the file and its content differs from what smith
	// last wrote: a hand edit. derive never overwrites it silently.
	StatusDrift Status = "drift"
	// StatusDiverged: the file exists, the record does not list it and it
	// differs from what the renderer wants. derive leaves it alone.
	StatusDiverged Status = "diverged"
	// StatusBroken: smith cannot edit the file (a region without its markers,
	// a file that does not parse).
	StatusBroken Status = "broken"
	// StatusEjected: a person owns the file; smith does not touch it.
	StatusEjected Status = "ejected"
)

// Item is one derived file: what the renderer wants, what is on disk and what
// derive will do.
type Item struct {
	File protocol.DerivedFile
	// Plugin is the plugin whose renderer produced the file.
	Plugin string
	Status Status
	// Owner is the owner in the record, or smith for a file not yet recorded.
	Owner string
	// Detail says why, in a sentence.
	Detail string
	// Exists tells whether the file is on disk.
	Exists bool
	// Wanted and Present are the content smith owns, as the renderer wants it
	// and as the file has it (nil when it cannot be read).
	Wanted, Present []byte
	// Merged is the new content of the whole file, set for the statuses derive
	// writes (stale, missing) and for drift and diverged, if replaced.
	Merged []byte
	// Entry is the record entry of the file, nil when it is not listed.
	Entry *derivelock.Entry
}

// NeedsAction reports whether derive changes something for the item: it
// writes the file or the record. A drifted, diverged or broken file needs a
// decision, so it counts too.
func (i Item) NeedsAction() bool {
	return i.Status != StatusCurrent && i.Status != StatusEjected
}

// Plan is every derived file of the workspace, sorted by path.
type Plan struct {
	Items []Item
}

// Find returns the item of a path, or nil.
func (p *Plan) Find(path string) *Item {
	for i := range p.Items {
		if p.Items[i].File.Path == path {
			return &p.Items[i]
		}
	}
	return nil
}

// Pending lists the items that need action.
func (p *Plan) Pending() []Item {
	var out []Item
	for _, i := range p.Items {
		if i.NeedsAction() {
			out = append(out, i)
		}
	}
	return out
}
