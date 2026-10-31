package why

// Kind says what sort of file a path is.
type Kind string

// Kinds of file.
const (
	// KindDerived: a plugin renderer wrote the file or part of it, and smith.lock records it.
	KindDerived Kind = "derived"
	// KindEjected: the file was derived once and a person owns it now.
	KindEjected Kind = "ejected"
	// KindGenerated: a smith command wrote the file and its header or marker says so.
	KindGenerated Kind = "generated"
	// KindWorkspace: a file smith reads and a person writes (smith.toml).
	KindWorkspace Kind = "workspace"
	// KindProject: a file inside a project that no record claims.
	KindProject Kind = "project"
	// KindUnknown: a file that no project contains and no record claims.
	KindUnknown Kind = "unknown"
)

// Owners of a file.
const (
	// OwnerSmith: smith rewrites the file when its inputs change.
	OwnerSmith = "smith"
	// OwnerYou: a person owns the file.
	OwnerYou = "you"
)

// Part is one record entry of the file: a renderer that owns all of it or a
// part of it.
type Part struct {
	// Renderer is the stable id of the renderer, for example node.tsconfig.references.
	Renderer string `json:"renderer"`
	// Version is the renderer version in the record.
	Version int `json:"version"`
	// Mode is file, region or keys.
	Mode string `json:"mode"`
	// Owner is smith or you.
	Owner string `json:"owner"`
	// State is the derive status (current, stale, drift...), empty when derive is off.
	State string `json:"state,omitempty"`
	// Detail is the derive sentence that explains the state.
	Detail string `json:"detail,omitempty"`
}

// Explanation answers `smith why` for one file.
type Explanation struct {
	// Path is workspace-relative, with forward slashes.
	Path string `json:"path"`
	// Known tells whether smith has anything to say about the file.
	Known bool `json:"known"`
	// Kind is the sort of file.
	Kind Kind `json:"kind"`
	// Owner is smith or you; empty for an unknown file.
	Owner string `json:"owner,omitempty"`
	// Summary is one sentence: who owns the file and why.
	Summary string `json:"summary"`
	// Project is the project that contains the file.
	Project string `json:"project,omitempty"`
	// Source is the intent, fact or choice that produced the file.
	Source string `json:"source,omitempty"`
	// Rule is the renderer or smith command that wrote it.
	Rule string `json:"rule,omitempty"`
	// Parts are the record entries of a derived file.
	Parts []Part `json:"parts,omitempty"`
	// Stack and Choices are those of the project.
	Stack   string   `json:"stack,omitempty"`
	Choices []string `json:"choices,omitempty"`
	// Change says how to change the file, one step per line.
	Change []string `json:"change"`
}
