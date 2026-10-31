package protocol

import (
	"path"
	"slices"
)

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
	// Concurrency is how many requests the plugin can work on at the same
	// time; the host never has more in flight. Absent or 0 means 1, which is
	// what every plugin that says nothing does: the host sends the next
	// request only after the previous one is answered. N above 1 allows that
	// many, and ConcurrencyUnlimited (-1) leaves the number to the host. Any
	// other negative value is invalid and read as 1.
	Concurrency int `json:"concurrency,omitempty"`
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
	// Kinds are the project kinds the plugin can scaffold with generate.
	// Absent or empty means all three (app, lib and package): what every
	// plugin that says nothing does. A plugin that scaffolds only some, such
	// as an app, lists them, so the host offers and the conformance suite
	// asks for just those.
	Kinds []ProjectKind `json:"kinds,omitempty"`
	// Follows are the shape keys (see Project.Shape) the plugin applies when
	// it generates a project with GenerateParams.Shape. Where siblings split
	// on a key, the host asks the person only for a key that a plugin of the
	// stack follows. Absent means the plugin follows none.
	Follows []string `json:"follows,omitempty"`
	// Choices are the options on an axis (framework, test, style) the plugin
	// owns. The host lists them (smith choices) and checks a selection of them
	// (smith add --with). Absent means the plugin owns none.
	Choices []Choice `json:"choices,omitempty"`
}

// Scaffolds says whether the plugin can scaffold a project of the kind.
func (r InitializeResult) Scaffolds(kind ProjectKind) bool {
	return len(r.Kinds) == 0 || slices.Contains(r.Kinds, kind)
}

// ScaffoldKinds are the project kinds the plugin can scaffold, in the order
// app, lib, package.
func (r InitializeResult) ScaffoldKinds() []ProjectKind {
	var out []ProjectKind
	for _, kind := range []ProjectKind{ProjectApp, ProjectLib, ProjectPackage} {
		if r.Scaffolds(kind) {
			out = append(out, kind)
		}
	}
	return out
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
	// ChangeKind answers what changing a project's kind edits (smith kind).
	ChangeKind bool `json:"changeKind,omitempty"`
	// Published reports what a registry has of a project.
	Published bool `json:"published,omitempty"`
	// Dependencies installs and uninstalls packages in its projects.
	Dependencies bool `json:"dependencies,omitempty"`
	// Declarations lists and rewrites declared dependency specs (sync, up).
	Declarations bool `json:"declarations,omitempty"`
	// Latest reports the newest releases of dependencies and pinned tools.
	Latest bool `json:"latest,omitempty"`
	// Cancel says the plugin stops the work of a request when the host sends
	// a cancel notification with its id, then answers that request with its
	// result (if the work was already done) or the request-cancelled error.
	// A host never sends the notification to a plugin without it.
	Cancel bool `json:"cancel,omitempty"`
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
