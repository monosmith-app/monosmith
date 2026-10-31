package protocol

// Ownership modes of a derived file. They say how much of the file the
// renderer owns, so how much of it smith may rewrite.
const (
	// DeriveFile means the renderer owns the whole file.
	DeriveFile = "file"
	// DeriveRegion means the renderer owns the lines between a
	// "smith:begin <renderer>" and a "smith:end <renderer>" marker; the rest of
	// the file is the person's.
	DeriveRegion = "region"
	// DeriveKeys means the renderer owns the listed top-level keys of a JSON or
	// TOML file; the other keys are the person's.
	DeriveKeys = "keys"
)

// DeriveParams asks a plugin for the files it can derive from facts and the
// project graph. The plugin writes nothing: it answers with the content, and
// the host decides, from its ownership record, whether to write it.
type DeriveParams struct {
	// Projects are the projects the plugin serves.
	Projects []Project `json:"projects"`
	// Graph maps a project name to the names of the projects it depends on
	// inside the workspace, sorted. The host resolved these edges: a plugin
	// does not repeat that work.
	Graph map[string][]string `json:"graph,omitempty"`
	// Owned are the workspace-relative paths the host's ownership record says
	// smith owns. A plugin proposes such a file again even when it is missing
	// on disk, so deleting a derived file and deriving restores it. A file that
	// is not owned and not on disk is not proposed.
	Owned []string `json:"owned,omitempty"`
}

// DeriveResult carries the files a plugin derives.
type DeriveResult struct {
	Files []DerivedFile `json:"files"`
}

// DerivedFile is the content one renderer derives for one file.
type DerivedFile struct {
	// Project is the project the file belongs to.
	Project string `json:"project"`
	// Path is workspace-relative, with forward slashes.
	Path string `json:"path"`
	// Mode is DeriveFile, DeriveRegion or DeriveKeys.
	Mode string `json:"mode"`
	// Renderer is the stable id of the renderer, for example
	// "node.tsconfig.references". It names the region markers.
	Renderer string `json:"renderer"`
	// Version is the renderer's version. A new version means new output for
	// the same inputs, so the host rewrites an unedited file.
	Version int `json:"version"`
	// Content is the whole file (DeriveFile) or the lines of the region
	// (DeriveRegion), without the markers. Unused for DeriveKeys.
	Content string `json:"content,omitempty"`
	// Comment is the line comment leader of the file ("//", "#"): the host
	// writes a generated header after it in a DeriveFile. Empty for a format
	// without comments, then the file has no header.
	Comment string `json:"comment,omitempty"`
	// Format is "json" or "toml" for DeriveKeys.
	Format string `json:"format,omitempty"`
	// Keys are the owned top-level keys and their values (DeriveKeys).
	Keys map[string]any `json:"keys,omitempty"`
}
