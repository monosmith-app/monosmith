package protocol

import "path"

// InitializeParams is sent once, first, by the host.
type InitializeParams struct {
	// ProtocolVersions lists every major version the host can speak.
	ProtocolVersions []int `json:"protocolVersions"`
	Host             Peer  `json:"host"`
	// WorkspaceRoot is an absolute path using the host OS separators.
	WorkspaceRoot string `json:"workspaceRoot"`
}

// InitializeResult tells the host who the plugin is and what it can do.
type InitializeResult struct {
	// ProtocolVersion is the version the plugin chose from the host's list.
	ProtocolVersion int          `json:"protocolVersion"`
	Plugin          Peer         `json:"plugin"`
	Kind            PluginKind   `json:"kind"`
	Capabilities    Capabilities `json:"capabilities"`
	// Languages are the values a user passes to --lang to reach this plugin.
	Languages []string `json:"languages,omitempty"`
	// Manifests are file base names (package.json, go.mod), or patterns of
	// them (*.csproj), whose presence in a directory may make it a project of
	// this plugin.
	Manifests []string `json:"manifests,omitempty"`
	// Roles are the levels the plugin can be chosen for in a project's
	// make-up (a stack, a recipe). Empty means the default of its kind (see
	// DefaultRoles). A role is a level: there is no separate level field,
	// because one plugin can play several (a scaffold that is also a
	// toolchain).
	Roles []Role `json:"roles,omitempty"`
	// Provides are the slots a toolchain plugin fills: the task names it
	// owns ("test" for a test runner). A slot is exclusive: two plugins of
	// one project may not provide the same task.
	Provides []string `json:"provides,omitempty"`
	// Requires are plugins that must also be in the project for this one to
	// work: a Node toolchain requires the node language plugin.
	Requires []PluginRef `json:"requires,omitempty"`
	// Conflicts are plugins that must not be in the same project.
	Conflicts []PluginRef `json:"conflicts,omitempty"`
	// Stacks are ready-made stacks the plugin publishes.
	Stacks []Stack `json:"stacks,omitempty"`
}

// EffectiveRoles are the plugin's roles, or the default of its kind.
func (r InitializeResult) EffectiveRoles() []Role {
	if len(r.Roles) > 0 {
		return r.Roles
	}
	return DefaultRoles(r.Kind)
}

// Peer names one side of the connection.
type Peer struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// PluginKind separates plugins that understand a language from plugins that
// add to projects a language plugin discovered (toolchains, structure rules,
// scaffolds, deploy targets), which are all of kind convention.
type PluginKind string

// Plugin kinds.
const (
	KindLanguage   PluginKind = "language"
	KindConvention PluginKind = "convention"
)

// Capabilities are the optional methods a plugin implements.
type Capabilities struct {
	Discover bool `json:"discover,omitempty"`
	Generate bool `json:"generate,omitempty"`
	Version  bool `json:"version,omitempty"`
	Publish  bool `json:"publish,omitempty"`
	CI       bool `json:"ci,omitempty"`
	Doctor   bool `json:"doctor,omitempty"`
	// Contribute is a convention plugin's: it adds tasks to projects other
	// plugins discovered.
	Contribute bool `json:"contribute,omitempty"`
	// Remove cleans up a removed project's traces outside its directory.
	Remove bool `json:"remove,omitempty"`
	// Published reports what a registry has of a project.
	Published bool `json:"published,omitempty"`
	// Dependencies installs and uninstalls packages in its projects.
	Dependencies bool `json:"dependencies,omitempty"`
	// Declarations lists and rewrites declared dependency specs (sync, up).
	Declarations bool `json:"declarations,omitempty"`
	// Latest reports the newest releases of dependencies and pinned tools.
	Latest bool `json:"latest,omitempty"`
	// Components lists a project's dependencies for its SBOM.
	Components bool `json:"components,omitempty"`
}

// IsManifest reports whether a file's base name is one of a plugin's
// manifests: equal to a name, or matching a pattern (*.csproj).
func IsManifest(manifests []string, base string) bool {
	for _, m := range manifests {
		if m == base {
			return true
		}
		if ok, _ := path.Match(m, base); ok {
			return true
		}
	}
	return false
}
