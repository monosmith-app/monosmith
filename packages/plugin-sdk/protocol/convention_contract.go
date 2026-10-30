package protocol

// ContributeParams hands a plugin the projects it applies to: those whose
// stack includes it, or, for a workspace-wide convention, those of the
// languages it supports that smith.toml enables it for.
type ContributeParams struct {
	Projects []Project `json:"projects"`
}

// ContributeResult is what a convention adds, by project name.
type ContributeResult struct {
	Projects map[string]Contribution `json:"projects,omitempty"`
}

// Contribution is a convention's addition to one project.
type Contribution struct {
	// Tasks are new tasks. A task the project already has is kept as it is.
	Tasks map[string]Task `json:"tasks,omitempty"`
	// Replace are tasks that replace the project's task of the same name,
	// or are added when it has none.
	Replace map[string]Task `json:"replace,omitempty"`
	// DependsOn adds dependencies to tasks the project already has, by task
	// name ({"lint": ["slice-check"]}).
	DependsOn map[string][]string `json:"dependsOn,omitempty"`
	// Release is how the project's release files are made, for a project that
	// releases through GitHub releases and leaves them out of its smith.toml
	// [projects.<name>.release] table; what the table sets wins.
	Release *ReleaseDefaults `json:"release,omitempty"`
}

// ReleaseDefaults are release settings a plugin supplies for a project: the
// same fields as the release-from-a-command settings of smith.toml, used for
// the ones the project's table leaves out.
type ReleaseDefaults struct {
	// Build is the shell command that makes the files, run at the workspace
	// root on a runner of each platform.
	Build string `json:"build"`
	// Assets are workspace-relative globs of the files Build leaves.
	Assets []string `json:"assets"`
	// Platforms are the runner platforms (windows-x64, macos-arm64, macos-x64,
	// linux-x64, linux-arm64).
	Platforms []string `json:"platforms"`
	// AssetName renames each file on upload; see the asset_name setting.
	AssetName string `json:"assetName,omitempty"`
	// Checksums attaches a checksums file for the files a runner uploads.
	Checksums bool `json:"checksums,omitempty"`
}
