package protocol

// DiscoverParams asks a plugin to describe the projects behind manifests the
// host found. Paths are workspace-relative and use forward slashes.
type DiscoverParams struct {
	Manifests []string `json:"manifests"`
}

// DiscoverResult describes every manifest that is a project of this plugin.
// A manifest the plugin does not own is simply left out.
type DiscoverResult struct {
	Projects []Project `json:"projects"`
}

// Project is one unit of the monorepo as a plugin sees it.
type Project struct {
	// Name is unique across the workspace; it is what users type.
	Name string `json:"name"`
	// Root is the workspace-relative directory, forward slashes.
	Root     string      `json:"root"`
	Language string      `json:"language"`
	Kind     ProjectKind `json:"kind"`
	// Exports are the identifiers other projects import this one by: an npm
	// package name, a Go module path, a Python import name.
	Exports []string `json:"exports,omitempty"`
	// Imports are identifiers this project's sources import. The host keeps
	// the ones another project exports and turns them into graph edges.
	Imports []string `json:"imports,omitempty"`
	// Version is the version recorded in the manifest, if it has one.
	Version string `json:"version,omitempty"`
	// Toolchain is the version of the compiler or runtime the project's tasks
	// run with ("node 20.11.0", "go 1.23.4"), as the plugin probed it. It goes
	// into every task's cache key, so a result from another version is a
	// miss. Empty when the plugin cannot tell, and then it is left out.
	Toolchain string `json:"toolchain,omitempty"`
	// Description is the one-line description the manifest records (package.json
	// description, pyproject.toml project.description, the pom's <description>),
	// if it has one. Conventions use it where a registry asks for a summary.
	Description string `json:"description,omitempty"`
	// TagPrefix overrides the release tag prefix; empty means "<name>@".
	// Go modules use "<root>/v" so `go get` can resolve the tag.
	TagPrefix string          `json:"tagPrefix,omitempty"`
	Tasks     map[string]Task `json:"tasks,omitempty"`
	Tags      []string        `json:"tags,omitempty"`
	// Stack is the stack recorded in the project's manifest, if any.
	Stack string `json:"stack,omitempty"`
	// Settings are plugins' settings for this project from smith.toml, by
	// plugin name: [projects.<project>.settings.<name>] (what a recipe's
	// settings became when it created the project), and for a convention its
	// [conventions.<name>] settings and the project's
	// [projects.<project>.<name>] on top. Set by the host; a plugin reads its
	// own entry.
	Settings map[string]map[string]any `json:"settings,omitempty"`
	// Registries are the private registries the project installs from and
	// publishes to, from smith.toml [registries]; nil means the public ones.
	// Set by the host.
	Registries *Registries `json:"registries,omitempty"`
}

// SetToolchain records the toolchain on every discovered project. The probe
// runs only when there is a project, and once.
func (r *DiscoverResult) SetToolchain(probe func() string) {
	if len(r.Projects) == 0 {
		return
	}
	toolchain := probe()
	for i := range r.Projects {
		r.Projects[i].Toolchain = toolchain
	}
}

// ProjectKind is the universal kind vocabulary.
type ProjectKind string

// Project kinds. An app is deployed, never published; a lib is internal and
// never published; a package is published.
const (
	ProjectApp     ProjectKind = "app"
	ProjectLib     ProjectKind = "lib"
	ProjectPackage ProjectKind = "package"
)

// Task is a command the host can run, hash and cache.
type Task struct {
	// Command is an argv array; it is never passed through a shell.
	Command []string `json:"command"`
	// Cwd is workspace-relative; empty means the project root.
	Cwd string            `json:"cwd,omitempty"`
	Env map[string]string `json:"env,omitempty"`
	// Inputs are globs relative to the project root. Empty means every
	// tracked or untracked-but-not-ignored file of the project.
	Inputs []string `json:"inputs,omitempty"`
	// Outputs are directories relative to the project root, restored on a
	// cache hit.
	Outputs []string `json:"outputs,omitempty"`
	// DependsOn lists task names: "build" runs this project's build first,
	// "^build" runs the build of every project this one imports first.
	DependsOn []string `json:"dependsOn,omitempty"`
	Cache     bool     `json:"cache,omitempty"`
	// Description is shown by `smith explain` and the TUI.
	Description string `json:"description,omitempty"`
	// Verify asks CI's verify phase (smith ci verify) to run the task too,
	// after lint, typecheck, test and build, for the affected projects that
	// have it: an app's container image, so a broken Dockerfile fails the
	// pull request, not the release. It is skipped, with a warning, where
	// its program is not installed.
	Verify bool `json:"verify,omitempty"`
}
