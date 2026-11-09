package releaseview

// GitState is what the repository says about pushing a release: the branch, the
// remote that receives the tags, and how far the branch is ahead of its upstream.
type GitState struct {
	// Known is false when the repository was not read.
	Known bool
	// Branch is the current branch ("" in a detached checkout).
	Branch string
	// Remote is the remote that receives the release tags; HasRemote says whether it exists.
	Remote    string
	HasRemote bool
	// Upstream is true when the branch follows a remote branch; Ahead is how many
	// commits it holds that the remote branch lacks.
	Upstream bool
	Ahead    int
}
