package workspace

// SiblingRepo is another repository the workspace's build needs beside it
// ([[ci.sibling_repos]]). The generated GitHub Actions workflow checks it out
// into the workspace, exports the path in an environment variable and runs the
// pre-build command before anything else in the job.
type SiblingRepo struct {
	// Repo is the GitHub repository, owner/name.
	Repo string `toml:"repo"`
	// Ref is the branch, tag or commit to check out. Default: the repository's
	// default branch.
	Ref string `toml:"ref"`
	// Env names the environment variable that holds the checkout's absolute
	// path ($GITHUB_WORKSPACE/<dir>) in every later step.
	Env string `toml:"env"`
	// Dir is the checkout's directory, relative to the workspace root.
	// Default: the repository's name.
	Dir string `toml:"dir"`
	// PreBuild is a shell command run in the checkout, after the toolchains are
	// installed and before smith sets the workspace up (cargo build --release
	// -p some-plugin). Default: none.
	PreBuild string `toml:"pre_build"`
	// TokenSecret names the GitHub secret holding a token that can read the
	// repository, when the job's own token cannot (a private repository).
	// Default: the job's token.
	TokenSecret string `toml:"token_secret"`
}
