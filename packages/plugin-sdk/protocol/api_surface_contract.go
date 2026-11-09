package protocol

// APISurfaceParams asks a plugin for the public API names of one project.
type APISurfaceParams struct {
	Project Project `json:"project"`
	// Ref is the git revision to read the project at (the tag of its last
	// release). Empty means the working tree.
	Ref string `json:"ref,omitempty"`
}

// APISurfaceResult is the public API of a project: the names its entry
// points export. A host compares the names of two revisions; it never reads
// their meaning, so a plugin chooses its own spelling and keeps it stable.
type APISurfaceResult struct {
	// Names are the exported names, sorted. A name an entry point other than
	// the main one exports is prefixed with that entry point and a colon
	// ("./cli:run"). An empty list is an answer: the project exports nothing.
	Names []string `json:"names"`
	// Unreadable says why the API could not be listed (no entry point, no
	// source at the revision). When set, Names is empty and the host skips
	// the project without a warning.
	Unreadable string `json:"unreadable,omitempty"`
}
