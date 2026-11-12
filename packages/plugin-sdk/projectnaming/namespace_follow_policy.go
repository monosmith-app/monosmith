package projectnaming

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ecosystems maps the language a plugin reports (protocol.Project.Language) to
// the naming syntax of its packages.
var ecosystems = map[string]Ecosystem{
	"node": Npm, "rust": Cargo, "python": Python, "jvm": Maven, "dart": Dart,
	"cpp": CMake, "ruby": Gem, "php": Composer, "dotnet": DotNet,
}

// EcosystemOf is the naming syntax of a language's packages.
//
// Parameters: language is protocol.Project.Language.
// Returns: the ecosystem, and false for a language whose projects carry no
// namespace in their name (Go, Swift, Lua, and the rest).
func EcosystemOf(language string) (Ecosystem, bool) {
	eco, ok := ecosystems[language]
	return eco, ok
}

// Follows reports whether the name of an existing package starts with the
// workspace namespace the way Name would put it there: @acme/ for npm, acme-
// for Cargo, Python and Maven, acme_ for Dart, CMake and Gem, acme/ for
// Composer and Acme for .NET. A workspace with no namespace follows nothing,
// so every name does.
//
// Parameters: ws is the workspace; eco is the naming syntax; pkg is the
// package name its manifest gives.
// Returns: true when pkg starts with the namespace.
func Follows(ws protocol.WorkspaceInfo, eco Ecosystem, pkg string) bool {
	ns := Namespace(ws)
	if Kebab(ns) == "" {
		return true
	}
	switch eco {
	case Npm:
		return strings.HasPrefix(strings.ToLower(pkg), strings.ToLower(Scope(ws))+"/")
	case Cargo, Python, Maven:
		return strings.HasPrefix(Kebab(pkg), Kebab(ns)+"-")
	case Dart, CMake, Gem:
		return strings.HasPrefix(Snake(pkg), Snake(ns)+"_")
	case Composer:
		return strings.HasPrefix(strings.ToLower(pkg), Kebab(ns)+"/")
	case DotNet:
		return strings.HasPrefix(strings.ToLower(pkg), strings.ToLower(Pascal(ns)))
	default:
		return true
	}
}
