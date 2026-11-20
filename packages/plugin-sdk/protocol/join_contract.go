package protocol

// JoinParams asks a language plugin whether each of its projects is joined to
// the language's workspace (npm "workspaces", the members of a uv workspace, a
// .NET solution), and how to join the ones that are not. The plugin writes
// nothing: the host shows the fix and applies it (smith doctor --fix, smith
// import).
type JoinParams struct {
	Projects []Project `json:"projects"`
}

// JoinResult has one state for each project the plugin was asked about. A
// project the plugin leaves out is joined.
type JoinResult struct {
	Projects []JoinState `json:"projects"`
}

// JoinState says whether one project is joined to its language's workspace.
type JoinState struct {
	// Project is the project's name, as JoinParams gave it.
	Project string `json:"project"`
	// Joined is true when the workspace covers the project, or when the
	// language has no workspace the project must be part of.
	Joined bool `json:"joined"`
	// Workspace names what the project joins, in words, such as
	// `the root package.json "workspaces"`.
	Workspace string `json:"workspace,omitempty"`
	// Problem says, in one line, what is missing and what it breaks, when
	// Joined is false.
	Problem string `json:"problem,omitempty"`
	// Fix joins the project, when Joined is false and the plugin can write
	// the change: the manifest's new content, then the commands to run.
	Fix *Fix `json:"fix,omitempty"`
}
