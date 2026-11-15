package derivelock

import "slices"

// FileName is the ownership record, next to smith.toml.
const FileName = "smith.lock"

// SchemaVersion is the smith.lock schema this build reads and writes.
const SchemaVersion = 1

// Owners of a derived file.
const (
	// OwnerSmith means smith rewrites the file when its inputs change.
	OwnerSmith = "smith"
	// OwnerYou means a person owns the file and smith never touches it.
	OwnerYou = "you"
)

// Entry is the record of one derived file.
type Entry struct {
	// Path is workspace-relative, with forward slashes.
	Path string `toml:"path"`
	// Owner is OwnerSmith or OwnerYou.
	Owner string `toml:"owner"`
	// Renderer is the id of the renderer that wrote the file.
	Renderer string `toml:"renderer"`
	// Version is the version of that renderer.
	Version int `toml:"version"`
	// Mode is the ownership mode: file, region or keys.
	Mode string `toml:"mode"`
	// Hash is the hash of the content smith owns, as it last wrote it.
	Hash string `toml:"hash"`
	// Wanted is the hash of what the renderer wanted when the entry was
	// written. It differs from Hash for an adopted hand edit, and it tells a
	// change of the renderer's inputs apart from an edit of the file.
	Wanted string `toml:"wanted"`
}

// Record is smith.lock.
type Record struct {
	Schema int     `toml:"schema"`
	Files  []Entry `toml:"file"`
}

// Find returns the entry of a path and renderer, or nil. A path can have
// several entries: two renderers may own different keys or regions of one file.
func (r *Record) Find(path, renderer string) *Entry {
	for i := range r.Files {
		if r.Files[i].Path == path && r.Files[i].Renderer == renderer {
			return &r.Files[i]
		}
	}
	return nil
}

// Set adds the entry, or replaces the entry of the same path and renderer.
func (r *Record) Set(e Entry) {
	if old := r.Find(e.Path, e.Renderer); old != nil {
		*old = e
		return
	}
	r.Files = append(r.Files, e)
}

// EntriesFor lists the entries of a path.
func (r *Record) EntriesFor(path string) []*Entry {
	var out []*Entry
	for i := range r.Files {
		if r.Files[i].Path == path {
			out = append(out, &r.Files[i])
		}
	}
	return out
}

// OwnedBySmith lists the paths smith owns, sorted, each once.
func (r *Record) OwnedBySmith() []string {
	var out []string
	for _, e := range r.Files {
		if e.Owner == OwnerSmith {
			out = append(out, e.Path)
		}
	}
	slices.Sort(out)
	return slices.Compact(out)
}
