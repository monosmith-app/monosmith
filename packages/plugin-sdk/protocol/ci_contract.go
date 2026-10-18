package protocol

// CIParams gives the plugin the projects it owns so it can describe what CI
// needs for them.
type CIParams struct {
	Projects []Project `json:"projects"`
}

// CIResult is the plugin's contribution to every generated pipeline.
type CIResult struct {
	// Toolchains are rendered as provider-native setup steps.
	Toolchains []Toolchain `json:"toolchains,omitempty"`
	// Setup steps run in `smith ci setup`, in order, from the workspace root.
	Setup []Step `json:"setup,omitempty"`
	// Audit steps run in `smith ci audit`.
	Audit []Step `json:"audit,omitempty"`
	// PostRelease steps run at the end of `smith ci release`, whether or not
	// anything was released, with the release's credentials: work that must
	// follow every push to the release branch (setting a registry
	// description from a file). They run only after a release that did not
	// fail, from the workspace root, and a step that is not Blocking only
	// warns.
	PostRelease []Step `json:"postRelease,omitempty"`
	// Credentials names the environment variables `smith ci release` needs.
	Credentials []Credential `json:"credentials,omitempty"`
	// Permissions are the GitHub Actions token permissions the release needs
	// beyond the defaults ("packages: write" to push to ghcr.io). Other
	// providers have no such permissions and ignore them.
	Permissions []string `json:"permissions,omitempty"`
	// Authenticate are provider-native steps that sign in to a private
	// registry before setup (Azure Pipelines' npmAuthenticate@0, say). A
	// provider the plugin gives no step for gets none: its credentials come
	// from the environment instead.
	Authenticate []Toolchain `json:"authenticate,omitempty"`
}

// Toolchain is a runtime the pipeline must install, and how. Plugins asking
// for the same name share one setup step (the first plugin's).
type Toolchain struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	// GitHub is the GitHub Actions step that installs it (an action).
	GitHub *ActionStep `json:"github,omitempty"`
	// Azure is the Azure Pipelines step that installs it (a task).
	Azure *ActionStep `json:"azure,omitempty"`
	// Install is a shell command that installs it, run on a provider the
	// plugin gives no native step for.
	Install string `json:"install,omitempty"`
	// Devcontainer is the dev container feature that installs it at the
	// same version, so the editor's container matches CI.
	Devcontainer *DevcontainerFeature `json:"devcontainer,omitempty"`
}

// DevcontainerFeature is a dev container feature and its options
// (https://containers.dev/features).
type DevcontainerFeature struct {
	// ID is the feature reference, e.g. "ghcr.io/devcontainers/features/go:1".
	ID      string            `json:"id"`
	Options map[string]string `json:"options,omitempty"`
}

// ActionStep is a provider-native step: a GitHub action or an Azure task,
// with its inputs in the order they are written.
type ActionStep struct {
	// Uses is the action ("actions/setup-go@v7") or task ("GoTool@0").
	Uses   string  `json:"uses"`
	Inputs []Input `json:"inputs,omitempty"`
}

// Input is one input of an ActionStep. Values are written as quoted YAML
// strings, except true and false.
type Input struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Step is one command a CI phase runs.
type Step struct {
	Name    string   `json:"name"`
	Command []string `json:"command"`
	// Dir is the workspace-relative directory to run in; empty means the
	// workspace root.
	Dir string `json:"dir,omitempty"`
	// Blocking false makes a failure a warning instead of an error.
	Blocking bool `json:"blocking"`
}

// Credential is a secret the release phase needs.
type Credential struct {
	Env         string `json:"env"`
	Description string `json:"description"`
	// Required false means release works without it (e.g. a dry run target).
	Required bool `json:"required"`
	// Setup gives it to `smith ci setup` (and the native job's setup) as
	// well: installs from a private registry need it, not only the release.
	Setup bool `json:"setup,omitempty"`
	// Tasks gives it to every step that runs tasks (verify, pack, native)
	// as well: the remote cache's key and credentials are read there, not
	// only by the release.
	Tasks bool `json:"tasks,omitempty"`
}
