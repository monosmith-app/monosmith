package protocol

// GenerateParams asks a plugin to scaffold a project. The host chooses the
// directory from the kind (apps/, libs/, packages/).
type GenerateParams struct {
	Kind ProjectKind `json:"kind"`
	Name string      `json:"name"`
	// Directory is workspace-relative, forward slashes; it does not exist yet.
	Directory string `json:"directory"`
	// Workspace carries what a template needs to know about the monorepo.
	Workspace WorkspaceInfo `json:"workspace"`
	// Stack is the stack the project is created with. The plugin that writes
	// the manifest records it there, so discovery reports it back.
	Stack string `json:"stack,omitempty"`
	// Convention is true when the plugin is asked as a workspace convention
	// ([conventions.<name>] applying to the new project), not as one of the
	// stack's plugins. A plugin adds only what every matching project needs
	// then: a suite or starter files belong to the stack that names it.
	Convention bool `json:"convention,omitempty"`
	// Settings are this plugin's own settings for the new project, from the
	// recipe the project is created from (its settings.<plugin> table); the
	// same table is later the plugin's entry in Project.Settings. Nil when
	// there are none.
	Settings map[string]any `json:"settings,omitempty"`
	// Facts are the shared facts that hold for the new project, from
	// smith.toml: [projects.<name>.facts] over [facts]. A key that is absent
	// was set by no one: the plugin scaffolds its own default. Set by the
	// host; read only.
	Facts map[string]any `json:"facts,omitempty"`
	// Shape is the dialect the new project should have: the facts (see
	// Project.Shape) that most projects of its kind and language agree on, as
	// the host counted them. A plugin applies the keys it follows (see
	// InitializeResult.Follows) and ignores the rest. Nil when there is
	// nothing to follow, or the person asked for a fresh project.
	Shape map[string]string `json:"shape,omitempty"`
}

// WorkspaceInfo is the subset of workspace settings plugins may use.
type WorkspaceInfo struct {
	Name string `json:"name"`
	// Scope is the npm scope for packages, with its @ (e.g. "@acme").
	Scope string `json:"scope,omitempty"`
	// ModulePrefix is the Go module path prefix, e.g. github.com/acme/repo.
	ModulePrefix string `json:"modulePrefix,omitempty"`
}

// GenerateResult lists what the plugin wrote, workspace-relative.
type GenerateResult struct {
	Files []string `json:"files"`
	// Notes are printed to the user after generation.
	Notes []string `json:"notes,omitempty"`
	// InstallPending is true when the plugin changed what the workspace
	// depends on and left installing it to the host. A project is made by
	// several plugins in turn; the host asks each plugin that reported this
	// to install once (dependencies/install with Quiet set) after the last of
	// them has generated, instead of once per plugin.
	InstallPending bool `json:"installPending,omitempty"`
}
