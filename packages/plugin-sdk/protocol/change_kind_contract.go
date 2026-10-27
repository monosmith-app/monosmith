package protocol

// ChangeKindParams asks a plugin what changing a project's kind edits. The
// plugin writes nothing: it answers with the files' new contents, and the
// host applies them, then moves the project's directory.
type ChangeKindParams struct {
	// Project is the project as it is now.
	Project Project `json:"project"`
	// To is the kind it becomes.
	To ProjectKind `json:"to"`
	// NewRoot is the workspace-relative directory the host moves the project
	// to afterwards; it equals Project.Root when the directory stays.
	NewRoot string `json:"newRoot"`
	// Projects are every project of the workspace, so the plugin can find
	// the ones whose manifests and sources name this one.
	Projects []Project `json:"projects,omitempty"`
}

// ChangeKindResult is what the change edits. Every path in Files is the path
// the file has before the move.
type ChangeKindResult struct {
	// Refused says why this kind makes no sense for the project's language
	// ("a Bicep template is deployed, never published"); the host stops and
	// shows it. Empty when the change is fine.
	Refused string `json:"refused,omitempty"`
	// Files are the edits: manifests, workspace files and dependents that
	// name the project's kind or its directory.
	Files []FileEdit `json:"files,omitempty"`
	// Warnings name what the change loses or leaves for the user (a binary
	// an app has and a lib does not, a lockfile to refresh).
	Warnings []string `json:"warnings,omitempty"`
}
