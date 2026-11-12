package projectnaming

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Namespace is the word every language puts before a project's name: the
// workspace's namespace, else (an older host, or a namespace left out) its
// name.
//
// Parameters: ws is the workspace the project is made in.
// Returns: the namespace, empty when the workspace has neither.
func Namespace(ws protocol.WorkspaceInfo) string {
	if ns := strings.TrimSpace(ws.Namespace); ns != "" {
		return ns
	}
	return strings.TrimSpace(ws.Name)
}

// Scope is the npm scope with its @: the workspace's explicit scope, else the
// namespace in kebab-case.
//
// Parameters: ws is the workspace the project is made in.
// Returns: the scope (@acme), empty when there is no namespace.
func Scope(ws protocol.WorkspaceInfo) string {
	if scope := strings.TrimPrefix(strings.TrimSpace(ws.Scope), "@"); scope != "" {
		return "@" + scope
	}
	if ns := Kebab(Namespace(ws)); ns != "" {
		return "@" + ns
	}
	return ""
}

// GoModulePrefix is the prefix of a Go module path: the workspace's explicit
// module prefix, else the placeholder example.com/<namespace> (a repository
// with no git remote).
//
// Parameters: ws is the workspace the module is made in.
// Returns: the prefix, never empty.
func GoModulePrefix(ws protocol.WorkspaceInfo) string {
	if ws.ModulePrefix != "" {
		return ws.ModulePrefix
	}
	element := pathElement(Namespace(ws))
	if element == "" {
		element = "workspace"
	}
	return "example.com/" + element
}

// GroupID is a new workspace's Maven groupId, dev.<namespace>: a placeholder
// for the owner's reverse domain.
//
// Parameters: ws is the workspace the module is made in.
// Returns: the groupId.
func GroupID(ws protocol.WorkspaceInfo) string { return "dev." + Flat(Namespace(ws)) }

// Name is the default name of a new project in an ecosystem: the namespace,
// then the project name, shaped to the ecosystem's syntax. A project name
// that already starts with the namespace is not prefixed twice (a project
// acme-core in namespace acme stays acme-core). The person's --package-name
// wins over this and is applied by the caller.
//
// Parameters: ws is the workspace; eco is the naming syntax; project is the
// project name as the person gave it.
// Returns: the name; the project name alone, shaped, when there is no namespace.
func Name(ws protocol.WorkspaceInfo, eco Ecosystem, project string) string {
	ns := Namespace(ws)
	switch eco {
	case Npm:
		if scope := Scope(ws); scope != "" {
			return scope + "/" + project
		}
		return project
	case Cargo:
		return joined(Kebab(ns), Kebab(project), "-", true)
	case Python, Maven:
		return joined(Kebab(ns), Kebab(project), "-", false)
	case Dart, CMake, Gem:
		return joined(Snake(ns), Snake(project), "_", false)
	case Composer:
		vendor := Kebab(ns)
		if vendor == "" {
			vendor = Kebab(project)
		}
		return vendor + "/" + Kebab(project)
	case DotNet:
		if prefix, name := Kebab(ns), Kebab(project); prefix == "" || strings.HasPrefix(name, prefix+"-") {
			return Pascal(project)
		}
		return Pascal(ns) + "." + Pascal(project)
	case Swift, ObjC:
		return Pascal(project)
	case VBA:
		return vbaIdentifier(project)
	default: // Lua, and an unknown ecosystem: the project name as given.
		return project
	}
}

// joined puts prefix before name, unless prefix is empty or name starts with it.
func joined(prefix, name, sep string, letterFirst bool) string {
	if letterFirst && startsWithDigit(prefix) {
		prefix = "x" + prefix
	}
	if prefix == "" || strings.HasPrefix(name, prefix+sep) {
		return name
	}
	return prefix + sep + name
}
