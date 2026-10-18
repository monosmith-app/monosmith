package protocol

// Role is a part a plugin plays in a project's make-up: its level. Some
// levels are exclusive (one plugin per project, or per slot) and the others
// additive (any number, and they only add).
type Role string

// Plugin roles. A project has exactly one language and at most one scaffold;
// a toolchain is exclusive per slot (the tasks it provides); structure and
// deploy plugins are additive.
const (
	// RoleLanguage finds projects from manifests, turns imports into graph
	// edges, writes versions and publishes.
	RoleLanguage Role = "language"
	// RoleToolchain decides how a project builds, tests and lints.
	RoleToolchain Role = "toolchain"
	// RoleStructure checks layout and code rules.
	RoleStructure Role = "structure"
	// RoleScaffold writes a new project's starter files.
	RoleScaffold Role = "scaffold"
	// RoleDeploy adds pipeline steps and deploy targets.
	RoleDeploy Role = "deploy"
)

// Roles are every known role, in the order a project's make-up lists them.
var Roles = []Role{RoleLanguage, RoleToolchain, RoleScaffold, RoleStructure, RoleDeploy}

// Exclusive reports whether a project has at most one plugin at this level
// (language, scaffold). A toolchain is exclusive per slot instead.
func (r Role) Exclusive() bool { return r == RoleLanguage || r == RoleScaffold }

// Additive reports whether any number of plugins can share the level, each
// only adding (structure, deploy).
func (r Role) Additive() bool { return r == RoleStructure || r == RoleDeploy }

// Known reports whether the role is one of Roles.
func (r Role) Known() bool {
	for _, k := range Roles {
		if k == r {
			return true
		}
	}
	return false
}

// DefaultRoles are the roles of a plugin that declares none: a language
// plugin brings its own default toolchain and scaffold; a convention plugin
// is a structure rule.
func DefaultRoles(kind PluginKind) []Role {
	if kind == KindLanguage {
		return []Role{RoleLanguage, RoleToolchain, RoleScaffold}
	}
	return []Role{RoleStructure}
}
